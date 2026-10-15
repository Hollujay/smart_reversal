import React, { useState } from 'react';
import {
  ArrowLeft,
  ArrowRight,
  Edit2,
  Building,
} from 'lucide-react';
import { Button } from '../../components/Button';
import type { ParsedAlertDetails } from './reverseitTypes';

export interface S3CheckDetailsProps {
  parsed: ParsedAlertDetails;
  onBack: () => void;
  onConfirmDetails: (confirmed: ParsedAlertDetails) => void;
}

export const S3CheckDetails: React.FC<S3CheckDetailsProps> = ({
  parsed,
  onBack,
  onConfirmDetails,
}) => {
  const [details, setDetails] = useState<ParsedAlertDetails>(parsed);
  const [isEditing, setIsEditing] = useState(false);
  const [isConfirmedByUser, setIsConfirmedByUser] = useState(true);

  const handleFieldChange = (field: keyof ParsedAlertDetails, value: string) => {
    setDetails((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  return (
    <div className="max-w-2xl mx-auto py-4 px-4 space-y-6 animate-in fade-in duration-200">
      <button
        type="button"
        onClick={onBack}
        className="inline-flex items-center gap-1.5 text-[13px] text-[var(--text-secondary)] hover:text-[var(--text-primary)] font-medium cursor-pointer"
      >
        <ArrowLeft className="w-4 h-4" />
        <span>Re-enter alert</span>
      </button>

      {/* Screen Title */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-[20px] font-bold text-[var(--text-primary)]">
            Step 2: Check Extracted Details
          </h2>
          <p className="text-[13px] text-[var(--text-muted)] mt-1">
            Please verify the numbers and references extracted from your alert before we check bank records.
          </p>
        </div>
        <Button
          size="xs"
          variant="secondary"
          onClick={() => setIsEditing(!isEditing)}
          leftIcon={<Edit2 className="w-3.5 h-3.5" />}
        >
          {isEditing ? 'Done Editing' : 'Edit Fields'}
        </Button>
      </div>

      {/* Non-Wema Alert Notice */}
      {details.isNonWemaAlert && (
        <div className="p-3 rounded-lg bg-blue-50 border border-blue-200 text-blue-900 text-[12px] flex items-start gap-2.5">
          <Building className="w-4 h-4 text-blue-700 shrink-0 mt-0.5" />
          <div>
            <strong className="font-semibold">Issuing Institution: {details.issuingBank}</strong>
            <p className="text-[11px] text-blue-800 mt-0.5">
              We parsed this from an external bank alert. ReverseIt will still compute your 
              official Central Bank of Nigeria (CBN) regulatory countdown and draft your formal complaint letter.
            </p>
          </div>
        </div>
      )}

      {/* Details Table / Form */}
      <div className="bg-[var(--surface)] border border-[var(--surface-border)] rounded-xl overflow-hidden shadow-xs divide-y divide-[var(--surface-border)] text-[13px]">
        {/* Amount */}
        <div className="p-4 flex items-center justify-between">
          <div>
            <span className="text-[11px] font-bold uppercase tracking-wider text-[var(--text-muted)] block">
              Debited Amount
            </span>
            {isEditing ? (
              <input
                type="text"
                value={details.amount}
                onChange={(e) => handleFieldChange('amount', e.target.value)}
                className="mt-1 px-2 py-1 border rounded text-[14px] font-mono-financial font-bold"
              />
            ) : (
              <span className="text-[18px] font-bold font-mono-financial text-[var(--text-primary)]">
                {details.currency} {details.amount}
              </span>
            )}
          </div>
          <span className="text-[11px] font-mono-financial px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-800 border border-emerald-200">
            {details.confidence.amount}% Confidence
          </span>
        </div>

        {/* Account Tail */}
        <div className="p-4 flex items-center justify-between">
          <div>
            <span className="text-[11px] font-bold uppercase tracking-wider text-[var(--text-muted)] block">
              Account Number
            </span>
            {isEditing ? (
              <input
                type="text"
                value={details.accountTail}
                onChange={(e) => handleFieldChange('accountTail', e.target.value)}
                className="mt-1 px-2 py-1 border rounded font-mono-financial"
              />
            ) : (
              <span className="font-mono-financial text-[14px] text-[var(--text-primary)]">
                •••• {details.accountTail}
              </span>
            )}
          </div>
          <span className="text-[11px] font-mono-financial px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-800 border border-emerald-200">
            {details.confidence.accountTail}% Masked
          </span>
        </div>

        {/* Date & Time */}
        <div className="p-4 flex items-center justify-between">
          <div>
            <span className="text-[11px] font-bold uppercase tracking-wider text-[var(--text-muted)] block">
              Transaction Date & Time
            </span>
            {isEditing ? (
              <input
                type="text"
                value={details.dateTime}
                onChange={(e) => handleFieldChange('dateTime', e.target.value)}
                className="mt-1 px-2 py-1 border rounded font-mono-financial w-64"
              />
            ) : (
              <span className="font-mono-financial text-[13px] text-[var(--text-primary)]">
                {details.dateTime}
              </span>
            )}
          </div>
          <span className="text-[11px] text-[var(--text-muted)]">West Africa Time</span>
        </div>

        {/* Channel & Reference */}
        <div className="p-4 flex items-center justify-between">
          <div>
            <span className="text-[11px] font-bold uppercase tracking-wider text-[var(--text-muted)] block">
              Channel & Reference (RRN / Session)
            </span>
            {isEditing ? (
              <input
                type="text"
                value={details.reference}
                onChange={(e) => handleFieldChange('reference', e.target.value)}
                className="mt-1 px-2 py-1 border rounded font-mono-financial w-64"
              />
            ) : (
              <div className="flex items-center gap-2 mt-0.5">
                <span className="px-2 py-0.5 rounded bg-sky-100 text-sky-800 font-bold text-[11px]">
                  {details.channel}
                </span>
                <span className="font-mono-financial text-[12px] text-[var(--text-secondary)]">
                  {details.reference}
                </span>
              </div>
            )}
          </div>
          <span className="text-[11px] font-mono-financial px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-800 border border-emerald-200">
            {details.confidence.reference}% Confidence
          </span>
        </div>

        {/* Narration */}
        <div className="p-4">
          <span className="text-[11px] font-bold uppercase tracking-wider text-[var(--text-muted)] block">
            Narration / Description
          </span>
          {isEditing ? (
            <textarea
              rows={2}
              value={details.narration}
              onChange={(e) => handleFieldChange('narration', e.target.value)}
              className="mt-1 w-full px-2 py-1 border rounded font-mono-financial text-[12px]"
            />
          ) : (
            <p className="font-mono-financial text-[12px] text-[var(--text-secondary)] mt-0.5 break-all">
              {details.narration}
            </p>
          )}
        </div>
      </div>

      {/* Confirmation Toggle */}
      <div className="p-4 bg-[var(--canvas-subdued)] border border-[var(--surface-border)] rounded-lg">
        <label className="flex items-center gap-3 cursor-pointer select-none">
          <input
            type="checkbox"
            checked={isConfirmedByUser}
            onChange={(e) => setIsConfirmedByUser(e.target.checked)}
            className="w-4 h-4 rounded text-sky-600 focus:ring-sky-500 border-slate-300 cursor-pointer"
          />
          <span className="text-[13px] font-medium text-[var(--text-primary)]">
            I confirm that these transaction details reflect the debit alert I received.
          </span>
        </label>
      </div>

      {/* Navigation Actions */}
      <div className="flex items-center justify-between pt-2">
        <Button variant="ghost" size="md" onClick={onBack}>
          Back
        </Button>
        <Button
          variant="primary"
          size="md"
          disabled={!isConfirmedByUser}
          onClick={() => onConfirmDetails(details)}
          rightIcon={<ArrowRight className="w-4 h-4" />}
        >
          Match Against Bank Records
        </Button>
      </div>
    </div>
  );
};
