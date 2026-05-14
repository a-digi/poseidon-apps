import { useState } from 'react';
import { X } from 'lucide-react';
import { SaveButton, CancelButton } from '../../common/Buttons';
import type { User } from '../../api';

interface UserFormProps {
  initial: User | null;
  onSave: (payload: Omit<User, 'id' | 'createdAt' | 'updatedAt'>) => Promise<void>;
  onClose: () => void;
}

interface FormState {
  name: string;
  email: string;
  notes: string;
}

const input = 'w-full px-3 py-2 rounded-md border border-slate-200 bg-white text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-500';
const inputErr = 'w-full px-3 py-2 rounded-md border border-red-500 bg-white text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-red-500';

function init(u: User | null): FormState {
  return {
    name: u?.name ?? '',
    email: u?.email ?? '',
    notes: u?.notes ?? '',
  };
}

export function UserForm({ initial, onSave, onClose }: UserFormProps) {
  const [form, setForm] = useState<FormState>(() => init(initial));
  const [nameInvalid, setNameInvalid] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const set = (k: keyof FormState) =>
    (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      setForm(prev => ({ ...prev, [k]: e.target.value }));
      if (k === 'name') setNameInvalid(false);
    };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.name.trim()) { setNameInvalid(true); return; }
    setBusy(true);
    setError(null);
    try {
      await onSave({
        name: form.name.trim(),
        email: form.email,
        notes: form.notes,
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
          <h2 className="text-lg font-semibold text-slate-900">{initial ? 'Edit user' : 'Add user'}</h2>
          <button type="button" onClick={onClose} aria-label="Close" className="text-slate-400 hover:text-slate-700 transition-colors">
            <X className="w-5 h-5" />
          </button>
        </div>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <label className="flex flex-col gap-1 text-sm font-medium text-slate-700">
            <span>Name <span className="text-red-500">*</span></span>
            <input type="text" value={form.name} onChange={set('name')} className={nameInvalid ? inputErr : input} />
            {nameInvalid && <span className="text-xs text-red-600">Name is required</span>}
          </label>
          <label className="flex flex-col gap-1 text-sm font-medium text-slate-700">
            Email
            <input type="text" value={form.email} onChange={set('email')} className={input} />
          </label>
          <label className="flex flex-col gap-1 text-sm font-medium text-slate-700">
            Notes
            <textarea rows={2} value={form.notes} onChange={set('notes')} className={input} />
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
