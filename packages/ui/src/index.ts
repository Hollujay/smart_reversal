/**
 * SMARTREVERSAL Shared UI & Design Tokens
 * Standard tokens for enterprise financial operations console.
 */

export const colors = {
  navyDark: '#0b132b',
  navyCard: '#1c2541',
  slateAccent: '#3a506b',
  cyanHighlight: '#48cae4',
  tealAccent: '#5bc0be',
  successGreen: '#10b981',
  warningAmber: '#f59e0b',
  dangerRed: '#ef4444',
  textPrimary: '#f8fafc',
  textSecondary: '#94a3b8',
  borderDefault: '#334155',
} as const;

export type FinancialStatus =
  | 'SUCCESS'
  | 'PENDING'
  | 'FAILED'
  | 'UNCERTAIN'
  | 'REVERSED'
  | 'PENDING_MAKER'
  | 'PENDING_CHECKER';

export interface StatusBadgeProps {
  status: FinancialStatus;
  label?: string;
}

export const getStatusColor = (status: FinancialStatus): string => {
  switch (status) {
    case 'SUCCESS':
    case 'REVERSED':
      return colors.successGreen;
    case 'PENDING':
    case 'PENDING_CHECKER':
    case 'PENDING_MAKER':
      return colors.warningAmber;
    case 'FAILED':
    case 'UNCERTAIN':
      return colors.dangerRed;
    default:
      return colors.textSecondary;
  }
};
