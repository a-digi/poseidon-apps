import { Loader2, Pencil, Plus, RefreshCw, Trash2, X } from 'lucide-react';

const iconBtn =
  'relative overflow-hidden w-10 h-10 flex items-center justify-center rounded-full active:scale-90 transition-all duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:opacity-40 disabled:cursor-not-allowed';

type NativeBtn = Omit<React.ButtonHTMLAttributes<HTMLButtonElement>, 'className' | 'type'>;

interface SaveButtonProps extends NativeBtn {
  loading?: boolean;
}

export function SaveButton({ loading = false, disabled, children = 'Save', ...rest }: SaveButtonProps) {
  return (
    <button
      type="submit"
      disabled={disabled ?? loading}
      className="relative overflow-hidden flex items-center gap-2 px-4 py-2 rounded-md text-sm font-semibold bg-gradient-to-r from-slate-900 to-slate-700 text-white shadow-sm hover:shadow-md active:scale-95 disabled:opacity-40 disabled:cursor-not-allowed transition-all duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-900 focus-visible:ring-offset-2"
      {...rest}
    >
      <span aria-hidden={true} className="absolute -right-5 top-1/2 -translate-y-1/2 w-10 h-10 rounded-full bg-white/15 pointer-events-none" />
      <span className="relative flex items-center gap-2">
        {loading && <Loader2 className="w-4 h-4 animate-spin" />}
        {children}
      </span>
    </button>
  );
}

export function CancelButton({ children = 'Cancel', ...rest }: NativeBtn) {
  return (
    <button
      type="button"
      className="relative overflow-hidden px-4 py-2 rounded-md text-sm font-semibold bg-gradient-to-br from-slate-200 to-slate-300 text-slate-800 hover:from-slate-300 hover:to-slate-400 active:scale-95 disabled:opacity-40 disabled:cursor-not-allowed transition-all duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-400 focus-visible:ring-offset-2"
      {...rest}
    >
      <span aria-hidden={true} className="absolute -left-5 top-1/2 -translate-y-1/2 w-10 h-10 rounded-full bg-slate-900/10 pointer-events-none" />
      <span className="relative">{children}</span>
    </button>
  );
}

export function AddButton({ children = 'Add new', ...rest }: NativeBtn) {
  return (
    <button
      type="button"
      className="relative overflow-hidden flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold bg-gradient-to-br from-slate-700 via-slate-800 to-slate-950 text-white shadow-md hover:shadow-xl active:scale-95 transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-900 focus-visible:ring-offset-2"
      {...rest}
    >
      <span aria-hidden={true} className="absolute -right-6 top-1/2 -translate-y-1/2 w-12 h-12 rounded-full bg-white/20 pointer-events-none" />
      <span className="relative flex items-center gap-2">
        <Plus className="w-4 h-4" />
        {children}
      </span>
    </button>
  );
}

export function EditButton({ 'aria-label': ariaLabel = 'Edit', ...rest }: NativeBtn) {
  return (
    <button
      type="button"
      aria-label={ariaLabel}
      className="relative overflow-hidden w-8 h-8 flex items-center justify-center rounded-md text-slate-400 hover:text-slate-900 hover:bg-gradient-to-br hover:from-slate-100 hover:to-slate-300 active:scale-90 transition-all duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-400 focus-visible:ring-offset-1"
      {...rest}
    >
      <Pencil className="w-4 h-4 relative" />
    </button>
  );
}

export function DeleteButton({ 'aria-label': ariaLabel = 'Delete', ...rest }: NativeBtn) {
  return (
    <button
      type="button"
      aria-label={ariaLabel}
      className="relative overflow-hidden w-8 h-8 flex items-center justify-center rounded-md text-slate-400 hover:text-red-700 hover:bg-gradient-to-br hover:from-red-50 hover:to-red-200 active:scale-90 transition-all duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-red-400 focus-visible:ring-offset-1"
      {...rest}
    >
      <Trash2 className="w-4 h-4 relative" />
    </button>
  );
}

export function AddIconButton({ 'aria-label': ariaLabel = 'Add', ...rest }: NativeBtn) {
  return (
    <button
      type="button"
      aria-label={ariaLabel}
      className={`${iconBtn} bg-gradient-to-br from-slate-700 via-slate-800 to-slate-950 text-white shadow-md hover:shadow-xl focus-visible:ring-slate-900`}
      {...rest}
    >
      <span aria-hidden={true} className="absolute -right-4 top-1/2 -translate-y-1/2 w-8 h-8 rounded-full bg-white/20 pointer-events-none" />
      <Plus className="w-5 h-5 relative" />
    </button>
  );
}

export function EditIconButton({ 'aria-label': ariaLabel = 'Edit', ...rest }: NativeBtn) {
  return (
    <button
      type="button"
      aria-label={ariaLabel}
      className={`${iconBtn} bg-gradient-to-br from-slate-100 to-slate-200 text-slate-700 hover:from-slate-200 hover:to-slate-300 shadow-sm hover:shadow-md focus-visible:ring-slate-400`}
      {...rest}
    >
      <span aria-hidden={true} className="absolute -right-4 top-1/2 -translate-y-1/2 w-8 h-8 rounded-full bg-slate-900/5 pointer-events-none" />
      <Pencil className="w-4 h-4 relative" />
    </button>
  );
}

export function RemoveIconButton({ 'aria-label': ariaLabel = 'Remove', ...rest }: NativeBtn) {
  return (
    <button
      type="button"
      aria-label={ariaLabel}
      className={`${iconBtn} bg-gradient-to-br from-red-100 to-red-200 text-red-700 hover:from-red-200 hover:to-red-300 shadow-sm hover:shadow-md focus-visible:ring-red-400`}
      {...rest}
    >
      <span aria-hidden={true} className="absolute -right-4 top-1/2 -translate-y-1/2 w-8 h-8 rounded-full bg-red-800/10 pointer-events-none" />
      <Trash2 className="w-4 h-4 relative" />
    </button>
  );
}

export function CancelIconButton({ 'aria-label': ariaLabel = 'Cancel', ...rest }: NativeBtn) {
  return (
    <button
      type="button"
      aria-label={ariaLabel}
      className={`${iconBtn} bg-gradient-to-br from-slate-100 to-slate-200 text-slate-500 hover:from-slate-200 hover:to-slate-300 hover:text-slate-700 shadow-sm hover:shadow-md focus-visible:ring-slate-400`}
      {...rest}
    >
      <X className="w-4 h-4" />
    </button>
  );
}

export function ChangeIconButton({ 'aria-label': ariaLabel = 'Change', ...rest }: NativeBtn) {
  return (
    <button
      type="button"
      aria-label={ariaLabel}
      className={`${iconBtn} bg-gradient-to-br from-slate-700 via-slate-800 to-slate-950 text-white shadow-md hover:shadow-xl focus-visible:ring-slate-900`}
      {...rest}
    >
      <span aria-hidden={true} className="absolute -right-4 top-1/2 -translate-y-1/2 w-8 h-8 rounded-full bg-white/20 pointer-events-none" />
      <RefreshCw className="w-4 h-4 relative" />
    </button>
  );
}
