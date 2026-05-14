import { useState } from 'react';
import { X } from 'lucide-react';
import { SaveButton, CancelButton } from '../../common/Buttons';
import { Dropdown } from '../../common/Dropdown';
import { CURRENCY_OPTIONS } from '../../lib/currencies';
import type { Business, Invoice, Tag, User } from '../../api';
import { dateInputFromEpoch, epochFromDateInput } from '../../lib/format';
import { AttachmentDialog } from './AttachmentDialog';
import { InvoiceAttachments } from './InvoiceAttachments';
import { useIsMobile } from '../../lib/useIsMobile';

interface InvoiceFormProps {
  initial: Invoice | null;
  businesses: Business[];
  users: User[];
  tags: Tag[];
  onSave: (payload: Omit<Invoice, 'id' | 'createdAt' | 'updatedAt'>) => Promise<void>;
  onClose: () => void;
}

interface FormState {
  businessId: string;
  amount: string;
  currency: string;
  description: string;
  userIds: string[];
  tagIds: string[];
  issuedAt: string;
  dueAt: string;
  paid: boolean;
}

interface Invalid {
  businessId?: boolean;
  amount?: boolean;
  issuedAt?: boolean;
  dueAt?: boolean;
  description?: boolean;
}

const input = 'w-full px-3 py-2 rounded-md border border-slate-200 bg-white text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-500';
const inputErr = 'w-full px-3 py-2 rounded-md border border-red-500 bg-white text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-red-500';

function init(i: Invoice | null): FormState {
  return {
    businessId: i?.businessId ?? '',
    amount: i != null ? String(i.amount) : '',
    currency: i?.currency ?? 'EUR',
    description: i?.description ?? '',
    userIds: i?.userIds ?? [],
    tagIds: i?.tagIds ?? [],
    issuedAt: i ? dateInputFromEpoch(i.issuedAt) : '',
    dueAt: i ? dateInputFromEpoch(i.dueAt) : '',
    paid: i?.paid ?? true,
  };
}

