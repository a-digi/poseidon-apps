import { useRef, useState } from 'react';
import { X } from 'lucide-react';
import { SaveButton, CancelButton, AddIconButton, ChangeIconButton, RemoveIconButton } from '../../common/Buttons';
import { Dropdown } from '../../common/Dropdown';
import type { Business, Tag } from '../../api';
import { BusinessLogo } from './BusinessLogo';

export type LogoAction =
  | { type: 'noop' }
  | { type: 'upload'; dataUrl: string }
  | { type: 'delete' };

interface BusinessFormProps {
  initial: Business | null;
  tags: Tag[];
  onSave: (
    payload: Omit<Business, 'id' | 'createdAt' | 'updatedAt' | 'logoType' | 'logo'>,
    logoAction: LogoAction,
  ) => Promise<void>;
  onClose: () => void;
}

interface FormState {
  name: string;
  taxId: string;
  email: string;
  address: string;
  notes: string;
  logoDataUrl: string;
  logoTouched: boolean;
  tagIds: string[];
}

const input = 'w-full px-3 py-2 rounded-md border border-slate-200 bg-white text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-500';
const inputErr = 'w-full px-3 py-2 rounded-md border border-red-500 bg-white text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-red-500';

function init(b: Business | null): FormState {
  return {
    name: b?.name ?? '',
    taxId: b?.taxId ?? '',
    email: b?.email ?? '',
    address: b?.address ?? '',
    notes: b?.notes ?? '',
    logoDataUrl: b?.logo ?? '',
    logoTouched: false,
    tagIds: b?.tagIds ?? [],
  };
}

export function BusinessForm({ initial, tags, onSave, onClose }: BusinessFormProps) {
  const [form, setForm] = useState<FormState>(() => init(initial));
  const [nameInvalid, setNameInvalid] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [logoError, setLogoError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const set = (k: keyof Pick<FormState, 'name' | 'taxId' | 'email' | 'address' | 'notes'>) =>
    (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      setForm(prev => ({ ...prev, [k]: e.target.value }));
      if (k === 'name') setNameInvalid(false);
    };

  const onPickLogo = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    if (file.size > 1024 * 1024) {
      setLogoError('Logo must be ≤ 1 MB');
      e.target.value = '';
      return;
    }
    const reader = new FileReader();
    reader.onload = () => {
      setForm(prev => ({ ...prev, logoDataUrl: String(reader.result ?? ''), logoTouched: true }));
      setLogoError(null);
    };
    reader.readAsDataURL(file);
  };

  const onRemoveLogo = () => {
    setForm(prev => ({ ...prev, logoDataUrl: '', logoTouched: true }));
    setLogoError(null);
    if (fileInputRef.current) fileInputRef.current.value = '';
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.name.trim()) { setNameInvalid(true); return; }
    setBusy(true);
    setError(null);
    const payload = {
      name: form.name.trim(),
      taxId: form.taxId,
      email: form.email,
      address: form.address,
      notes: form.notes,
      tagIds: form.tagIds,
    };
    const logoAction: LogoAction = !form.logoTouched
      ? { type: 'noop' }
      : form.logoDataUrl
        ? { type: 'upload', dataUrl: form.logoDataUrl }
        : { type: 'delete' };
    try {
      await onSave(payload, logoAction);
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
          <h2 className="text-lg font-semibold text-slate-900">{initial ? 'Edit business' : 'Add business'}</h2>
          <button type="button" onClick={onClose} aria-label="Close" className="text-slate-400 hover:text-slate-700 transition-colors">
            <X className="w-5 h-5" />
          </button>
        </div>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <div className="flex items-center gap-3">
              <BusinessLogo name={form.name || '?'} logo={form.logoDataUrl || undefined} size={64} />
              <div className="flex items-center gap-2">
                {form.logoDataUrl ? (
                  <>
                    <ChangeIconButton
                      aria-label="Change logo"
                      onClick={() => fileInputRef.current?.click()}
                      disabled={busy}
                    />
                    <RemoveIconButton
                      aria-label="Remove logo"
                      onClick={onRemoveLogo}
                      disabled={busy}
                    />
                  </>
                ) : (
                  <AddIconButton
                    aria-label="Add logo"
                    onClick={() => fileInputRef.current?.click()}
                    disabled={busy}
                  />
                )}
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/png,image/jpeg,image/webp,image/gif"
                  className="hidden"
                  onChange={onPickLogo}
                />
              </div>
            </div>
            {logoError && <span className="text-xs text-red-600">{logoError}</span>}
          </div>
          <label className="flex flex-col gap-1 text-sm font-medium text-slate-700">
            <span>Name <span className="text-red-500">*</span></span>
            <input type="text" value={form.name} onChange={set('name')} className={nameInvalid ? inputErr : input} />
            {nameInvalid && <span className="text-xs text-red-600">Name is required</span>}
          </label>
          <label className="flex flex-col gap-1 text-sm font-medium text-slate-700">
            Tax ID
            <input type="text" value={form.taxId} onChange={set('taxId')} className={input} />
          </label>
          <label className="flex flex-col gap-1 text-sm font-medium text-slate-700">
            Email
            <input type="text" value={form.email} onChange={set('email')} className={input} />
          </label>
          <label className="flex flex-col gap-1 text-sm font-medium text-slate-700">
            Address
            <textarea rows={2} value={form.address} onChange={set('address')} className={input} />
          </label>
          <label className="flex flex-col gap-1 text-sm font-medium text-slate-700">
            Notes
            <textarea rows={2} value={form.notes} onChange={set('notes')} className={input} />
          </label>
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
