import {
  Clock,
  ArrowRight,
  Plus,
} from 'lucide-react';
import { Button } from '../../components/Button';
import type { CustomerDisputeCase } from './reverseitTypes';

export interface MyCasesViewProps {
  cases: CustomerDisputeCase[];
  onSelectCase: (disputeCase: CustomerDisputeCase) => void;
  onStartNewReport: () => void;
}

export const MyCasesView: React.FC<MyCasesViewProps> = ({
  cases,
  onSelectCase,
  onStartNewReport,
}) => {
  return (
    <div className="max-w-3xl mx-auto py-4 px-4 space-y-6 animate-in fade-in duration-200">
      {/* Title & Action */}
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h2 className="text-[20px] font-bold text-[var(--text-primary)]">
            My Dispute Cases ({cases.length})
          </h2>
          <p className="text-[13px] text-[var(--text-muted)] mt-0.5">
            Real-time status tracking and Central Bank of Nigeria (CBN) resolution clocks.
          </p>
        </div>
        <Button
          variant="primary"
          size="sm"
          onClick={onStartNewReport}
          leftIcon={<Plus className="w-4 h-4" />}
        >
          Report Another Issue
        </Button>
      </div>

      {/* Cases List */}
      {cases.length === 0 ? (
        <div className="p-12 text-center bg-[var(--surface)] border border-[var(--surface-border)] rounded-xl space-y-3">
          <Clock className="w-12 h-12 mx-auto text-slate-300" />
          <h3 className="text-[16px] font-bold text-[var(--text-primary)]">
            No active disputes filed
          </h3>
          <p className="text-[13px] text-[var(--text-muted)] max-w-sm mx-auto">
            You currently have no open transaction investigations. Report an alert anytime something goes wrong.
          </p>
          <Button variant="primary" size="md" onClick={onStartNewReport}>
            Report a Transaction Problem
          </Button>
        </div>
      ) : (
        <div className="space-y-3">
          {cases.map((c) => (
            <div
              key={c.id}
              onClick={() => onSelectCase(c)}
              className="p-5 bg-[var(--surface)] border border-[var(--surface-border)] hover:border-sky-500 rounded-xl shadow-xs hover:shadow-sm transition-all cursor-pointer space-y-3 group"
            >
              <div className="flex flex-wrap items-start justify-between gap-2">
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-bold text-[15px] font-mono-financial text-[var(--text-primary)] group-hover:text-sky-700">
                      {c.caseNumber}
                    </span>
                    <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-sky-100 text-sky-800">
                      {c.channel}
                    </span>
                  </div>
                  <div className="text-[12px] font-mono-financial text-[var(--text-muted)] mt-0.5">
                    Ref: {c.reference} &bull; Reported {new Date(c.reportedAt).toLocaleDateString()}
                  </div>
                </div>

                <div className="text-right">
                  <span className="text-[18px] font-bold font-mono-financial text-[var(--text-primary)] block">
                    {c.currency} {c.amount}
                  </span>
                  <span className="text-[11px] font-semibold text-sky-700 bg-sky-50 px-2 py-0.5 rounded border border-sky-200">
                    {c.statusLabel}
                  </span>
                </div>
              </div>

              {/* Status summary line */}
              <p className="text-[12px] text-[var(--text-secondary)] line-clamp-1">
                {c.customerSafeStatusText}
              </p>

              {/* Countdown & details pill */}
              <div className="flex items-center justify-between pt-2 border-t border-[var(--surface-border-subtle)] text-[11px]">
                <div className="flex items-center gap-1.5 font-semibold text-sky-700">
                  <Clock className="w-3.5 h-3.5" />
                  <span>{c.timeRemainingText}</span>
                </div>
                <span className="text-sky-600 group-hover:underline flex items-center gap-1 font-medium">
                  <span>Track lifecycle</span>
                  <ArrowRight className="w-3 h-3" />
                </span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