export function InvoiceForm({ initial, businesses, users, tags, onSave, onClose }: InvoiceFormProps) {
  const [form, setForm] = useState<FormState>(() => init(initial));
  const [invalid, setInvalid] = useState<Invalid>({});
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [showAttachmentDialog, setShowAttachmentDialog] = useState(false);
  const [attachmentRefreshKey, setAttachmentRefreshKey] = useState(0);
  const isMobile = useIsMobile();

  const setField = <K extends keyof FormState>(k: K, v: FormState[K]) => {
    setForm(prev => ({ ...prev, [k]: v }));
    setInvalid(prev => { const next = { ...prev }; delete next[k as keyof Invalid]; return next; });
  };

  const set = (k: keyof Pick<FormState, 'amount' | 'description' | 'issuedAt' | 'dueAt'>) =>
    (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
      setField(k, e.target.value as FormState[typeof k]);

  const onBusinessChange = (newBusinessId: string) => {
    const newBiz = businesses.find(b => b.id === newBusinessId);
    const bizTags = newBiz?.tagIds ?? [];
    setForm(prev => ({
      ...prev,
      businessId: newBusinessId,
      tagIds: Array.from(new Set([...prev.tagIds, ...bizTags])),
    }));
    setInvalid(prev => { const next = { ...prev }; delete next.businessId; return next; });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const iv: Invalid = {};
    if (form.businessId === '' && form.description.trim() === '') iv.description = true;
    const amt = parseFloat(form.amount);
    if (!form.amount || !isFinite(amt)) iv.amount = true;
    if (!form.issuedAt) iv.issuedAt = true;
    if (Object.keys(iv).length) { setInvalid(iv); return; }
    setBusy(true);
    setError(null);
    try {
      await onSave({
        businessId: form.businessId,
        amount: amt,
        currency: form.currency || 'EUR',
        description: form.description,
        userIds: form.userIds,
        tagIds: form.tagIds,
        issuedAt: epochFromDateInput(form.issuedAt),
        dueAt: form.dueAt ? epochFromDateInput(form.dueAt) : 0,
        paid: form.paid,
      });
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  };

  const businessOptions = businesses.map(b => ({ value: b.id, label: b.name }));

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div className="bg-white rounded-xl shadow-xl p-6 max-w-lg w-full max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-slate-900">{initial ? 'Edit invoice' : 'Add invoice'}</h2>
          <button type="button" onClick={onClose} aria-label="Close" className="text-slate-400 hover:text-slate-700 transition-colors">
            <X className="w-5 h-5" />
          </button>
        </div>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <span className="text-sm font-medium text-slate-700">Business</span>
            <Dropdown
              mode="single"
              options={businessOptions}
              value={form.businessId}
              onChange={onBusinessChange}
              placeholder="Select a business"
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <label className="flex flex-col gap-1 text-sm font-medium text-slate-700">
              <span>Amount <span className="text-red-500">*</span></span>
              <input type="number" step="0.01" placeholder="0.00" value={form.amount} onChange={set('amount')} className={invalid.amount ? inputErr : input} />
            </label>
            <div className="flex flex-col gap-1">
              <span className="text-sm font-medium text-slate-700">Currency</span>
              <Dropdown
                mode="single"
                options={CURRENCY_OPTIONS}
                value={form.currency}
                onChange={v => setField('currency', v)}
                placeholder="Select currency"
              />
            </div>
          </div>
          <label className="flex flex-col gap-1 text-sm font-medium text-slate-700">
            <span>
              Description{form.businessId === '' && <> <span className="text-red-500">*</span></>}
            </span>
            {form.businessId === '' && (
              <span className="text-xs font-normal text-slate-500">Required when no business is selected.</span>
            )}
            <textarea
              rows={2}
              value={form.description}
              onChange={set('description')}
              className={invalid.description ? inputErr : input}
            />
          </label>
          <div className="flex flex-col gap-1">
            <span className="text-sm font-medium text-slate-700">Users</span>
            <Dropdown
              mode="multi"
              options={users.map(u => ({ value: u.id, label: u.name }))}
              value={form.userIds}
              onChange={ids => setField('userIds', ids)}
              placeholder="Select users (optional)"
            />
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-sm font-medium text-slate-700">Tags</span>
            <Dropdown
              mode="multi"
              options={tags.map(t => ({ value: t.id, label: t.name }))}
              value={form.tagIds}
              onChange={ids => setForm(prev => ({ ...prev, tagIds: ids }))}
              placeholder="Select tags (optional)"
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <label className="flex flex-col gap-1 text-sm font-medium text-slate-700">
              <span>Issued at <span className="text-red-500">*</span></span>
              <input type="date" value={form.issuedAt} onChange={set('issuedAt')} className={invalid.issuedAt ? inputErr : input} />
            </label>
            <label className="flex flex-col gap-1 text-sm font-medium text-slate-700">
              <span>Due at</span>
              <input type="date" value={form.dueAt} onChange={set('dueAt')} className={input} />
            </label>
          </div>
          <label className="flex items-center gap-2 text-sm font-medium text-slate-700 cursor-pointer">
            <input type="checkbox" checked={form.paid} onChange={e => setField('paid', e.target.checked)} className="w-4 h-4" />
            Paid
          </label>
          {error && <div className="text-sm text-red-700 bg-red-50 border border-red-200 rounded-md px-3 py-2">{error}</div>}
          {initial !== null && (
            <InvoiceAttachments invoiceId={initial.id} refreshKey={attachmentRefreshKey} />
          )}
          {!isMobile && (
            <button
              type="button"
              onClick={() => setShowAttachmentDialog(true)}
              className="w-full px-3 py-2 rounded-md border border-slate-300 bg-white text-sm text-slate-700 hover:bg-slate-50 transition-colors text-left"
            >
              Upload from phone
            </button>
          )}
          <div className="flex justify-end gap-2 pt-2">
            <CancelButton onClick={onClose} disabled={busy} />
            <SaveButton loading={busy} />
          </div>
        </form>
      </div>
      {showAttachmentDialog && (
        <AttachmentDialog invoiceId={initial?.id} onClose={() => { setShowAttachmentDialog(false); setAttachmentRefreshKey(k => k + 1); }} />
      )}
    </div>
  );
}
