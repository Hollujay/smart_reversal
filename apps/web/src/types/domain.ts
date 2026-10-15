/**
 * Domain types matching packages/api-contracts/openapi.yaml
 */

export type ThemeMode = 'daylight' | 'soft-dim';
export type DensityMode = 'comfortable' | 'compact';

export type Role =
  | 'ANALYST_MAKER'
  | 'SUPERVISOR_CHECKER'
  | 'AUDITOR'
  | 'ADMIN';

export type Channel = 'NIP' | 'POS' | 'CARD' | 'WEBHOOK' | 'DIRECT_DEBIT';

export type FinancialStatus =
  | 'detected'
  | 'awaiting'
  | 'investigating'
  | 'approval'
  | 'submitted'
  | 'verifying'
  | 'resolved'
  | 'failed'
  | 'suppressed';

export type Severity = 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';

export type ExceptionType =
  | 'DEBIT_WITHOUT_CREDIT'
  | 'DUPLICATE_DEBIT'
  | 'STATUS_MISMATCH'
  | 'SUSPENSE_BALANCE'
  | 'TIMEOUT_EXCEPTION';

export interface Transaction {
  id: string;
  reference: string;
  rrn: string;
  stan: string;
  amount: string; // Stored as precise string representation of decimal
  currency: string;
  channel: Channel;
  status: FinancialStatus;
  senderAccount: string;
  beneficiaryAccount: string;
  senderBank: string;
  beneficiaryBank: string;
  createdAt: string;
  settledAt?: string;
  maskedPan?: string;
  hasDiscrepancy: boolean;
  exceptionType?: ExceptionType;
  modelConfidence?: number; // 0-100%
  isVerified: boolean;
}

export interface FinancialException {
  id: string;
  transactionId: string;
  reference: string;
  type: ExceptionType;
  severity: Severity;
  amount: string;
  currency: string;
  channel: Channel;
  status: FinancialStatus;
  description: string;
  detectedAt: string;
  modelConfidence: number; // e.g. 96
  modelExplanation: string;
  assignedTo?: string;
}

export interface EvidenceItem {
  id: string;
  type: 'FACT' | 'INFERRED' | 'MISSING' | 'UNRESOLVED';
  title: string;
  source: string;
  timestamp?: string;
  details: string;
  verified: boolean;
}

export interface InvestigationCase {
  id: string;
  caseNumber: string;
  exceptionId: string;
  reference: string;
  status: FinancialStatus;
  priority: Severity;
  assignedAnalyst: string;
  createdAt: string;
  updatedAt: string;
  evidence: EvidenceItem[];
  proposedAction?: string;
  aiSuggestedRemedy?: string;
  modelConfidence: number;
}

export interface ReversalProposal {
  id: string;
  transactionId: string;
  reference: string;
  originalAmount: string;
  reversalAmount: string;
  currency: string;
  channel: Channel;
  status: FinancialStatus; // 'approval', 'submitted', 'resolved', 'failed'
  proposedBy: string; // Maker
  proposedAt: string;
  reasonCode: string;
  justification: string;
  authorizedBy?: string; // Checker
  authorizedAt?: string;
  rejectionReason?: string;
  idempotencyKey: string;
  destinationAccount: string;
  destinationBank: string;
}

export interface AuditEvent {
  id: string;
  timestamp: string;
  actorId: string;
  actorRole: Role;
  action: string;
  entityType: 'TRANSACTION' | 'REVERSAL' | 'INVESTIGATION' | 'AUTH' | 'CONFIG';
  entityId: string;
  prevHash: string;
  entryHash: string;
  details: Record<string, unknown>;
  ipAddress: string;
}

export interface ReconciliationSummary {
  cycleDate: string;
  channel: Channel;
  totalTransactions: number;
  matchedCount: number;
  unmatchedCount: number;
  totalVolume: string;
  discrepancyVolume: string;
  reconciliationRate: number; // e.g. 98.4%
  status: 'COMPLETED' | 'IN_PROGRESS' | 'DISCREPANCY_FLAGGED';
}
