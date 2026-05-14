import { useState } from 'react';
import { X } from 'lucide-react';
import { SaveButton, CancelButton } from '../../common/Buttons';
import { Dropdown } from '../../common/Dropdown';
import { CURRENCY_OPTIONS } from '../../lib/currencies';
import type { Business, Frequency, RecurringInvoice, Tag, User } from '../../api';
import { dateInputFromEpoch, epochFromDateInput } from '../../lib/format';

interface RecurringFormProps {
  initial: RecurringInvoice | null;
  businesses: Business[];
  users: User[];
  tags: Tag[];
  onSave: (payload: Omit<RecurringInvoice, 'id' | 'createdAt' | 'updatedAt'>) => Promise<void>;
  onClose: () => void;
}

interface FormState {
  businessId: string;
  amount: string;
  currency: string;
  description: string;
  frequency: Frequency | '';
  startAt: string;
  endAt: string;
  active: boolean;
  issueDayOfWeek: string;
  issueDayOfMonth: string;
  issueMonthOfYear: string;
  anchorTouched: boolean;
  userIds: string[];
  tagIds: string[];
}

interface Invalid {
  businessId?: boolean;
  amount?: boolean;
  frequency?: boolean;
  startAt?: boolean;
  issueDayOfWeek?: boolean;
  issueDayOfMonth?: boolean;
  issueMonthOfYear?: boolean;
  description?: boolean;
}

const input = 'w-full px-3 py-2 rounded-md border border-slate-200 bg-white text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-500';
const inputErr = 'w-full px-3 py-2 rounded-md border border-red-500 bg-white text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-red-500';

const FREQUENCY_OPTIONS = [
  { value: 'daily',   label: 'Daily' },
  { value: 'weekly',  label: 'Weekly' },
  { value: 'monthly', label: 'Monthly' },
  { value: 'yearly',  label: 'Yearly' },
];

const WEEKDAY_OPTIONS = [
  { value: '1', label: 'Monday' },
  { value: '2', label: 'Tuesday' },
  { value: '3', label: 'Wednesday' },
  { value: '4', label: 'Thursday' },
  { value: '5', label: 'Friday' },
  { value: '6', label: 'Saturday' },
  { value: '7', label: 'Sunday' },
];

const MONTH_OPTIONS = [
  { value: '1', label: 'January' },
  { value: '2', label: 'February' },
  { value: '3', label: 'March' },
  { value: '4', label: 'April' },
  { value: '5', label: 'May' },
  { value: '6', label: 'June' },
  { value: '7', label: 'July' },
  { value: '8', label: 'August' },
  { value: '9', label: 'September' },
  { value: '10', label: 'October' },
  { value: '11', label: 'November' },
  { value: '12', label: 'December' },
];

function anchorDefaults(startDate: string): { dow: number; dom: number; moy: number } {
  if (!startDate) return { dow: 1, dom: 1, moy: 1 };
  const d = new Date(`${startDate}T00:00:00Z`);
  if (Number.isNaN(d.getTime())) return { dow: 1, dom: 1, moy: 1 };
  const dow = ((d.getUTCDay() + 6) % 7) + 1; // Sun=0..Sat=6 → Mon=1..Sun=7 (ISO)
  return { dow, dom: d.getUTCDate(), moy: d.getUTCMonth() + 1 };
}

function init(r: RecurringInvoice | null): FormState {
  const startInput = r ? dateInputFromEpoch(r.startAt) : '';
  const defaults = anchorDefaults(startInput);
  const dow = r && r.issueDayOfWeek ? r.issueDayOfWeek : defaults.dow;
  const dom = r && r.issueDayOfMonth ? r.issueDayOfMonth : defaults.dom;
  const moy = r && r.issueMonthOfYear ? r.issueMonthOfYear : defaults.moy;
  return {
    businessId: r?.businessId ?? '',
    amount: r != null ? String(r.amount) : '',
    currency: r?.currency ?? 'EUR',
    description: r?.description ?? '',
    frequency: r?.frequency ?? '',
    startAt: startInput,
    endAt: r?.endAt ? dateInputFromEpoch(r.endAt) : '',
    active: r?.active ?? true,
    issueDayOfWeek: String(dow),
    issueDayOfMonth: String(dom),
    issueMonthOfYear: String(moy),
    anchorTouched: false,
    userIds: r?.userIds ?? [],
    tagIds: r?.tagIds ?? [],
  };
}

