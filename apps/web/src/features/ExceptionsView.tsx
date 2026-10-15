import React, { useState } from 'react';
import { Sparkles, ShieldAlert } from 'lucide-react';
import { Button } from '../components/Button';
import { StatusBadge } from '../components/StatusBadge';
import { RiskConfidenceIndicator } from '../components/RiskConfidenceIndicator';
import { useApp } from '../context/AppContext';
import { INITIAL_EXCEPTIONS } from '../services/mockSimulatorData';
import type { FinancialException } from '../types/domain';

export const ExceptionsView: React.FC = () => {
  const { setActiveView, setSelectedTransactionId, addToast } = useApp();
  const [exceptions] = useState<FinancialException[]>(() => INITIAL_EXCEPTIONS);

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-4 p-4 bg-[var(--surface)] border border-[var(--surface-border)] rounded-md shadow-xs">
        <div>
          <h2 className="text-[16px] font-bold text-[var(--text-primary)]">
            Detected Financial Exceptions & Anomaly Clusters
          </h2>
          <p className="text-[13px] text-[var(--text-muted)] mt-0.5">
            Automated exception classification comparing switch settlement reports and debit legs.
          </p>
        </div>
      </div>

      <div className="border border-[var(--surface-border)] rounded-md bg-[var(--surface)] shadow-xs divide-y divide-[var(--surface-border)]">
        <div className="px-4 py-3 bg-[var(--canvas-subdued)] font-semibold text-[13px] text-[var(--text-secondary)] flex justify-between">
          <span>Active Flagged Exceptions ({exceptions.length})</span>
          <span className="font-mono-financial text-[11px] text-[var(--text-muted)]">
            Classification Engine: XGBoost + Heuristic Rule Tree
          </span>
        </div>

        <div className="divide-y divide-[var(--surface-border-subtle)]">
          {exceptions.map((exc) => {
            return (
              <div
                key={exc.id}
                className="p-4 hover:bg-[var(--surface-raised)] transition-colors flex flex-wrap items-center justify-between gap-4 text-[13px]"
              >
                <div className="space-y-1.5 max-w-2xl">
                  <div className="flex flex-wrap items-center gap-2">
                    <span
                      className={`text-[11px] font-bold px-2 py-0.5 rounded font-mono-financial border ${
                        exc.severity === 'CRITICAL'
                          ? 'bg-rose-100 text-rose-900 border-rose-300'
                          : exc.severity === 'HIGH'
                          ? 'bg-amber-100 text-amber-900 border-amber-300'
                          : 'bg-slate-100 text-slate-800 border-slate-300'
                      }`}
                    >
                      {exc.severity}
                    </span>
                    <strong className="font-mono-financial text-[14px] text-[var(--text-primary)]">
                      {exc.reference}
                    </strong>
                    <span className="text-[11px] font-mono-financial px-1.5 py-0.2 rounded bg-[var(--canvas-subdued)] border border-[var(--surface-border)]">
                      {exc.channel}
                    </span>
                    <StatusBadge status={exc.status} size="sm" />
                    <RiskConfidenceIndicator modelConfidence={exc.modelConfidence} />
                  </div>

                  <p className="text-[13px] text-[var(--text-primary)] leading-relaxed">
                    {exc.description}
                  </p>

                  <div className="p-2 rounded bg-sky-50/50 border border-sky-200 text-sky-900 text-[12px] flex items-start gap-1.5">
                    <Sparkles className="w-3.5 h-3.5 text-sky-700 shrink-0 mt-0.5" />
                    <span>{exc.modelExplanation}</span>
                  </div>
                </div>

                <div className="text-right flex flex-col items-end gap-2">
                  <div className="font-bold font-mono-financial text-[18px] text-[var(--text-primary)]">
                    {exc.currency} {Number(exc.amount).toLocaleString('en-US', { minimumFractionDigits: 2 })}
                  </div>

                  <div className="flex items-center gap-2">
                    <Button
                      size="sm"
                      variant="secondary"
                      onClick={() => {
                        setSelectedTransactionId(exc.transactionId);
                        setActiveView('transactions');
                      }}
                    >
                      Inspect Ledger
                    </Button>
                    <Button
                      size="sm"
                      variant="primary"
                      leftIcon={<ShieldAlert className="w-3.5 h-3.5" />}
                      onClick={() => {
                        setActiveView('reversals');
                        addToast({
                          type: 'info',
                          title: 'Drafting Reversal',
                          message: `Pre-populating reversal proposal for ${exc.reference}`,
                        });
                      }}
                    >
                      Propose Reversal
                    </Button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};
