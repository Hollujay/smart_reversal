import React, { useState } from 'react';
import { ShieldCheck, ShieldAlert } from 'lucide-react';
import { Button } from '../components/Button';
import { StatusBadge } from '../components/StatusBadge';
import { ApprovalDialog } from '../components/ApprovalDialog';
import { IrreversibleConfirmDialog } from '../components/IrreversibleConfirmDialog';
import { useApp } from '../context/AppContext';
import { apiClient } from '../services/apiClient';
import { INITIAL_REVERSALS } from '../services/mockSimulatorData';
import type { ReversalProposal } from '../types/domain';

export const ReversalsView: React.FC = () => {
  const { role, addToast } = useApp();
  const [reversals, setReversals] = useState<ReversalProposal[]>(() => INITIAL_REVERSALS);
  const [selectedProposal, setSelectedProposal] = useState<ReversalProposal | null>(null);
  const [isApprovalOpen, setIsApprovalOpen] = useState(false);
  const [isIrreversibleOpen, setIsIrreversibleOpen] = useState(false);
  const [pendingAction, setPendingAction] = useState<{ id: string; decision: 'APPROVE' | 'REJECT'; notes?: string } | null>(null);

  const fetchReversals = async () => {
    const list = await apiClient.getReversals();
    setReversals(list);
  };

  const handleOpenReview = (proposal: ReversalProposal) => {
    setSelectedProposal(proposal);
    setIsApprovalOpen(true);
  };

  const handleAuthorizeStep = async (proposalId: string, decision: 'APPROVE' | 'REJECT', notes?: string) => {
    if (decision === 'APPROVE') {
      // Require irreversible confirmation dialog
      setPendingAction({ id: proposalId, decision, notes });
      setIsIrreversibleOpen(true);
    } else {
      await apiClient.authorizeReversal({
        reversalId: proposalId,
        authorizedBy: `Supervisor (${role})`,
        decision: 'REJECT',
        notes,
      });
      addToast({
        type: 'warning',
        title: 'Reversal Proposal Rejected',
        message: 'The proposed reversal was rejected and logged to the audit trail.',
      });
      fetchReversals();
    }
  };

  const handleFinalIrreversibleConfirm = async () => {
    if (!pendingAction) return;
    await apiClient.authorizeReversal({
      reversalId: pendingAction.id,
      authorizedBy: `Supervisor (${role})`,
      decision: 'APPROVE',
      notes: pendingAction.notes,
    });
    addToast({
      type: 'success',
      title: 'Reversal Dispatched & Settled',
      message: 'Dual-control authorization complete. Reversal payload dispatched to Core Switch.',
    });
    setPendingAction(null);
    fetchReversals();
  };

  const isChecker = role === 'SUPERVISOR_CHECKER' || role === 'ADMIN';

  return (
    <div className="space-y-6">
      {/* Header Info */}
      <div className="flex flex-wrap items-center justify-between gap-4 p-4 bg-[var(--surface)] border border-[var(--surface-border)] rounded-md shadow-xs">
        <div>
          <h2 className="text-[16px] font-bold text-[var(--text-primary)] flex items-center gap-2">
            <ShieldAlert className="w-5 h-5 text-purple-700" />
            <span>Maker-Checker Controlled Financial Reversals</span>
          </h2>
          <p className="text-[13px] text-[var(--text-muted)] mt-0.5">
            Strict dual-authorization protocol: No single operator can both propose and authorize a financial reversal.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <div className="text-[12px] px-3 py-1 bg-[var(--canvas-subdued)] border border-[var(--surface-border)] rounded font-mono-financial">
            Active Role: <strong className="text-[var(--text-primary)]">{role}</strong>
          </div>
        </div>
      </div>

      {/* Reversals List */}
      <div className="border border-[var(--surface-border)] rounded-md bg-[var(--surface)] shadow-xs divide-y divide-[var(--surface-border)]">
        <div className="px-4 py-3 bg-[var(--canvas-subdued)] font-semibold text-[13px] text-[var(--text-secondary)] flex justify-between">
          <span>Active Dual-Authorization Proposals ({reversals.length})</span>
          <span className="font-mono-financial text-[11px] text-[var(--text-muted)]">
            Idempotency Cryptographically Guaranteed
          </span>
        </div>

        <div className="divide-y divide-[var(--surface-border-subtle)]">
          {reversals.map((rev) => {
            const isPending = rev.status === 'approval';
            return (
              <div
                key={rev.id}
                className="p-4 hover:bg-[var(--surface-raised)] transition-colors flex flex-wrap items-center justify-between gap-4 text-[13px]"
              >
                <div className="space-y-1 max-w-xl">
                  <div className="flex items-center gap-2">
                    <strong className="font-mono-financial text-[14px] text-[var(--text-primary)]">
                      {rev.reference}
                    </strong>
                    <span className="text-[11px] font-mono-financial px-1.5 py-0.2 rounded bg-[var(--canvas-subdued)] border border-[var(--surface-border)]">
                      {rev.channel}
                    </span>
                    <StatusBadge status={rev.status} size="sm" />
                  </div>

                  <p className="text-[12px] text-[var(--text-secondary)] leading-relaxed">
                    Reason: <strong>{rev.reasonCode}</strong> • "{rev.justification}"
                  </p>

                  <div className="text-[11px] text-[var(--text-muted)] font-mono-financial flex flex-wrap items-center gap-3">
                    <span>Drafted by: {rev.proposedBy}</span>
                    <span>• At: {new Date(rev.proposedAt).toLocaleTimeString()}</span>
                    <span>• Idempotency Key: {rev.idempotencyKey.substring(0, 18)}...</span>
                  </div>
                </div>

                <div className="flex items-center gap-4">
                  <div className="text-right">
                    <div className="text-[11px] text-[var(--text-muted)] font-medium">Reversal Amount</div>
                    <div className="text-[18px] font-bold font-mono-financial text-[var(--text-primary)]">
                      {rev.currency} {Number(rev.reversalAmount).toLocaleString('en-US', { minimumFractionDigits: 2 })}
                    </div>
                  </div>

                  {isPending ? (
                    <Button
                      variant={isChecker ? 'primary' : 'secondary'}
                      size="sm"
                      onClick={() => handleOpenReview(rev)}
                      leftIcon={<ShieldCheck className="w-4 h-4" />}
                    >
                      {isChecker ? 'Review & Authorize' : 'View Proposal Dossier'}
                    </Button>
                  ) : (
                    <span className="px-2.5 py-1 text-[11px] font-semibold rounded bg-emerald-50 text-emerald-900 border border-emerald-300 font-mono-financial">
                      Executed by {rev.authorizedBy?.split(' ')[0]}
                    </span>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Maker-Checker Review Dialog */}
      <ApprovalDialog
        proposal={selectedProposal}
        isOpen={isApprovalOpen}
        onClose={() => setIsApprovalOpen(false)}
        onAuthorize={handleAuthorizeStep}
      />

      {/* Irreversible Confirm Dialog */}
      <IrreversibleConfirmDialog
        isOpen={isIrreversibleOpen}
        onClose={() => setIsIrreversibleOpen(false)}
        onConfirm={handleFinalIrreversibleConfirm}
        title="High-Stakes Financial Reversal Dispatch"
        description="Executing this transaction will permanently credit the originator account and post an irreversible reversal journal entry to Core Banking Switch ledger. This action cannot be undone."
        confirmKeyword="CONFIRM REVERSAL"
        actionButtonText="Permanently Execute Reversal"
      />
    </div>
  );
};
