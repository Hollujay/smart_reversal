// Command bootstrap creates the first institution and operator of an
// environment. The password is read from BOOTSTRAP_PASSWORD, never from an
// argument, so it does not appear in the process list or shell history.
//
//	BOOTSTRAP_PASSWORD='...' bootstrap -institution DEMO-BANK -email ops@example.test -roles ADMIN_INTEGRATION
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/smartreversal/smartreversal/apps/api/internal/config"
	"github.com/smartreversal/smartreversal/apps/api/internal/identity"
	"github.com/smartreversal/smartreversal/apps/api/internal/platform/db"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "bootstrap:", err)
		os.Exit(1)
	}
}

func run() error {
	code := flag.String("institution", "", "institution code, capitals, digits and dashes")
	instName := flag.String("institution-name", "", "institution display name (default: the code)")
	email := flag.String("email", "", "operator email")
	name := flag.String("name", "", "operator display name (default: the email)")
	roles := flag.String("roles", "", "comma-separated roles, for example ADMIN_INTEGRATION or REVERSAL_OFFICER,SUPERVISOR")
	flag.Parse()

	password := os.Getenv("BOOTSTRAP_PASSWORD")
	if password == "" {
		return fmt.Errorf("set BOOTSTRAP_PASSWORD")
	}
	var roleList []string
	for _, r := range strings.Split(*roles, ",") {
		if r = strings.TrimSpace(r); r != "" {
			roleList = append(roleList, r)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("configuration: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Open(ctx, cfg.DatabaseURL(), db.Options{MaxConns: 2})
	if err != nil {
		return err
	}
	defer pool.Close()

	tenant, user, err := identity.Bootstrap(ctx, pool, identity.BootstrapSpec{
		InstitutionCode: *code, InstitutionName: *instName, Email: *email, DisplayName: *name,
		Password: password, Roles: roleList,
	})
	if err != nil {
		return err
	}
	fmt.Printf("institution %s (%s)\noperator %s\n", *code, tenant, user)
	return nil
}
