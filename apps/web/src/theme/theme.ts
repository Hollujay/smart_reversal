import { colors, type FinancialStatus, getStatusColor } from '@smartreversal/ui';

export const theme = {
  colors: {
    ...colors,
    brandPrimary: '#38bdf8',
    brandSecondary: '#6366f1',
    surfaceDark: '#080d1a',
    surfaceCard: 'rgba(15, 23, 42, 0.75)',
    surfaceCardHover: 'rgba(30, 41, 59, 0.85)',
    surfaceBorder: 'rgba(51, 65, 85, 0.6)',
  },
  typography: {
    fontSans: "'Inter', sans-serif",
    fontMono: "'JetBrains Mono', monospace",
  },
  shadows: {
    glow: '0 0 25px rgba(56, 189, 248, 0.15)',
    card: '0 10px 25px -5px rgba(0, 0, 0, 0.5)',
  },
};

export { colors, getStatusColor };
export type { FinancialStatus };
