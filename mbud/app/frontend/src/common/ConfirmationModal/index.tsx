import { useCallback, useEffect, useState } from 'react';

export interface ConfirmOptions {
  message: string;
  confirmLabel?: string;
  destructive?: boolean;
}

interface InternalRequest {
  message: string;
  confirmLabel: string;
  destructive: boolean;
  resolve: (b: boolean) => void;
}

interface ModalProps {
  message: string;
  confirmLabel: string;
  destructive: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

function ConfirmationModal({ message, confirmLabel, destructive, onConfirm, onCancel }: ModalProps) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onCancel(); };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onCancel]);

  const confirmCls = destructive
    ? 'px-4 py-2 rounded-md text-sm font-semibold bg-gradient-to-br from-red-600 to-red-800 text-white shadow-sm hover:shadow-md hover:from-red-500 hover:to-red-700 active:scale-95 transition-all duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-red-400 focus-visible:ring-offset-2'
    : 'px-4 py-2 rounded-md text-sm font-semibold bg-gradient-to-r from-slate-900 to-slate-700 text-white shadow-sm hover:shadow-md active:scale-95 transition-all duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-900 focus-visible:ring-offset-2';

  const cancelCls = 'px-4 py-2 rounded-md text-sm font-semibold bg-gradient-to-br from-slate-200 to-slate-300 text-slate-800 hover:from-slate-300 hover:to-slate-400 active:scale-95 transition-all duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-400 focus-visible:ring-offset-2';

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      onClick={onCancel}
    >
      <div
        role="dialog"
        aria-modal="true"
        className="bg-white rounded-xl shadow-xl p-6 max-w-sm w-full"
        onClick={e => e.stopPropagation()}
      >
        <p className="text-sm text-slate-900 mb-4">{message}</p>
        <div className="flex justify-end gap-2">
          <button type="button" onClick={onCancel} className={cancelCls}>Cancel</button>
          <button type="button" onClick={onConfirm} autoFocus className={confirmCls}>{confirmLabel}</button>
        </div>
      </div>
    </div>
  );
}

export function useConfirm() {
  const [request, setRequest] = useState<InternalRequest | null>(null);

  const confirm = useCallback((opts: ConfirmOptions | string) =>
    new Promise<boolean>(resolve => {
      const o = typeof opts === 'string' ? { message: opts } : opts;
      setRequest({
        message: o.message,
        confirmLabel: o.confirmLabel ?? 'Delete',
        destructive: o.destructive ?? true,
        resolve,
      });
    }), []);

  const close = (result: boolean) => {
    if (!request) return;
    request.resolve(result);
    setRequest(null);
  };

  const dialog = request && (
    <ConfirmationModal
      message={request.message}
      confirmLabel={request.confirmLabel}
      destructive={request.destructive}
      onConfirm={() => close(true)}
      onCancel={() => close(false)}
    />
  );

  return { confirm, dialog };
}
