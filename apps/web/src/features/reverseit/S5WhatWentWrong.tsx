import React, { useState } from 'react';
import {
  ArrowLeft,
  ArrowRight,
  AlertTriangle,
  Repeat,
  CreditCard,
  Building,
  HelpCircle,
  Sparkles,
} from 'lucide-react';
import { Button } from '../../components/Button';
import type { DisputeType, MatchedTransactionCandidate } from './reverseitTypes';

export interface S5WhatWentWrongProps {
  candidate: MatchedTransactionCandidate | null;
  onBack: () => void;
  onDisputeSelected: (type: DisputeType) => void;
}

export const S5WhatWentWrong: React.FC<S5WhatWentWrongProps> = ({
  candidate,
  onBack,
  onDisputeSelected,
}) => {
  // Preselect based on channel
  const initialType: DisputeType =
    candidate?.channel === 'POS'
      ? 'pos_web_failure'
      : candidate?.channel === 'ATM'
      ? 'atm_failure'
      : 'failed_transfer';

  const [selectedType, setSelectedType] = useState<DisputeType>(initialType);

  const options: Array<{
    type: DisputeType;
    title: string;
    description: string;
    icon: React.ReactNode;
    warning?: string;
  }> = [
    {
      type: 'failed_transfer',
      title: 'Failed Transfer (Debit Without Credit)',
      description: 'Money was deducted from your account, but the recipient confirms they did not receive it.',
      icon: <ArrowRight className="w-5 h-5 text-sky-600" />,
    },
    {
      type: 'duplicate_debit',
      title: 'Duplicate Debit',
      description: 'You were debited two or more times for a single payment, purchase, or transfer attempt.',
      icon: <Repeat className="w-5 h-5 text-indigo-600" />,
    },
    {
      type: 'pos_web_failure',
      title: 'POS Terminal or Web Checkout Failure',
      description: 'The POS terminal printed "Declined" or timed out, but your account was still debited.',
      icon: <CreditCard className="w-5 h-5 text-purple-600" />,
    },
    {
      type: 'atm_failure',
      title: 'ATM Cash Not Dispensed / Partial Dispense',
      description: 'The ATM failed to dispense cash or dispensed an incomplete amount after debiting your balance.',
      icon: <Building className="w-5 h-5 text-emerald-600" />,
    },
    {
      type: 'wrong_account_or_amount',
      title: 'Wrong Account Number or Wrong Amount Sent',
      description: 'You mistakenly entered the wrong account number or transferred more than intended.',
      icon: <HelpCircle className="w-5 h-5 text-amber-600" />,
      warning:
        'Note: Wrong recipient errors require an interbank recall request and depend on the receiving customer funds availability.',
    },
  ];

  return (
    <div className="max-w-2xl mx-auto py-4 px-4 space-y-6 animate-in fade-in duration-200">
      <button
        type="button"
        onClick={onBack}
        className="inline-flex items-center gap-1.5 text-[13px] text-[var(--text-secondary)] hover:text-[var(--text-primary)] font-medium cursor-pointer"
      >
        <ArrowLeft className="w-4 h-4" />
        <span>Back to candidates</span>
      </button>

      {/* Screen Title */}
      <div>
        <h2 className="text-[20px] font-bold text-[var(--text-primary)]">
          Step 4: What Went Wrong?
        </h2>
        <p className="text-[13px] text-[var(--text-muted)] mt-1">
          Select the exact nature of the failure. This maps deterministically to your Central Bank of Nigeria (CBN) regulatory countdown.
        </p>
      </div>

      {/* AI Recommendation Pill */}
      <div className="p-3 bg-sky-50 border border-sky-200 rounded-lg flex items-center justify-between text-[12px] text-sky-900">
        <div className="flex items-center gap-2">
          <Sparkles className="w-4 h-4 text-sky-700" />
          <span>
            <strong>AI Suggestion: </strong>
            We recommend <strong>"{options.find((o) => o.type === initialType)?.title}"</strong> based on switch response codes.
          </span>
        </div>
        <span className="text-[11px] font-semibold text-sky-800">Preselected</span>
      </div>

      {/* Dispute Options Radio List */}
      <div className="space-y-3">
        {options.map((opt) => (
          <div
            key={opt.type}
            onClick={() => setSelectedType(opt.type)}
            className={`p-4 rounded-xl border-2 transition-all cursor-pointer ${
              selectedType === opt.type
                ? 'bg-[var(--surface)] border-sky-600 shadow-sm ring-1 ring-sky-600'
                : 'bg-[var(--surface)] border-[var(--surface-border)] hover:border-slate-300'
            }`}
          >
            <div className="flex items-start gap-3.5">
              <input
                type="radio"
                name="dispute-type"
                checked={selectedType === opt.type}
                onChange={() => setSelectedType(opt.type)}
                className="mt-1 w-4 h-4 text-sky-600 focus:ring-sky-500 cursor-pointer"
              />
              <div className="space-y-1 flex-1">
                <div className="flex items-center gap-2">
                  <span className="font-bold text-[14px] text-[var(--text-primary)]">
                    {opt.title}
                  </span>
                </div>
                <p className="text-[12px] text-[var(--text-muted)] leading-relaxed">
                  {opt.description}
                </p>
                {opt.warning && selectedType === opt.type && (
                  <div className="mt-2 p-2.5 rounded bg-amber-50 text-amber-900 border border-amber-200 text-[11px] flex items-start gap-2">
                    <AlertTriangle className="w-4 h-4 text-amber-700 shrink-0 mt-0.5" />
                    <span>{opt.warning}</span>
                  </div>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Navigation Buttons */}
      <div className="flex items-center justify-between pt-2">
        <Button variant="ghost" size="md" onClick={onBack}>
          Back
        </Button>
        <Button
          variant="primary"
          size="md"
          onClick={() => onDisputeSelected(selectedType)}
          rightIcon={<ArrowRight className="w-4 h-4" />}
        >
          View Regulatory Deadline
        </Button>
      </div>
    </div>
  );
};
