import React, { useState } from 'react';
import { Sparkles, ShieldAlert } from 'lucide-react';
import { Button } from '../components/Button';
import { EvidencePanel } from '../components/EvidencePanel';
import { StatusBadge } from '../components/StatusBadge';
import { RiskConfidenceIndicator } from '../components/RiskConfidenceIndicator';
import { useApp } from '../context/AppContext';
import { INITIAL_INVESTIGATIONS } from '../services/mockSimulatorData';
import type { InvestigationCase } from '../types/domain';

export const InvestigationsView: React.FC = () => {
  const { setActiveView, addToast } = useApp();
  const [cases] = useState<InvestigationCase[]>(() => INITIAL_INVESTIGATIONS);
  const [selectedCase, setSelectedCase] = useState<InvestigationCase | null>(() => INITIAL_INVESTIGATIONS[0] || null);

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-4 p-4 bg-[var(--surface)] border border-[var(--surface-border)] rounded-md shadow-xs">
        <div>
          <h2 className="text-[16px] font-bold text-[var(--text-primary)]">
            Exception Investigation & Root-Cause Analysis
          </h2>
          <p className="text-[13px] text-[var(--text-muted)] mt-0.5">
            Dossier investigation combining verified ledger facts, AI anomaly clustering, and counterparty switch logs.
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Cases List */}
        <div className="border border-[var(--surface-border)] rounded-md bg-[var(--surface)] shadow-xs divide-y divide-[var(--surface-border)]">
          <div className="px-4 py-3 bg-[var(--canvas-subdued)] font-semibold text-[13px] text-[var(--text-secondary)]">
            Active Cases ({cases.length})
          </div>

          <div className="divide-y divide-[var(--surface-border-subtle)]">
            {cases.map((c) => {
              const isSelected = selectedCase?.id === c.id;
              return (
                <div
                  key={c.id}
                  onClick={() => setSelectedCase(c)}
                  className={`p-3.5 cursor-pointer transition-colors space-y-1 text-[13px] ${
                    isSelected ? 'bg-[var(--primary-subtle)]/40' : 'hover:bg-[var(--surface-raised)]'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <strong className="font-mono-financial text-[var(--text-primary)]">
                      {c.caseNumber}
                    </strong>
                    <span className="text-[11px] font-semibold text-rose-700 bg-rose-50 px-1.5 py-0.2 rounded border border-rose-200">
                      {c.priority}
                    </span>
                  </div>

                  <div className="text-[12px] text-[var(--text-secondary)] font-mono-financial">
                    Ref: {c.reference}
                  </div>

                  <div className="flex items-center justify-between pt-1">
                    <StatusBadge status={c.status} size="sm" />
                    <RiskConfidenceIndicator modelConfidence={c.modelConfidence} />
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Evidence Panel & Remedy */}
        <div className="lg:col-span-2 space-y-4">
          {selectedCase && (
            <>
              {/* AI Recommended Remediation */}
              <div className="p-4 bg-sky-50/70 border border-sky-300 rounded-md text-sky-950 space-y-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2 font-bold text-[13px] text-sky-900">
                    <Sparkles className="w-4 h-4 text-sky-700" />
                    <span>Explainable AI Resolution Proposal</span>
                  </div>
                  <RiskConfidenceIndicator modelConfidence={selectedCase.modelConfidence} />
                </div>
                <p className="text-[13px] text-sky-900 leading-relaxed">
                  {selectedCase.aiSuggestedRemedy}
                </p>
                <div className="pt-2 flex items-center justify-end">
                  <Button
                    size="sm"
                    variant="primary"
                    onClick={() => {
                      setActiveView('reversals');
                      addToast({
                        type: 'info',
                        title: 'Drafting Reversal',
                        message: `Loaded evidence into Maker-Checker draft for ${selectedCase.reference}`,
                      });
                    }}
                    leftIcon={<ShieldAlert className="w-4 h-4" />}
                  >
                    Draft Proposed Resolution
                  </Button>
                </div>
              </div>

              {/* 4-Quadrant Evidence Dossier (Fact, Inferred, Missing, Unresolved) */}
              <EvidencePanel
                evidence={selectedCase.evidence}
                caseNumber={selectedCase.caseNumber}
                reference={selectedCase.reference}
              />
            </>
          )}
        </div>
      </div>
    </div>
  );
};
