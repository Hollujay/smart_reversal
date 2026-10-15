import React, { useState, useEffect } from 'react';
import type { Role, ThemeMode, DensityMode } from '../types/domain';
import { AppContext } from './AppContextInstance';
import type { EnvironmentMode, ToastItem } from './AppContextTypes';

export const AppProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [theme, setThemeState] = useState<ThemeMode>('daylight');
  const [density, setDensityState] = useState<DensityMode>('comfortable');
  const [role, setRole] = useState<Role>('ANALYST_MAKER');
  const [environment, setEnvironment] = useState<EnvironmentMode>('SIMULATED_SANDBOX');
  const [activeView, setActiveView] = useState<string>('dashboard');
  const [globalSearchOpen, setGlobalSearchOpen] = useState(false);
  const [sessionTimeoutOpen, setSessionTimeoutOpen] = useState(false);
  const [toasts, setToasts] = useState<ToastItem[]>([]);
  const [selectedTransactionId, setSelectedTransactionId] = useState<string | null>(null);

  // Sync theme with DOM root data attributes
  const setTheme = (newTheme: ThemeMode) => {
    setThemeState(newTheme);
    document.documentElement.setAttribute('data-theme', newTheme);
  };

  const setDensity = (newDensity: DensityMode) => {
    setDensityState(newDensity);
    document.documentElement.setAttribute('data-density', newDensity);
  };

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
    document.documentElement.setAttribute('data-density', density);
  }, [theme, density]);

  // Keyboard shortcut listener for Ctrl+K / Cmd+K
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        setGlobalSearchOpen((prev) => !prev);
      }
      if (e.key === 'Escape') {
        setGlobalSearchOpen(false);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  const dismissToast = (id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  };

  const addToast = (t: Omit<ToastItem, 'id'>) => {
    setToasts((prev) => {
      const id = `toast-${prev.length + 1}`;
      const newToast: ToastItem = { ...t, id };
      setTimeout(() => {
        dismissToast(id);
      }, 6000);
      return [...prev, newToast];
    });
  };

  return (
    <AppContext.Provider
      value={{
        theme,
        setTheme,
        density,
        setDensity,
        role,
        setRole,
        environment,
        setEnvironment,
        activeView,
        setActiveView,
        globalSearchOpen,
        setGlobalSearchOpen,
        sessionTimeoutOpen,
        setSessionTimeoutOpen,
        toasts,
        addToast,
        dismissToast,
        selectedTransactionId,
        setSelectedTransactionId,
      }}
    >
      {children}
    </AppContext.Provider>
  );
};


/* eslint-disable-next-line react-refresh/only-export-components */
export { useApp } from './useApp';