function toPayload(form: FormState, amount: number, endAt: number | undefined): Omit<RecurringInvoice, 'id' | 'createdAt' | 'updatedAt'> {
  const payload: Omit<RecurringInvoice, 'id' | 'createdAt' | 'updatedAt'> = {
    businessId: form.businessId,
    amount,
    currency: form.currency || 'EUR',
    description: form.description,
    frequency: form.frequency as Frequency,
    startAt: epochFromDateInput(form.startAt),
    endAt,
    active: form.active,
    userIds: form.userIds,
    tagIds: form.tagIds,
  };
  if (form.frequency === 'weekly') {
    payload.issueDayOfWeek = parseInt(form.issueDayOfWeek, 10);
  } else if (form.frequency === 'monthly') {
    payload.issueDayOfMonth = parseInt(form.issueDayOfMonth, 10);
  } else if (form.frequency === 'yearly') {
    payload.issueMonthOfYear = parseInt(form.issueMonthOfYear, 10);
    payload.issueDayOfMonth = parseInt(form.issueDayOfMonth, 10);
  }
  return payload;
}

export function RecurringForm({ initial, businesses, users, tags, onSave, onClose }: RecurringFormProps) {
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

  const setField = (k: keyof FormState) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    const rawValue = e.target.type === 'checkbox' ? (e.target as HTMLInputElement).checked : e.target.value;
    setForm(prev => {
      const next: FormState = { ...prev, [k]: rawValue } as FormState;
      if ((k === 'frequency' || k === 'startAt') && !prev.anchorTouched) {
        const baseStart = k === 'startAt' ? String(rawValue) : prev.startAt;
        const defaults = anchorDefaults(baseStart);
        next.issueDayOfWeek = String(defaults.dow);
        next.issueDayOfMonth = String(defaults.dom);
        next.issueMonthOfYear = String(defaults.moy);
      }
      return next;
    });
    setInvalid(prev => { const next = { ...prev }; delete next[k as keyof Invalid]; return next; });
  };

  const setFrequency = (v: string) => {
    setForm(prev => {
      const next: FormState = { ...prev, frequency: v as Frequency | '' };
      if (!prev.anchorTouched) {
        const defaults = anchorDefaults(prev.startAt);
        next.issueDayOfWeek   = String(defaults.dow);
        next.issueDayOfMonth  = String(defaults.dom);
        next.issueMonthOfYear = String(defaults.moy);
      }
      return next;
    });
    setInvalid(prev => { const next = { ...prev }; delete next.frequency; return next; });
  };

  const setAnchor = (k: 'issueDayOfWeek' | 'issueDayOfMonth' | 'issueMonthOfYear', v: string) => {
    setForm(prev => ({ ...prev, [k]: v, anchorTouched: true }));
    setInvalid(prev => { const next = { ...prev }; delete next[k]; return next; });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const iv: Invalid = {};
    if (form.businessId === '' && form.description.trim() === '') iv.description = true;
    const amt = parseFloat(form.amount);
    if (!form.amount || !isFinite(amt)) iv.amount = true;
    if (!form.frequency) iv.frequency = true;
    if (!form.startAt) iv.startAt = true;
    if (form.frequency === 'weekly') {
      const dow = parseInt(form.issueDayOfWeek, 10);
      if (!Number.isFinite(dow) || dow < 1 || dow > 7) iv.issueDayOfWeek = true;
    } else if (form.frequency === 'monthly') {
      const dom = parseInt(form.issueDayOfMonth, 10);
      if (!Number.isFinite(dom) || dom < 1 || dom > 31) iv.issueDayOfMonth = true;
    } else if (form.frequency === 'yearly') {
      const moy = parseInt(form.issueMonthOfYear, 10);
      if (!Number.isFinite(moy) || moy < 1 || moy > 12) iv.issueMonthOfYear = true;
      const dom = parseInt(form.issueDayOfMonth, 10);
      if (!Number.isFinite(dom) || dom < 1 || dom > 31) iv.issueDayOfMonth = true;
    }
    if (Object.keys(iv).length) { setInvalid(iv); return; }
    setBusy(true);
    setError(null);
    try {
      const endAt = epochFromDateInput(form.endAt) || undefined;
      await onSave(toPayload(form, amt, endAt));
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
          <h2 className="text-lg font-semibold text-slate-900">{initial ? 'Edit recurring invoice' : 'Add recurring invoice'}</h2>
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
              <input type="number" step="0.01" placeholder="0.00" value={form.amount} onChange={setField('amount')} className={invalid.amount ? inputErr : input} />
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
              onChange={setField('description')}
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
          <div className="flex flex-col gap-1">
            <span className="text-sm font-medium text-slate-700">
              Frequency <span className="text-red-500">*</span>
            </span>
            <div className={invalid.frequency ? 'ring-1 ring-red-500 rounded-md' : ''}>
              <Dropdown
                mode="single"
                options={FREQUENCY_OPTIONS}
                value={form.frequency}
                onChange={setFrequency}
                placeholder="Select frequency"
              />
            </div>
          </div>
          {form.frequency === 'weekly' && (
            <div className="flex flex-col gap-1">
              <span className="text-sm font-medium text-slate-700">
                Day of week <span className="text-red-500">*</span>
              </span>
              <div className={invalid.issueDayOfWeek ? 'ring-1 ring-red-500 rounded-md' : ''}>
                <Dropdown
                  mode="single"
                  options={WEEKDAY_OPTIONS}
                  value={form.issueDayOfWeek}
                  onChange={v => setAnchor('issueDayOfWeek', v)}
                  placeholder="Select day"
                />
              </div>
            </div>
          )}
          {form.frequency === 'monthly' && (
            <label className="flex flex-col gap-1 text-sm font-medium text-slate-700">
              <span>Day of month <span className="text-red-500">*</span></span>
              <input
                type="number"
                min={1}
                max={31}
                step={1}
                value={form.issueDayOfMonth}
                onChange={e => setAnchor('issueDayOfMonth', e.target.value)}
                className={invalid.issueDayOfMonth ? inputErr : input}
              />
            </label>
          )}
          {form.frequency === 'yearly' && (
            <div className="grid grid-cols-2 gap-3">
              <div className="flex flex-col gap-1">
                <span className="text-sm font-medium text-slate-700">
                  Month <span className="text-red-500">*</span>
                </span>
                <div className={invalid.issueMonthOfYear ? 'ring-1 ring-red-500 rounded-md' : ''}>
                  <Dropdown
                    mode="single"
                    options={MONTH_OPTIONS}
                    value={form.issueMonthOfYear}
                    onChange={v => setAnchor('issueMonthOfYear', v)}
                    placeholder="Select month"
                  />
                </div>
              </div>
              <label className="flex flex-col gap-1 text-sm font-medium text-slate-700">
                <span>Day <span className="text-red-500">*</span></span>
                <input
                  type="number"
                  min={1}
                  max={31}
                  step={1}
                  value={form.issueDayOfMonth}
                  onChange={e => setAnchor('issueDayOfMonth', e.target.value)}
                  className={invalid.issueDayOfMonth ? inputErr : input}
                />
              </label>
            </div>
          )}
          <div className="grid grid-cols-2 gap-3">
            <label className="flex flex-col gap-1 text-sm font-medium text-slate-700">
              <span>Start at <span className="text-red-500">*</span></span>
              <input type="date" value={form.startAt} onChange={setField('startAt')} className={invalid.startAt ? inputErr : input} />
            </label>
            <label className="flex flex-col gap-1 text-sm font-medium text-slate-700">
              End at
              <input type="date" value={form.endAt} onChange={setField('endAt')} className={input} />
            </label>
          </div>
          <label className="flex items-center gap-2 text-sm font-medium text-slate-700 cursor-pointer">
            <input type="checkbox" checked={form.active} onChange={setField('active')} className="w-4 h-4" />
            Active
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
