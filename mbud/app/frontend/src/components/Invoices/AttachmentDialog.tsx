import { useEffect, useRef, useState } from 'react';
import { X, Trash2 } from 'lucide-react';
import QRCode from 'qrcode';
import {
  type Attachment,
  attachSessionToInvoice,
  createUploadSession,
  deleteAttachment,
  getLanAddresses,
  getUploadSession,
} from '../../api';

interface AttachmentDialogProps {
  invoiceId?: string;
  onClose: () => void;
}

type Phase =
  | { kind: 'loading' }
  | { kind: 'error'; message: string }
  | { kind: 'ready'; url: string; token: string; qrDataUrl: string };

function mimeBadge(mime: string): string {
  if (mime.startsWith('image/')) return 'Image';
  if (mime === 'application/pdf') return 'PDF';
  return mime;
}

export function AttachmentDialog({ invoiceId, onClose }: AttachmentDialogProps) {
  const [phase, setPhase] = useState<Phase>({ kind: 'loading' });
  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const [copied, setCopied] = useState(false);
  const [attaching, setAttaching] = useState(false);
  const [attachError, setAttachError] = useState<string | null>(null);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function init() {
      try {
        const [addresses, session] = await Promise.all([
          getLanAddresses(),
          createUploadSession(invoiceId),
        ]);

        if (cancelled) return;

        if (addresses.length === 0) {
          setPhase({ kind: 'error', message: 'No LAN address found. Connect the desktop to a Wi-Fi network and try again.' });
          return;
        }

        const ip = addresses[0];
        const url = `http://${ip}:2014/plugins/mbud/?mode=mobile&t=${session.token}`;
        const qrDataUrl = await QRCode.toDataURL(url, { errorCorrectionLevel: 'M', margin: 2, scale: 6 });

        if (cancelled) return;

        setPhase({ kind: 'ready', url, token: session.token, qrDataUrl });
      } catch (err) {
        if (!cancelled) {
          setPhase({ kind: 'error', message: err instanceof Error ? err.message : String(err) });
        }
      }
    }

    init();
    return () => { cancelled = true; };
  }, [invoiceId]);

  useEffect(() => {
    if (phase.kind !== 'ready') return;

    const token = phase.token;

    pollRef.current = setInterval(async () => {
      try {
        const result = await getUploadSession(token);
        setAttachments(result.attachments);
      } catch {
        // silently ignore poll errors
      }
    }, 1500);

    return () => {
      if (pollRef.current !== null) clearInterval(pollRef.current);
    };
  }, [phase.kind === 'ready' ? phase.token : null]);

  const handleCopy = () => {
    if (phase.kind !== 'ready') return;
    navigator.clipboard.writeText(phase.url).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

  const handleDeleteAttachment = async (id: string) => {
    await deleteAttachment(id);
    setAttachments(prev => prev.filter(a => a.id !== id));
  };

  const handleAttach = async () => {
    if (phase.kind !== 'ready' || !invoiceId) return;
    setAttaching(true);
    setAttachError(null);
    try {
      await attachSessionToInvoice(phase.token, invoiceId);
      onClose();
    } catch (err) {
      setAttachError(err instanceof Error ? err.message : String(err));
    } finally {
      setAttaching(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div className="bg-white rounded-xl shadow-xl p-6 max-w-lg w-full max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-slate-900">Upload from phone</h2>
          <button type="button" onClick={onClose} aria-label="Close" className="text-slate-400 hover:text-slate-700 transition-colors">
            <X className="w-5 h-5" />
          </button>
        </div>

        {phase.kind === 'loading' && (
          <div className="flex items-center justify-center py-12">
            <div className="w-8 h-8 border-4 border-slate-200 border-t-blue-500 rounded-full animate-spin" />
          </div>
        )}

        {phase.kind === 'error' && (
          <div className="text-sm text-red-700 bg-red-50 border border-red-200 rounded-md px-4 py-3">
            {phase.message}
          </div>
        )}

        {phase.kind === 'ready' && (
          <div className="flex flex-col gap-4">
            {invoiceId ? (
              <p className="text-sm text-slate-600">Scan to attach files</p>
            ) : (
              <p className="text-sm text-slate-500">Save the invoice first to attach uploaded files.</p>
            )}

            <img src={phase.qrDataUrl} alt="QR code" className="w-64 h-64 mx-auto rounded-lg" />

            <div className="flex items-center gap-2">
              <code className="flex-1 text-xs bg-slate-100 px-3 py-2 rounded-md text-slate-700 break-all font-mono">
                {phase.url}
              </code>
              <button
                type="button"
                onClick={handleCopy}
                className="shrink-0 px-3 py-2 rounded-md border border-slate-300 bg-white text-sm text-slate-700 hover:bg-slate-50 transition-colors"
              >
                {copied ? 'Copied' : 'Copy'}
              </button>
            </div>

            {invoiceId && (
              <div className="flex flex-col gap-2">
                <span className="text-sm font-medium text-slate-700">Uploaded files</span>
                {attachments.length === 0 ? (
                  <p className="text-sm text-slate-400">No files yet.</p>
                ) : (
                  <ul className="flex flex-col gap-1">
                    {attachments.map(a => (
                      <li key={a.id} className="flex items-center gap-2 py-1.5 px-3 rounded-md bg-slate-50 border border-slate-100">
                        <span className="flex-1 text-sm text-slate-800 truncate">{a.originalFilename}</span>
                        <span className="shrink-0 text-xs bg-slate-200 text-slate-600 rounded px-1.5 py-0.5">{mimeBadge(a.mime)}</span>
                        <button
                          type="button"
                          aria-label="Delete attachment"
                          onClick={() => handleDeleteAttachment(a.id)}
                          className="shrink-0 text-slate-400 hover:text-red-500 transition-colors"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            )}

            {attachError && (
              <div className="text-sm text-red-700 bg-red-50 border border-red-200 rounded-md px-3 py-2">{attachError}</div>
            )}

            {invoiceId && attachments.length > 0 && (
              <button
                type="button"
                onClick={handleAttach}
                disabled={attaching}
                className="w-full px-3 py-2 rounded-md bg-blue-600 text-white text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors"
              >
                {attaching ? 'Attaching...' : 'Attach to invoice'}
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
