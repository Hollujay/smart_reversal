import type {
  Transaction,
  FinancialException,
  InvestigationCase,
  ReversalProposal,
  AuditEvent,
  ReconciliationSummary,
  FinancialStatus,
  Severity,
  Channel,
  Role,
  ExceptionTeam,
  ExceptionNote,
  RootCause,
  ExternalDependency,
} from '../types/domain';
import {
  INITIAL_TRANSACTIONS,
  INITIAL_EXCEPTIONS,
  INITIAL_INVESTIGATIONS,
  INITIAL_REVERSALS,
  INITIAL_AUDIT_TRAIL,
  RECONCILIATION_RUNS,
} from './mockSimulatorData';

// In-memory reactive state for simulator
let transactions = [...INITIAL_TRANSACTIONS];
let exceptions = [...INITIAL_EXCEPTIONS];
let investigations = [...INITIAL_INVESTIGATIONS];
let reversals = [...INITIAL_REVERSALS];
let auditTrail = [...INITIAL_AUDIT_TRAIL];
const reconciliationRuns = [...RECONCILIATION_RUNS];

export interface QueryTransactionParams {
  reference?: string;
  channel?: Channel;
  status?: FinancialStatus;
  limit?: number;
}

export interface ProposeReversalParams {
  transactionId: string;
  reasonCode: string;
  justification: string;
  proposedBy: string;
}

export interface AuthorizeReversalParams {
  reversalId: string;
  authorizedBy: string;
  decision: 'APPROVE' | 'REJECT';
  notes?: string;
}

