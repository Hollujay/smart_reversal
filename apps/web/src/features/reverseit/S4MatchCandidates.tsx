import React, { useState } from 'react';
import {
  ArrowLeft,
  ArrowRight,
} from 'lucide-react';
import { Button } from '../../components/Button';
import type { ParsedAlertDetails, MatchedTransactionCandidate } from './reverseitTypes';

export interface S4MatchCandidatesProps {
  parsed: ParsedAlertDetails;
  onBack: () => void;
  onCandidateSelected: (candidate: MatchedTransactionCandidate | null) => void;
}

export const S4MatchCandidates: React.FC<S4MatchCandidatesProps> = ({
  parsed,
  onBack,
  onCandidateSelected,
}) => {
  // Generate realistic candidates based on parsed alert
  const [candidates] = useState<MatchedTransactionCandidate[]>([
    {
      id: 'tx-cand-01',
      reference: parsed.reference || 'NIP-DEMO-7781002',
      amount: parsed.amount,
      currency: parsed.currency,
      channel: parsed.channel || 'NIP',
      dateTime: parsed.dateTime,
      beneficiaryMasked: 'CHINEDU OKONKWO (GTB •••• 9921)',
      matchScore: 98,
      matchConfidence: 'HIGH',
      matchReason: 'Exact match on reference, amount, and account tail within 2 minutes',
      status: 'FAILED',
    },
    {
      id: 'tx-cand-02',
      reference: 'NIP-DEMO-7780998',
      amount: parsed.amount,
      currency: parsed.currency,
      channel: parsed.channel || 'NIP',
      dateTime: '14-Oct-2026 09:10:02 WAT',
      beneficiaryMasked: 'CHINEDU OKONKWO (GTB •••• 9921)',
      matchScore: 82,
      matchConfidence: 'MEDIUM',
      matchReason: 'Same beneficiary and amount 2 minutes prior; suspected retry candidate',
      status: 'PENDING',
    },
  ]);

  const [selectedId, setSelectedId] = useState<string | 'NONE'>(candidates[0]?.id || 'NONE');

  const handleProceed = () => {
    if (selectedId === 'NONE') {
      onCandidateSelected(null);
    } else {
      const selected = candidates.find((c) => c.id === selectedId) || null;
      onCandidateSelected(selected);
    }
  };

  return (
    <div className="max-w-2xl mx-auto py-4 px-4 space-y-6 animate-in fade-in duration-200">
      <button
        type="button"
        onClick={onBack}
        className="inline-flex items-center gap-1.5 text-[13px] text-[var(--text-secondary)] hover:text-[var(--text-primary)] font-medium cursor-pointer"
      >
        <ArrowLeft className="w-4 h-4" />
        <span>Back to details</span>
      </button>

      {/* Screen Title */}
      <div>
        <h2 className="text-[20px] font-bold text-[var(--text-primary)]">
          Step 3: We Found Matching Transactions
        </h2>
        <p className="text-[13px] text-[var(--text-muted)] mt-1">
          Our core reconciliation engine correlated your alert with recorded switch events.
          Please select the exact transaction you are disputing.
        </p>
      </div>

      {/* Candidate Cards List */}
      <div className="space-y-3">
        {candidates.map((cand) => (
          <div
            key={cand.id}
            onClick={() => setSelectedId(cand.id)}
            className={`p-4 rounded-xl border-2 transition-all cursor-pointer relative ${
              selectedId === cand.id
                ? 'bg-sky-50/60 border-sky-600 shadow-sm'
                : 'bg-[var(--surface)] border-[var(--surface-border)] hover:border-slate-300'
            }`}
          >
            <div className="flex items-start justify-between">
              <div>
                <div className="flex items-center gap-2">
                  <span className="font-bold text-[16px] font-mono-financial text-[var(--text-primary)]">
                    {cand.currency} {cand.amount}
                  </span>
                  <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-sky-100 text-sky-800">
                    {cand.channel} Channel
                  </span>
                </div>
                <div className="text-[12px] font-mono-financial text-[var(--text-secondary)] mt-1">
                  Ref: {cand.reference} &bull; {cand.dateTime}
                </div>
                <div className="text-[13px] font-medium text-[var(--text-primary)] mt-1">
                  Recipient: {cand.beneficiaryMasked}
                </div>
              </div>

              <div className="text-right">
                <span className="px-2 py-0.5 rounded-full text-[11px] font-bold font-mono-financial bg-emerald-100 text-emerald-800">
                  {cand.matchScore}% Match
                </span>
                <span className="block text-[11px] text-rose-700 font-semibold mt-1">
                  Status: {cand.status}
                </span>
              </div>
            </div>

            <p className="text-[11px] text-[var(--text-muted)] mt-2 pt-2 border-t border-slate-200/60">
              💡 {cand.matchReason}
            </p>
          </div>
        ))}

        {/* None of these option */}
        <div
          onClick={() => setSelectedId('NONE')}
          className={`p-4 rounded-xl border-2 transition-all cursor-pointer ${
            selectedId === 'NONE'
              ? 'bg-amber-50/60 border-amber-500 shadow-sm'
              : 'bg-[var(--surface)] border-[var(--surface-border)] hover:border-slate-300'
          }`}
        >
          <div className="flex items-center gap-3">
            <input
              type="radio"
              checked={selectedId === 'NONE'}
              onChange={() => setSelectedId('NONE')}
              className="w-4 h-4 text-amber-600"
            />
            <div>
              <span className="text-[14px] font-bold text-[var(--text-primary)] block">
                None of these match my transaction
              </span>
              <p className="text-[12px] text-[var(--text-muted)] mt-0.5">
                We will submit this as an Unmatched Claim (CUS001). Our operations team will pull the core banking ledger to locate your transaction.
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Navigation Buttons */}
      <div className="flex items-center justify-between pt-2">
        <Button variant="ghost" size="md" onClick={onBack}>
          Back
        </Button>
        <Button
          variant="primary"
          size="md"
          onClick={handleProceed}
          rightIcon={<ArrowRight className="w-4 h-4" />}
        >
          Confirm & Explain Issue
        </Button>
      </div>
    </div>
  );
};
