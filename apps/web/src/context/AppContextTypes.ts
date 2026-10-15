import type { Role, ThemeMode, DensityMode } from '../types/domain';

export type EnvironmentMode = 'SIMULATED_SANDBOX' | 'PRODUCTION';

export interface ToastItem {
  id: string;
  type: 'success' | 'warning' | 'error' | 'info';
  title: string;
  message?: string;
  action?: {
    label: string;
    onClick: () => void;
  };
}

export interface AppContextType {
  theme: ThemeMode;
  setTheme: (t: ThemeMode) => void;
  density: DensityMode;
  setDensity: (d: DensityMode) => void;
  role: Role;
  setRole: (r: Role) => void;
  environment: EnvironmentMode;
  setEnvironment: (env: EnvironmentMode) => void;
  activeView: string;
  setActiveView: (view: string) => void;
  globalSearchOpen: boolean;
  setGlobalSearchOpen: (open: boolean) => void;
  sessionTimeoutOpen: boolean;
  setSessionTimeoutOpen: (open: boolean) => void;
  toasts: ToastItem[];
  addToast: (toast: Omit<ToastItem, 'id'>) => void;
  dismissToast: (id: string) => void;
  selectedTransactionId: string | null;
  setSelectedTransactionId: (id: string | null) => void;
}