export const apiClient = {
  // GET /health
  async getHealth() {
    return {
      status: 'operational',
      version: '0.1.0',
      database: 'connected (PostgreSQL 16)',
      redis: 'connected (Redis 7.2)',
      nipSwitchAdapter: 'ONLINE (NIBSS Sandbox Gateway)',
      coreBankingLedger: 'SYNCHRONIZED',
      aiAnomalyEngine: 'ACTIVE (Latency 14ms)',
    };
  },

  // GET /transactions
  async getTransactions(params?: QueryTransactionParams): Promise<Transaction[]> {
    let result = [...transactions];
    if (params?.reference) {
      const q = params.reference.toLowerCase();
      result = result.filter(
        (t) =>
          t.reference.toLowerCase().includes(q) ||
          t.rrn.includes(q) ||
          t.stan.includes(q) ||
          t.senderAccount.includes(q) ||
          t.beneficiaryAccount.includes(q)
      );
    }
    if (params?.channel) {
      result = result.filter((t) => t.channel === params.channel);
    }
    if (params?.status) {
      result = result.filter((t) => t.status === params.status);
    }
    return result;
  },

  // GET /exceptions
  async getExceptions(params?: {
    severity?: Severity;
    channel?: Channel;
    status?: FinancialStatus;
    classificationCode?: string;
    source?: string;
  }): Promise<FinancialException[]> {
    let result = [...exceptions];
    if (params?.severity) {
      result = result.filter((e) => e.severity === params.severity);
    }
    if (params?.channel) {
      result = result.filter((e) => e.channel === params.channel);
    }
    if (params?.status) {
      result = result.filter((e) => e.status === params.status);
    }
    if (params?.classificationCode) {
      result = result.filter((e) => e.classificationCode === params.classificationCode);
    }
    if (params?.source) {
      result = result.filter((e) => e.source === params.source);
    }
    return result;
  },

  // CLAIM exception (with optimistic concurrency check)
  async claimException(
    exceptionId: string,
    analystId: string,
    analystName: string,
    expectedVersion: number
  ): Promise<FinancialException> {
    const exc = exceptions.find((e) => e.id === exceptionId);
    if (!exc) throw new Error('Exception not found');

    if (exc.version !== expectedVersion) {
      const err = new Error('CONCURRENCY_CONFLICT');
      (err as unknown as { currentVersion: number; updatedBy: string }).currentVersion = exc.version;
      (err as unknown as { currentVersion: number; updatedBy: string }).updatedBy = exc.assignedTo || 'Another Analyst';
      throw err;
    }

    const updated: FinancialException = {
      ...exc,
      assignedTo: analystId,
      ownerName: analystName,
      status: exc.status === 'detected' ? 'investigating' : exc.status,
      version: exc.version + 1,
      stateHistory: [
        ...exc.stateHistory,
        {
          from: exc.status,
          to: exc.status === 'detected' ? 'investigating' : exc.status,
          reasonCode: 'CLAIMED_BY_ANALYST',
          author: analystName,
          timestamp: new Date().toISOString(),
        },
      ],
    };

    exceptions = exceptions.map((e) => (e.id === exceptionId ? updated : e));
    return updated;
  },

  // ASSIGN exception
  async assignException(
    exceptionId: string,
    analystId: string,
    analystName: string,
    team: ExceptionTeam,
    expectedVersion: number
  ): Promise<FinancialException> {
    const exc = exceptions.find((e) => e.id === exceptionId);
    if (!exc) throw new Error('Exception not found');

    if (exc.version !== expectedVersion) {
      const err = new Error('CONCURRENCY_CONFLICT');
      (err as unknown as { currentVersion: number; updatedBy: string }).currentVersion = exc.version;
      (err as unknown as { currentVersion: number; updatedBy: string }).updatedBy = exc.assignedTo || 'Another Analyst';
      throw err;
    }

    const updated: FinancialException = {
      ...exc,
      assignedTo: analystId,
      ownerName: analystName,
      team,
      version: exc.version + 1,
    };

    exceptions = exceptions.map((e) => (e.id === exceptionId ? updated : e));
    return updated;
  },

  // UPDATE exception status with mandatory reason code
  async updateExceptionStatus(
    exceptionId: string,
    newStatus: FinancialStatus,
    reasonCode: string,
    author: string,
    justification: string,
    expectedVersion: number
  ): Promise<FinancialException> {
    const exc = exceptions.find((e) => e.id === exceptionId);
    if (!exc) throw new Error('Exception not found');

    if (exc.version !== expectedVersion) {
      const err = new Error('CONCURRENCY_CONFLICT');
      (err as unknown as { currentVersion: number; updatedBy: string }).currentVersion = exc.version;
      (err as unknown as { currentVersion: number; updatedBy: string }).updatedBy = exc.assignedTo || 'Another Analyst';
      throw err;
    }

    const updated: FinancialException = {
      ...exc,
      status: newStatus,
      version: exc.version + 1,
      stateHistory: [
        ...exc.stateHistory,
        {
          from: exc.status,
          to: newStatus,
          reasonCode,
          author,
          timestamp: new Date().toISOString(),
        },
      ],
      notes: [
        ...exc.notes,
        {
          id: `n-${Date.now()}`,
          author,
          authorRole: 'ANALYST_MAKER',
          timestamp: new Date().toISOString(),
          text: `Status changed to ${newStatus.toUpperCase()}: [${reasonCode}] ${justification}`,
        },
      ],
    };

    exceptions = exceptions.map((e) => (e.id === exceptionId ? updated : e));
    return updated;
  },

  // ADD note to exception
  async addExceptionNote(
    exceptionId: string,
    author: string,
    authorRole: Role,
    text: string
  ): Promise<FinancialException> {
    const exc = exceptions.find((e) => e.id === exceptionId);
    if (!exc) throw new Error('Exception not found');

    const newNote: ExceptionNote = {
      id: `note-${Date.now()}`,
      author,
      authorRole,
      timestamp: new Date().toISOString(),
      text,
    };

    const updated: FinancialException = {
      ...exc,
      version: exc.version + 1,
      notes: [...exc.notes, newNote],
    };

    exceptions = exceptions.map((e) => (e.id === exceptionId ? updated : e));
    return updated;
  },

  // LINK exception to incident
  async linkExceptionToIncident(
    exceptionId: string,
    incidentId: string,
    incidentTitle: string
  ): Promise<FinancialException> {
    const exc = exceptions.find((e) => e.id === exceptionId);
    if (!exc) throw new Error('Exception not found');

    const updated: FinancialException = {
      ...exc,
      incidentId,
      incidentTitle,
      version: exc.version + 1,
    };

    exceptions = exceptions.map((e) => (e.id === exceptionId ? updated : e));
    return updated;
  },

  // BULK assign exceptions
  async bulkAssignExceptions(
    exceptionIds: string[],
    analystId: string,
    analystName: string,
    team: ExceptionTeam
  ): Promise<FinancialException[]> {
    exceptions = exceptions.map((e) => {
      if (exceptionIds.includes(e.id)) {
        return {
          ...e,
          assignedTo: analystId,
          ownerName: analystName,
          team,
          version: e.version + 1,
        };
      }
      return e;
    });

    return exceptions.filter((e) => exceptionIds.includes(e.id));
  },

  // GET /investigations
  async getInvestigations(): Promise<InvestigationCase[]> {
    return [...investigations];
  },

  async getInvestigationCases(): Promise<InvestigationCase[]> {
    return this.getInvestigations();
  },

  // GET /investigations/:id
  async getInvestigationCase(id: string): Promise<InvestigationCase | null> {
    const found = investigations.find((c) => c.id === id || c.caseNumber === id || c.exceptionId === id);
    return found ? { ...found } : null;
  },

  // CONFIRM / EDIT Root Cause Tag
  async confirmRootCause(
    caseId: string,
    confirmedCause: RootCause,
    analyst: string = 'Olusola Adeyemi (Maker)'
  ): Promise<InvestigationCase> {
    const c = investigations.find((x) => x.id === caseId);
    if (!c) throw new Error('Case not found');

    const updated: InvestigationCase = {
      ...c,
      confirmedRootCause: confirmedCause,
      isRootCauseConfirmed: true,
      updatedAt: new Date().toISOString(),
      timeline: [
        ...c.timeline,
        {
          id: `tl-rc-${Date.now()}`,
          source: 'OPS',
          sourceName: 'Analyst Case Triage',
          sourceTimestamp: new Date().toISOString(),
          receivedTimestamp: new Date().toISOString(),
          clockSkewMs: 0,
          isLateArriving: false,
          title: `Root Cause Confirmed: ${confirmedCause}`,
          description: `Analyst ${analyst} confirmed root-cause diagnosis.`,
          payload: { confirmedCause, analyst },
          severity: 'SUCCESS',
        },
      ],
    };

    investigations = investigations.map((x) => (x.id === caseId ? updated : x));
    return updated;
  },

  // CONFIRM Duplicate Verdict
  async confirmDuplicateVerdict(
    caseId: string,
    verdict: 'CONFIRMED_DUPLICATE' | 'LEGITIMATE_RETRY',
    reason: string,
    analyst: string = 'Olusola Adeyemi (Maker)'
  ): Promise<InvestigationCase> {
    const c = investigations.find((x) => x.id === caseId);
    if (!c) throw new Error('Case not found');

    const updated: InvestigationCase = {
      ...c,
      status: verdict === 'CONFIRMED_DUPLICATE' ? 'approval' : 'resolved',
      duplicateComparison: c.duplicateComparison
        ? {
            ...c.duplicateComparison,
            ruleOutcome: verdict,
            verdictReason: reason,
          }
        : undefined,
      updatedAt: new Date().toISOString(),
      timeline: [
        ...c.timeline,
        {
          id: `tl-dup-${Date.now()}`,
          source: 'OPS',
          sourceName: 'Duplicate Decision Engine',
          sourceTimestamp: new Date().toISOString(),
          receivedTimestamp: new Date().toISOString(),
          clockSkewMs: 0,
          isLateArriving: false,
          title: `Duplicate Verdict: ${verdict}`,
          description: `Analyst ${analyst}: ${reason}`,
          payload: { verdict, reason, analyst },
          severity: verdict === 'CONFIRMED_DUPLICATE' ? 'WARNING' : 'SUCCESS',
        },
      ],
    };

    investigations = investigations.map((x) => (x.id === caseId ? updated : x));
    return updated;
  },

  // REQUEST Missing Evidence Check
  async requestMissingEvidenceCheck(
    caseId: string,
    evidenceId: string,
    targetSource: string,
    analyst: string = 'Olusola Adeyemi (Maker)'
  ): Promise<InvestigationCase> {
    const c = investigations.find((x) => x.id === caseId);
    if (!c) throw new Error('Case not found');

    const updatedEvidence = c.evidence.map((ev) =>
      ev.id === evidenceId
        ? {
            ...ev,
            freshness: 'FRESH' as const,
            freshnessLabel: 'Verification In-Flight',
            details: `${ev.details} [Check DISPATCHED to ${targetSource} by ${analyst}]`,
          }
        : ev
    );

    const updated: InvestigationCase = {
      ...c,
      evidence: updatedEvidence,
      updatedAt: new Date().toISOString(),
      timeline: [
        ...c.timeline,
        {
          id: `tl-ev-req-${Date.now()}`,
          source: 'OPS',
          sourceName: 'Counterparty Inquiry Service',
          sourceTimestamp: new Date().toISOString(),
          receivedTimestamp: new Date().toISOString(),
          clockSkewMs: 0,
          isLateArriving: false,
          title: `Evidence Check Dispatched to ${targetSource}`,
          description: `Automated query queued for missing counterparty leg (${evidenceId}).`,
          payload: { evidenceId, targetSource, analyst },
          severity: 'INFO',
        },
      ],
    };

    investigations = investigations.map((x) => (x.id === caseId ? updated : x));
    return updated;
  },

  // ADD External Dependency
  async addExternalDependency(
    caseId: string,
    dep: Omit<ExternalDependency, 'id'>
  ): Promise<InvestigationCase> {
    const c = investigations.find((x) => x.id === caseId);
    if (!c) throw new Error('Case not found');

    const newDep: ExternalDependency = {
      ...dep,
      id: `dep-${Date.now()}`,
      status: dep.status || 'PENDING',
    };

    const updated: InvestigationCase = {
      ...c,
      externalDependencies: [...c.externalDependencies, newDep],
      updatedAt: new Date().toISOString(),
    };

    investigations = investigations.map((x) => (x.id === caseId ? updated : x));
    return updated;
  },

  // CLOSE Investigation Case
  async closeInvestigationCase(
    caseId: string,
    resolutionCode: string,
    notes: string,
    analyst: string = 'Olusola Adeyemi (Maker)'
  ): Promise<InvestigationCase> {
    const c = investigations.find((x) => x.id === caseId);
    if (!c) throw new Error('Case not found');

    const updated: InvestigationCase = {
      ...c,
      status: 'resolved',
      exception: {
        ...c.exception,
        status: 'resolved',
        resolutionCode,
      },
      updatedAt: new Date().toISOString(),
      timeline: [
        ...c.timeline,
        {
          id: `tl-close-${Date.now()}`,
          source: 'OPS',
          sourceName: 'Case Resolution Terminal',
          sourceTimestamp: new Date().toISOString(),
          receivedTimestamp: new Date().toISOString(),
          clockSkewMs: 0,
          isLateArriving: false,
          title: `Case Closed: ${resolutionCode}`,
          description: `${analyst}: ${notes}`,
          payload: { resolutionCode, notes, analyst },
          severity: 'SUCCESS',
        },
      ],
    };

    investigations = investigations.map((x) => (x.id === caseId ? updated : x));
    return updated;
  },

  // GET /reversals
  async getReversals(): Promise<ReversalProposal[]> {
    return [...reversals];
  },

  // POST /reversals/propose (Maker Step)
  async proposeReversal(payload: ProposeReversalParams): Promise<ReversalProposal> {
    const txn = transactions.find((t) => t.id === payload.transactionId);
    if (!txn) throw new Error('Transaction not found');

    const newProposal: ReversalProposal = {
      id: `rev-prop-${Date.now()}`,
      transactionId: txn.id,
      reference: txn.reference,
      originalAmount: txn.amount,
      reversalAmount: txn.amount,
      currency: txn.currency,
      channel: txn.channel,
      status: 'approval', // Pending dual-control Checker
      proposedBy: payload.proposedBy,
      proposedAt: new Date().toISOString(),
      reasonCode: payload.reasonCode,
      justification: payload.justification,
      idempotencyKey: `idemp_${txn.reference.toLowerCase()}_${Math.random().toString(36).substring(2, 9)}`,
      destinationAccount: txn.senderAccount,
      destinationBank: txn.senderBank,
    };

    reversals = [newProposal, ...reversals];

    // Update transaction status
    transactions = transactions.map((t) =>
      t.id === txn.id ? { ...t, status: 'approval' } : t
    );

    // Audit Event
    auditTrail = [
      {
        id: `aud-${Date.now()}`,
        timestamp: new Date().toISOString(),
        actorId: payload.proposedBy,
        actorRole: 'ANALYST_MAKER',
        action: 'REVERSAL_PROPOSED',
        entityType: 'REVERSAL',
        entityId: newProposal.id,
        prevHash: auditTrail[0]?.entryHash || '0000',
        entryHash: Math.random().toString(36).substring(2, 18),
        details: { reference: txn.reference, amount: txn.amount, reason: payload.reasonCode },
        ipAddress: '127.0.0.1',
      },
      ...auditTrail,
    ];

    return newProposal;
  },

  // POST /reversals/{id}/authorize (Checker Step)
  async authorizeReversal(payload: AuthorizeReversalParams): Promise<ReversalProposal> {
    const prop = reversals.find((r) => r.id === payload.reversalId);
    if (!prop) throw new Error('Reversal proposal not found');

    const updatedStatus: FinancialStatus = payload.decision === 'APPROVE' ? 'resolved' : 'failed';

    const updated: ReversalProposal = {
      ...prop,
      status: updatedStatus,
      authorizedBy: payload.authorizedBy,
      authorizedAt: new Date().toISOString(),
      rejectionReason: payload.decision === 'REJECT' ? payload.notes : undefined,
    };

    reversals = reversals.map((r) => (r.id === prop.id ? updated : r));

    // Update linked transaction
    transactions = transactions.map((t) =>
      t.id === prop.transactionId ? { ...t, status: updatedStatus, hasDiscrepancy: false } : t
    );

    // Update linked exception if exists
    exceptions = exceptions.map((e) =>
      e.transactionId === prop.transactionId ? { ...e, status: updatedStatus } : e
    );

    // Immutable Audit Trail
    auditTrail = [
      {
        id: `aud-${Date.now()}`,
        timestamp: new Date().toISOString(),
        actorId: payload.authorizedBy,
        actorRole: 'SUPERVISOR_CHECKER',
        action: payload.decision === 'APPROVE' ? 'REVERSAL_AUTHORIZED_AND_DISPATCHED' : 'REVERSAL_REJECTED',
        entityType: 'REVERSAL',
        entityId: prop.id,
        prevHash: auditTrail[0]?.entryHash || '0000',
        entryHash: Math.random().toString(36).substring(2, 18),
        details: {
          reference: prop.reference,
          amount: prop.reversalAmount,
          decision: payload.decision,
          idempotencyKey: prop.idempotencyKey,
        },
        ipAddress: '127.0.0.1',
      },
      ...auditTrail,
    ];

    return updated;
  },

  // GET /audit-logs
  async getAuditLogs(): Promise<AuditEvent[]> {
    return [...auditTrail];
  },

  // GET /reconciliation/runs
  async getReconciliationRuns(): Promise<ReconciliationSummary[]> {
    return [...reconciliationRuns];
  },

  // Log Sensitive Data Reveal (Mandatory Audit Trail)
  async logDataReveal(field: string, entityId: string, actor: string, reason: string): Promise<void> {
    auditTrail = [
      {
        id: `aud-${Date.now()}`,
        timestamp: new Date().toISOString(),
        actorId: actor,
        actorRole: 'ANALYST_MAKER',
        action: 'SENSITIVE_DATA_REVEALED',
        entityType: 'TRANSACTION',
        entityId,
        prevHash: auditTrail[0]?.entryHash || '0000',
        entryHash: Math.random().toString(36).substring(2, 18),
        details: { field, justification: reason },
        ipAddress: '127.0.0.1',
      },
      ...auditTrail,
    ];
  },

  // Reset simulator
  resetSimulator() {
    transactions = [...INITIAL_TRANSACTIONS];
    exceptions = [...INITIAL_EXCEPTIONS];
    investigations = [...INITIAL_INVESTIGATIONS];
    reversals = [...INITIAL_REVERSALS];
    auditTrail = [...INITIAL_AUDIT_TRAIL];
  },
};
