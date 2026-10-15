import React, { useState, useEffect } from 'react';
import {
  Clock,
  ArrowLeft,
  ArrowRight,
} from 'lucide-react';
import { Button } from '../../components/Button';
import { REGULATORY_CLOCKS } from './reverseitData';
import type { DisputeType, RegulatoryClockInfo } from './reverseitTypes';

export interface S6YourDeadlineProps {
  disputeType: DisputeType;
  onBack: () => void;
  onProceedToComplaint: (clock: RegulatoryClockInfo) => void;
}

export const S6YourDeadline: React.FC<S6YourDeadlineProps> = ({
  disputeType,
  onBack,
  onProceedToComplaint,
}) => {
  const clock = REGULATORY_CLOCKS[disputeType];

  // Live countdown state
  const [secondsRemaining, setSecondsRemaining] = useState(
    clock.bankTargetHours * 3600 - 1450
  );

  useEffect(() => {
    const timer = setInterval(() => {
      setSecondsRemaining((prev) => Math.max(0, prev - 1));
    }, 1000);
    return () => clearInterval(timer);
  }, []);

  const hours = Math.floor(secondsRemaining / 3600);
  const minutes = Math.floor((secondsRemaining % 3600) / 60);
  const seconds = secondsRemaining % 60;

  return (
    <div className="max-w-2xl mx-auto py-4 px-4 space-y-6 animate-in fade-in duration-200">
      <button
        type="button"
        onClick={onBack}
        className="inline-flex items-center gap-1.5 text-[13px] text-[var(--text-secondary)] hover:text-[var(--text-primary)] font-medium cursor-pointer"
      >
        <ArrowLeft className="w-4 h-4" />
        <span>Change dispute type</span>
      </button>

      {/* Screen Title */}
      <div>
        <h2 className="text-[20px] font-bold text-[var(--text-primary)]">
          Step 5: Your Official Resolution Deadline
        </h2>
        <p className="text-[13px] text-[var(--text-muted)] mt-1">
          In accordance with Central Bank of Nigeria (CBN) directives, your bank is mandated 
          to resolve this transaction within the legal timeframe below.
        </p>
      </div>

      {/* Large Live Countdown Display */}
      <div className="p-6 rounded-2xl bg-gradient-to-br from-slate-900 to-indigo-950 text-white text-center shadow-lg border border-slate-700/60 space-y-4">
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-sky-500/20 text-sky-300 text-[12px] font-semibold border border-sky-400/30">
          <Clock className="w-4 h-4 animate-pulse" />
          <span>Active Countdown Clock (WAT)</span>
        </div>

        <div>
          <div className="text-[44px] md:text-[52px] font-extrabold font-mono-financial tracking-tight text-sky-300">
            {hours.toString().padStart(2, '0')}h : {minutes.toString().padStart(2, '0')}m :{' '}
            {seconds.toString().padStart(2, '0')}s
          </div>
          <p className="text-[13px] text-slate-300 font-medium mt-1">
            Bank Target Window (Accelerated SLA)
          </p>
        </div>

        {/* Dual Deadline Comparison Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-3 border-t border-slate-800 text-left text-[12px]">
          <div className="p-3 rounded-lg bg-slate-800/60 border border-slate-700/60 space-y-1">
            <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400 block">
              1. Bank Target SLA (Internal)
            </span>
            <span className="text-[14px] font-bold text-white font-mono-financial block">
              {clock.bankTargetLabel}
            </span>
            <span className="text-[11px] text-sky-400 block">
              Automated status check ladder running
            </span>
          </div>

          <div className="p-3 rounded-lg bg-slate-800/60 border border-slate-700/60 space-y-1">
            <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400 block">
              2. CBN Maximum Legal Clock
            </span>
            <span className="text-[14px] font-bold text-emerald-400 font-mono-financial block">
              {clock.regulatoryLabel}
            </span>
            <span className="text-[11px] text-slate-300 block">
              Mandatory statutory refund threshold
            </span>
          </div>
        </div>
      </div>

      {/* Regulatory Source Transparency */}
      <div className="p-4 bg-[var(--surface)] border border-[var(--surface-border)] rounded-xl space-y-2 text-[12px]">
        <div className="flex items-center justify-between text-[11px] font-bold uppercase tracking-wider text-[var(--text-muted)]">
          <span>Statutory Authority & Policy Lineage</span>
          <span className="text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded border border-emerald-200">
            {clock.status}
          </span>
        </div>
        <p className="text-[13px] text-[var(--text-primary)] font-medium">
          {clock.sourceAuthority}
        </p>
        <p className="text-[11px] text-[var(--text-muted)]">
          Effective Date: {clock.effectiveDate} &bull; Clock starts from debit detection time in WAT.
        </p>
      </div>

      {/* What Happens Next Guidance */}
      <div className="p-4 bg-[var(--canvas-subdued)] border border-[var(--surface-border)] rounded-xl space-y-2 text-[12px]">
        <h4 className="font-bold text-[13px] text-[var(--text-primary)]">
          What happens next?
        </h4>
        <ul className="space-y-1.5 text-[var(--text-muted)] list-disc pl-4 leading-relaxed">
          <li>We will auto-generate a verifiable formal complaint letter citing your exact transaction IDs.</li>
          <li>Your complaint will be queued as <strong>CUS001</strong> directly in the bank's exception workbench.</li>
          <li>If the receiving bank confirms the credit failed, your refund will post back to your account automatically.</li>
        </ul>
      </div>

      {/* Navigation Buttons */}
      <div className="flex items-center justify-between pt-2">
        <Button variant="ghost" size="md" onClick={onBack}>
          Back
        </Button>
        <Button
          variant="primary"
          size="md"
          onClick={() => onProceedToComplaint(clock)}
          rightIcon={<ArrowRight className="w-4 h-4" />}
        >
          Generate Formal Complaint
        </Button>
      </div>
    </div>
  );
};
