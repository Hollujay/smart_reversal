package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Role string

const (
	RoleAnalystMaker      Role = "ANALYST_MAKER"
	RoleSupervisorChecker Role = "SUPERVISOR_CHECKER"
	RoleAuditor           Role = "AUDITOR"
	RoleAdmin             Role = "ADMIN"
)

type userContextKey struct{}

// Claims defines standard JWT claims with role and user identifiers.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   Role   `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	secretKey []byte
}

func NewAuthService(secret string) *AuthService {
	return &AuthService{
		secretKey: []byte(secret),
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthTokenResponse struct {
	Token     string `json:"token"`
	Role      Role   `json:"role"`
	UserID    string `json:"userId"`
	ExpiresIn int64  `json:"expiresIn"`
}

// GenerateToken issues a signed JWT for the operator.
func (s *AuthService) GenerateToken(userID, email string, role Role, duration time.Duration) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "smartreversal-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateToken verifies and extracts claims from a raw bearer token.
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// AuthMiddleware extracts Bearer tokens and attaches Claims to the context.
func (s *AuthService) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
			return
		}

		claims, err := s.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, `{"error":"unauthorized: invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey{}, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRoles verifies that the authenticated user possesses one of the allowed roles.
func RequireRoles(allowedRoles ...Role) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserClaims(r.Context())
			if claims == nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			for _, role := range allowedRoles {
				if claims.Role == role || claims.Role == RoleAdmin {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, `{"error":"forbidden: insufficient maker-checker authorization privileges"}`, http.StatusForbidden)
		})
	}
}

// GetUserClaims retrieves the active user claims from context.
func GetUserClaims(ctx context.Context) *Claims {
	if val, ok := ctx.Value(userContextKey{}).(*Claims); ok {
		return val
	}
	return nil
}

// HandleLogin authenticates pre-seeded demo or database users.
func (s *AuthService) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	var role Role
	var userID string
	switch req.Email {
	case "analyst@smartreversal.io":
		role = RoleAnalystMaker
		userID = "usr-analyst-001"
	case "supervisor@smartreversal.io", "checker@smartreversal.io":
		role = RoleSupervisorChecker
		userID = "usr-supervisor-002"
	case "auditor@smartreversal.io":
		role = RoleAuditor
		userID = "usr-auditor-003"
	default:
		// Default to analyst for development login
		role = RoleAnalystMaker
		userID = "usr-dev-001"
	}

	token, err := s.GenerateToken(userID, req.Email, role, 24*time.Hour)
	if err != nil {
		http.Error(w, `{"error":"failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	resp := AuthTokenResponse{
		Token:     token,
		Role:      role,
		UserID:    userID,
		ExpiresIn: 86400,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
