import React, { useState } from 'react';
import { RefreshCw, FileSpreadsheet } from 'lucide-react';
import { Button } from '../components/Button';
import { useApp } from '../context/AppContext';
import { RECONCILIATION_RUNS } from '../services/mockSimulatorData';
import type { ReconciliationSummary } from '../types/domain';

export const ReconciliationView: React.FC = () => {
  const { addToast } = useApp();
  const [runs] = useState<ReconciliationSummary[]>(() => RECONCILIATION_RUNS);
  const [isRunning, setIsRunning] = useState(false);

  const handleTriggerRun = () => {
    setIsRunning(true);
    setTimeout(() => {
      setIsRunning(false);
      addToast({
        type: 'success',
        title: 'Reconciliation Cycle Completed',
        message: 'Reconciled 14,250 records from sample_settlement_nip.csv against core ledger.',
      });
    }, 1200);
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-4 p-4 bg-[var(--surface)] border border-[var(--surface-border)] rounded-md shadow-xs">
        <div>
          <h2 className="text-[16px] font-bold text-[var(--text-primary)]">
            Multi-Source Financial Reconciliation Engine
          </h2>
          <p className="text-[13px] text-[var(--text-muted)] mt-0.5">
            Two-way and three-way reconciliation matching core banking ledgers against switch settlement files.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="primary"
            size="md"
            isLoading={isRunning}
            leftIcon={<RefreshCw className="w-4 h-4" />}
            onClick={handleTriggerRun}
          >
            Run Intake Cycle
          </Button>
        </div>
      </div>

      {/* Reconciliation Channels Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {runs.map((run) => (
          <div
            key={run.channel}
            className="p-4 bg-[var(--surface)] border border-[var(--surface-border)] rounded-md shadow-xs space-y-3"
          >
            <div className="flex items-center justify-between">
              <span className="font-bold text-[14px] text-[var(--text-primary)]">
                {run.channel} Switch Channel
              </span>
              <span
                className={`text-[11px] font-semibold px-2 py-0.5 rounded font-mono-financial ${
                  run.status === 'COMPLETED'
                    ? 'bg-emerald-50 text-emerald-900 border border-emerald-300'
                    : 'bg-amber-50 text-amber-900 border border-amber-300'
                }`}
              >
                {run.status.replace(/_/g, ' ')}
              </span>
            </div>

            <div className="space-y-1">
              <div className="text-[11px] text-[var(--text-muted)]">Reconciliation Accuracy</div>
              <div className="text-[24px] font-bold font-mono-financial text-[var(--text-primary)]">
                {run.reconciliationRate}%
              </div>
            </div>

            <div className="border-t border-[var(--surface-border-subtle)] pt-2 text-[12px] space-y-1 font-mono-financial text-[var(--text-secondary)]">
              <div className="flex justify-between">
                <span className="text-[var(--text-muted)]">Total Transacted:</span>
                <span>{run.totalTransactions.toLocaleString()}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-[var(--text-muted)]">Matched Records:</span>
                <span className="text-emerald-700">{run.matchedCount.toLocaleString()}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-[var(--text-muted)]">Unmatched Discrepancies:</span>
                <span className="text-rose-700 font-bold">{run.unmatchedCount}</span>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Simulator Fixture Ingestion Card */}
      <div className="p-4 bg-[var(--surface)] border border-[var(--surface-border)] rounded-md shadow-xs space-y-3">
        <h3 className="text-[14px] font-semibold text-[var(--text-primary)] flex items-center gap-2">
          <FileSpreadsheet className="w-4 h-4 text-emerald-600" />
          <span>Active Simulator Settlement Fixtures</span>
        </h3>
        <p className="text-[12px] text-[var(--text-secondary)]">
          The simulator is automatically feeding transactions from local repository fixtures:
        </p>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-[12px] font-mono-financial">
          <div className="p-3 bg-[var(--canvas-subdued)] border border-[var(--surface-border)] rounded space-y-1">
            <div className="font-semibold text-[var(--text-primary)]">sample_settlement_nip.csv</div>
            <div className="text-[var(--text-muted)]">NIP Settlement sessions NIP9901 - NIP9905</div>
            <div className="text-emerald-700 font-medium">✓ Ingested & Synced with Core Ledger</div>
          </div>

          <div className="p-3 bg-[var(--canvas-subdued)] border border-[var(--surface-border)] rounded space-y-1">
            <div className="font-semibold text-[var(--text-primary)]">sample_pos_batch.json</div>
            <div className="text-[var(--text-muted)]">Terminal TERM-8801, STAN 004402, Response 91</div>
            <div className="text-amber-700 font-medium">⚠ Timeout Mismatch Exception Active</div>
          </div>
        </div>
      </div>
    </div>
  );
};
