import {
  Clock,
  CheckCircle2,
  ArrowLeft,
  ShieldCheck,
  ExternalLink,
} from 'lucide-react';
import { Button } from '../../components/Button';
import type { CustomerDisputeCase } from './reverseitTypes';

export interface S8CaseTrackerProps {
  disputeCase: CustomerDisputeCase;
  onBack: () => void;
  onViewEscalation: () => void;
}

export const S8CaseTracker: React.FC<S8CaseTrackerProps> = ({
  disputeCase,
  onBack,
  onViewEscalation,
}) => {
  return (
    <div className="max-w-2xl mx-auto py-4 px-4 space-y-6 animate-in fade-in duration-200">
      <button
        type="button"
        onClick={onBack}
        className="inline-flex items-center gap-1.5 text-[13px] text-[var(--text-secondary)] hover:text-[var(--text-primary)] font-medium cursor-pointer"
      >
        <ArrowLeft className="w-4 h-4" />
        <span>Back to My Cases</span>
      </button>

      {/* Case Header Card */}
      <div className="p-5 bg-[var(--surface)] border border-[var(--surface-border)] rounded-xl shadow-xs space-y-3">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div>
            <div className="flex items-center gap-2">
              <h2 className="text-[18px] font-bold text-[var(--text-primary)] font-mono-financial">
                {disputeCase.caseNumber}
              </h2>
              <span className="px-2 py-0.5 rounded text-[11px] font-bold uppercase tracking-wider bg-sky-100 text-sky-800">
                {disputeCase.channel}
              </span>
            </div>
            <p className="text-[12px] text-[var(--text-muted)] mt-0.5">
              Ref: {disputeCase.reference} &bull; Reported {new Date(disputeCase.reportedAt).toLocaleDateString()}
            </p>
          </div>

          <div className="text-right">
            <span className="text-[20px] font-bold font-mono-financial text-[var(--text-primary)] block">
              {disputeCase.currency} {disputeCase.amount}
            </span>
            <span className="text-[11px] font-semibold text-sky-700 bg-sky-50 px-2 py-0.5 rounded border border-sky-200">
              {disputeCase.statusLabel}
            </span>
          </div>
        </div>

        {/* Customer Safe Notice Box (Section 10.6 strictly observed) */}
        <div className="p-3 rounded-lg bg-sky-50/70 border border-sky-200 text-[12px] text-sky-950 flex items-start gap-2.5">
          <ShieldCheck className="w-4 h-4 text-sky-700 shrink-0 mt-0.5" />
          <div>
            <strong className="font-semibold">Official Status Notice: </strong>
            <span>{disputeCase.customerSafeStatusText}</span>
          </div>
        </div>
      </div>

      {/* Live SLA & Regulatory Deadlines */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-[12px]">
        <div className="p-4 rounded-xl bg-[var(--surface)] border border-[var(--surface-border)] space-y-1">
          <span className="text-[10px] font-bold uppercase tracking-wider text-[var(--text-muted)] flex items-center gap-1">
            <Clock className="w-3.5 h-3.5 text-sky-600" />
            <span>Bank Target Window</span>
          </span>
          <span className="text-[14px] font-bold font-mono-financial text-[var(--text-primary)] block">
            {disputeCase.timeRemainingText}
          </span>
          <p className="text-[11px] text-[var(--text-muted)]">
            Target: {new Date(disputeCase.bankTargetDeadline).toLocaleTimeString()} WAT
          </p>
        </div>

        <div className="p-4 rounded-xl bg-[var(--surface)] border border-[var(--surface-border)] space-y-1">
          <span className="text-[10px] font-bold uppercase tracking-wider text-[var(--text-muted)] flex items-center gap-1">
            <Clock className="w-3.5 h-3.5 text-emerald-600" />
            <span>CBN Regulatory Max</span>
          </span>
          <span className="text-[14px] font-bold font-mono-financial text-emerald-700 block">
            Mandated Statutory Clock
          </span>
          <p className="text-[11px] text-[var(--text-muted)]">
            Deadline: {new Date(disputeCase.regulatoryDeadline).toLocaleTimeString()} WAT
          </p>
        </div>
      </div>

      {/* Vertical Status Timeline (Screen S8) */}
      <div className="p-5 bg-[var(--surface)] border border-[var(--surface-border)] rounded-xl space-y-4">
        <h3 className="text-[14px] font-bold text-[var(--text-primary)]">
          Case Investigation Lifecycle
        </h3>

        <div className="space-y-6 relative before:absolute before:inset-0 before:left-3.5 before:w-0.5 before:bg-slate-200">
          {disputeCase.history.map((hist, i) => (
            <div key={hist.stage} className="relative flex items-start gap-4">
              <span
                className={`w-7 h-7 rounded-full flex items-center justify-center shrink-0 z-10 font-bold text-[11px] ${
                  hist.verified
                    ? 'bg-sky-600 text-white shadow-xs'
                    : 'bg-slate-100 text-slate-400 border border-slate-300'
                }`}
              >
                {hist.verified ? <CheckCircle2 className="w-4 h-4" /> : i + 1}
              </span>
              <div className="space-y-0.5 pt-0.5">
                <div className="flex items-center gap-2">
                  <span className="font-bold text-[13px] text-[var(--text-primary)]">
                    {hist.label}
                  </span>
                  <span className="text-[11px] font-mono-financial text-[var(--text-muted)]">
                    {hist.timestamp}
                  </span>
                </div>
                <p className="text-[12px] text-[var(--text-secondary)] leading-relaxed">
                  {hist.details}
                </p>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Escalation Ladder Action (Screen S9) */}
      <div className="p-4 bg-[var(--canvas-subdued)] border border-[var(--surface-border)] rounded-xl flex items-center justify-between">
        <div>
          <span className="text-[13px] font-bold text-[var(--text-primary)] block">
            Resolution Past Target or Unresolved?
          </span>
          <p className="text-[11px] text-[var(--text-muted)]">
            View the official 3-step escalation ladder (Bank Ombudsman & CBN Consumer Protection).
          </p>
        </div>
        <Button
          size="sm"
          variant="secondary"
          onClick={onViewEscalation}
          rightIcon={<ExternalLink className="w-3.5 h-3.5" />}
        >
          View Escalation Ladder
        </Button>
      </div>
    </div>
  );
};
