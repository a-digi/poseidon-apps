import { useState } from 'react';
import { X } from 'lucide-react';
import { SaveButton, CancelButton } from '../../common/Buttons';
import { Dropdown } from '../../common/Dropdown';
import { CURRENCY_OPTIONS } from '../../lib/currencies';
import type { Business, Tag, UpcomingInvoice, User } from '../../api';
import { dateInputFromEpoch, epochFromDateInput } from '../../lib/format';

interface UpcomingFormProps {
  initial: UpcomingInvoice | null;
  businesses: Business[];
  users: User[];
  tags: Tag[];
  onSave: (payload: Omit<UpcomingInvoice, 'id' | 'createdAt' | 'updatedAt'>) => Promise<void>;
  onClose: () => void;
}

interface FormState {
  businessId: string;
  amount: string;
  currency: string;
  description: string;
  dueAt: string;
  userIds: string[];
  tagIds: string[];
}

interface Invalid {
  businessId?: boolean;
  amount?: boolean;
  dueAt?: boolean;
  description?: boolean;
}

const input = 'w-full px-3 py-2 rounded-md border border-slate-200 bg-white text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-500';
const inputErr = 'w-full px-3 py-2 rounded-md border border-red-500 bg-white text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-red-500';

function init(u: UpcomingInvoice | null): FormState {
  return {
    businessId: u?.businessId ?? '',
    amount: u != null ? String(u.amount) : '',
    currency: u?.currency ?? 'EUR',
    description: u?.description ?? '',
    dueAt: u ? dateInputFromEpoch(u.dueAt) : '',
    userIds: u?.userIds ?? [],
    tagIds: u?.tagIds ?? [],
  };
}

export function UpcomingForm({ initial, businesses, users, tags, onSave, onClose }: UpcomingFormProps) {
  const [form, setForm] = useState<FormState>(() => init(initial));
  const [invalid, setInvalid] = useState<Invalid>({});
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

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

  const set = (k: keyof FormState) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    setForm(prev => ({ ...prev, [k]: e.target.value }));
    setInvalid(prev => { const next = { ...prev }; delete next[k as keyof Invalid]; return next; });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const iv: Invalid = {};
    if (form.businessId === '' && form.description.trim() === '') iv.description = true;
    const amt = parseFloat(form.amount);
    if (!form.amount || !isFinite(amt)) iv.amount = true;
    if (!form.dueAt) iv.dueAt = true;
    if (Object.keys(iv).length) { setInvalid(iv); return; }
    setBusy(true);
    setError(null);
    try {
      await onSave({
        businessId: form.businessId,
        amount: amt,
        currency: form.currency || 'EUR',
        description: form.description,
        dueAt: epochFromDateInput(form.dueAt),
        userIds: form.userIds,
        tagIds: form.tagIds,
      });
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div className="bg-white rounded-xl shadow-xl p-6 max-w-lg w-full max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-slate-900">{initial ? 'Edit upcoming invoice' : 'Add upcoming invoice'}</h2>
          <button type="button" onClick={onClose} aria-label="Close" className="text-slate-400 hover:text-slate-700 transition-colors">
            <X className="w-5 h-5" />
          </button>
        </div>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <span className="text-sm font-medium text-slate-700">Business</span>
            <Dropdown
              mode="single"
              options={businesses.map(b => ({ value: b.id, label: b.name }))}
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
                onChange={v => setForm(prev => ({ ...prev, currency: v }))}
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
              onChange={ids => setForm(prev => ({ ...prev, userIds: ids }))}
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
          <label className="flex flex-col gap-1 text-sm font-medium text-slate-700">
            <span>Due at <span className="text-red-500">*</span></span>
            <input type="date" value={form.dueAt} onChange={set('dueAt')} className={invalid.dueAt ? inputErr : input} />
          </label>
          {error && <div className="text-sm text-red-700 bg-red-50 border border-red-200 rounded-md px-3 py-2">{error}</div>}
          <div className="flex justify-end gap-2 pt-2">
            <CancelButton onClick={onClose} disabled={busy} />
            <SaveButton loading={busy} />
          </div>
        </form>
      </div>
    </div>
  );
}
