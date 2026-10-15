import {
  FileText,
  UploadCloud,
  Clock,
  ShieldCheck,
  CheckCircle2,
} from 'lucide-react';
import type { IntakeMode } from './reverseitTypes';

export interface S1ReportProblemProps {
  onSelectMode: (mode: IntakeMode) => void;
  onViewMyCases: () => void;
  activeCaseCount: number;
}

export const S1ReportProblem: React.FC<S1ReportProblemProps> = ({
  onSelectMode,
  onViewMyCases,
  activeCaseCount,
}) => {
  return (
    <div className="max-w-2xl mx-auto py-6 px-4 space-y-8 animate-in fade-in duration-200">
      {/* Primary Headline */}
      <div className="text-center space-y-2">
        <h2 className="text-[26px] font-extrabold text-[var(--text-primary)] tracking-tight">
          Money left your account but something went wrong?
        </h2>
        <p className="text-[14px] text-[var(--text-secondary)] max-w-lg mx-auto leading-relaxed">
          Report failed transfers, POS timeouts, or duplicate debits in seconds. 
          We parse your alert, assign the official Central Bank of Nigeria (CBN) resolution clock, 
          and track your money back to your account.
        </p>
      </div>

      {/* Two Big Action Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Choice 1: Paste Alert Text */}
        <button
          type="button"
          onClick={() => onSelectMode('PASTE_TEXT')}
          className="p-6 bg-[var(--surface)] border-2 border-[var(--surface-border)] hover:border-sky-500 rounded-xl shadow-xs hover:shadow-md transition-all text-left flex flex-col justify-between space-y-4 group cursor-pointer focus:outline-none focus:ring-2 focus:ring-sky-500"
        >
          <div className="p-3 w-12 h-12 rounded-lg bg-sky-50 text-sky-700 flex items-center justify-center group-hover:scale-105 transition-transform">
            <FileText className="w-6 h-6" />
          </div>
          <div>
            <h3 className="text-[16px] font-bold text-[var(--text-primary)] group-hover:text-sky-700">
              Paste Debit Alert
            </h3>
            <p className="text-[12px] text-[var(--text-muted)] mt-1 leading-relaxed">
              Copy and paste the SMS or email notification received from your bank.
            </p>
          </div>
          <span className="text-[12px] font-semibold text-sky-600 group-hover:underline flex items-center gap-1">
            <span>Continue with text</span>
            <span>&rarr;</span>
          </span>
        </button>

        {/* Choice 2: Upload Screenshot */}
        <button
          type="button"
          onClick={() => onSelectMode('UPLOAD_SCREENSHOT')}
          className="p-6 bg-[var(--surface)] border-2 border-[var(--surface-border)] hover:border-purple-500 rounded-xl shadow-xs hover:shadow-md transition-all text-left flex flex-col justify-between space-y-4 group cursor-pointer focus:outline-none focus:ring-2 focus:ring-purple-500"
        >
          <div className="p-3 w-12 h-12 rounded-lg bg-purple-50 text-purple-700 flex items-center justify-center group-hover:scale-105 transition-transform">
            <UploadCloud className="w-6 h-6" />
          </div>
          <div>
            <h3 className="text-[16px] font-bold text-[var(--text-primary)] group-hover:text-purple-700">
              Upload Screenshot
            </h3>
            <p className="text-[12px] text-[var(--text-muted)] mt-1 leading-relaxed">
              Upload a screenshot of your bank alert or mobile app debit receipt.
            </p>
          </div>
          <span className="text-[12px] font-semibold text-purple-600 group-hover:underline flex items-center gap-1">
            <span>Continue with image</span>
            <span>&rarr;</span>
          </span>
        </button>
      </div>

      {/* Trust & Regulatory Highlights */}
      <div className="p-4 rounded-lg bg-[var(--canvas-subdued)] border border-[var(--surface-border)] grid grid-cols-1 sm:grid-cols-3 gap-3 text-center text-[12px]">
        <div className="flex items-center justify-center gap-1.5 font-medium text-[var(--text-secondary)]">
          <ShieldCheck className="w-4 h-4 text-emerald-600" />
          <span>Local PII Masking</span>
        </div>
        <div className="flex items-center justify-center gap-1.5 font-medium text-[var(--text-secondary)]">
          <Clock className="w-4 h-4 text-sky-600" />
          <span>CBN 24h / 72h Clocks</span>
        </div>
        <div className="flex items-center justify-center gap-1.5 font-medium text-[var(--text-secondary)]">
          <CheckCircle2 className="w-4 h-4 text-purple-600" />
          <span>Verified Money Restoration</span>
        </div>
      </div>

      {/* Link to Existing Cases */}
      {activeCaseCount > 0 && (
        <div className="text-center pt-2">
          <button
            type="button"
            onClick={onViewMyCases}
            className="text-[13px] text-[var(--text-secondary)] hover:text-sky-600 underline font-medium cursor-pointer"
          >
            Looking for a previous report? View My Cases ({activeCaseCount})
          </button>
        </div>
      )}
    </div>
  );
};
