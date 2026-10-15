import React, { useState } from 'react';
import {
  ArrowLeft,
  Mail,
  Phone,
  Copy,
  Check,
} from 'lucide-react';
import { Button } from '../../components/Button';
import type { CustomerDisputeCase } from './reverseitTypes';

export interface S9EscalationLadderProps {
  disputeCase: CustomerDisputeCase;
  onBack: () => void;
}

export const S9EscalationLadder: React.FC<S9EscalationLadderProps> = ({
  disputeCase,
  onBack,
}) => {
  const [copiedStep, setCopiedStep] = useState<number | null>(null);

  const handleCopyText = (stepNum: number, text: string) => {
    navigator.clipboard.writeText(text);
    setCopiedStep(stepNum);
    setTimeout(() => setCopiedStep(null), 2000);
  };

  const steps = [
    {
      step: 1,
      title: 'Tier 1: Bank Transaction Dispute Desk',
      owner: 'Customer Care & Exception Desk',
      email: 'disputes@wemabank.com / purpleconnect@wemabank.com',
      phone: '0700-225-52528 (0700-ALAT-ALAT)',
      description:
        'First point of escalation if your transaction has passed the initial Bank Target window.',
      template: `ESCALATION: UNRESOLVED TRANSACTION EXCEPTION\nCase Reference: ${disputeCase.caseNumber}\nTransaction Reference: ${disputeCase.reference}\nAmount: ${disputeCase.currency} ${disputeCase.amount}\nDate: ${new Date(disputeCase.reportedAt).toLocaleDateString()}\nStatus: Awaiting bank resolution past SLA. Kindly expedite authoritative TSQ re-query.`,
    },
    {
      step: 2,
      title: 'Tier 2: Head Office Complaints & Ombudsman',
      owner: 'Customer Experience & Quality Assurance Unit',
      email: 'complaints@wemabank.com',
      phone: 'Direct Line: +234 1 277 8600',
      description:
        'Engage if Tier 1 has not resolved the dispute or if communication has stalled after 24 hours.',
      template: `TIER 2 ESCALATION: FORMAL COMPLAINT\nTracking ID: ${disputeCase.trackingId}\nCase ID: ${disputeCase.caseNumber}\nAmount: ${disputeCase.currency} ${disputeCase.amount}\nChannel: ${disputeCase.channel}\nAttention: Head of Operations / Quality Assurance\nRequesting immediate supervisory intervention for unresolved electronic debit.`,
    },
    {
      step: 3,
      title: 'Tier 3: Central Bank of Nigeria (CBN) CPD',
      owner: 'Consumer Protection Department, Central Bank of Nigeria',
      email: 'cpd@cbn.gov.ng',
      phone: 'CBN Head Office, Central Business District, Abuja',
      description:
        'Official statutory escalation for transactions exceeding maximum regulatory resolution timelines (NIP 24h / POS 72h).',
      template: `CENTRAL BANK OF NIGERIA — CONSUMER PROTECTION ESCALATION\nTo: Director, Consumer Protection Department (cpd@cbn.gov.ng)\nComplainant Account: ${disputeCase.maskedAccount}\nBank: Wema Bank Plc\nTransaction Reference: ${disputeCase.reference}\nAmount: ${disputeCase.currency} ${disputeCase.amount}\nOriginal Incident Date: ${new Date(disputeCase.reportedAt).toLocaleDateString()}\nStatutory Circular Violation: Failed Instant Payment / Electronic Dispense exceeding regulatory SLA deadline.`,
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
        <span>Back to case timeline</span>
      </button>

      {/* Screen Title */}
      <div>
        <h2 className="text-[20px] font-bold text-[var(--text-primary)]">
          Official 3-Step Escalation Ladder
        </h2>
        <p className="text-[13px] text-[var(--text-muted)] mt-1">
          If your transaction exceeds the statutory deadline without verified refund or credit advice, 
          follow this approved hierarchy in order.
        </p>
      </div>

      {/* Case Quick Pill */}
      <div className="p-3 bg-[var(--surface)] border border-[var(--surface-border)] rounded-lg flex items-center justify-between text-[12px]">
        <div className="font-mono-financial text-[var(--text-secondary)]">
          Disputing: <strong>{disputeCase.caseNumber}</strong> ({disputeCase.currency} {disputeCase.amount})
        </div>
        <span className="font-semibold text-rose-700 bg-rose-50 px-2 py-0.5 rounded border border-rose-200">
          Escalation Pack Ready
        </span>
      </div>

      {/* 3 Step Cards */}
      <div className="space-y-4">
        {steps.map((s) => (
          <div
            key={s.step}
            className="p-5 bg-[var(--surface)] border border-[var(--surface-border)] rounded-xl shadow-xs space-y-3"
          >
            <div className="flex items-start justify-between">
              <div className="flex items-center gap-2.5">
                <span className="w-6 h-6 rounded-full bg-slate-900 text-white flex items-center justify-center font-bold text-[12px]">
                  {s.step}
                </span>
                <h3 className="text-[14px] font-bold text-[var(--text-primary)]">
                  {s.title}
                </h3>
              </div>
              <span className="text-[11px] font-semibold text-slate-500 uppercase">
                {s.owner}
              </span>
            </div>

            <p className="text-[12px] text-[var(--text-muted)] leading-relaxed">
              {s.description}
            </p>

            <div className="p-3 rounded-lg bg-[var(--canvas-subdued)] space-y-1.5 text-[11px] font-mono-financial text-[var(--text-secondary)]">
              <div className="flex items-center gap-2">
                <Mail className="w-3.5 h-3.5 text-sky-600 shrink-0" />
                <span>{s.email}</span>
              </div>
              <div className="flex items-center gap-2">
                <Phone className="w-3.5 h-3.5 text-emerald-600 shrink-0" />
                <span>{s.phone}</span>
              </div>
            </div>

            {/* Template & Copy Button */}
            <div className="flex items-center justify-between pt-1">
              <span className="text-[11px] text-[var(--text-muted)]">
                Pre-formatted escalation notice ready
              </span>
              <Button
                size="xs"
                variant="secondary"
                onClick={() => handleCopyText(s.step, s.template)}
                leftIcon={copiedStep === s.step ? <Check className="w-3.5 h-3.5 text-emerald-600" /> : <Copy className="w-3.5 h-3.5" />}
              >
                {copiedStep === s.step ? 'Copied' : 'Copy Message Text'}
              </Button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
