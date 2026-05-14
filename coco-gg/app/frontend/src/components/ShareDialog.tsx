import { useEffect, useState } from 'react';
import { X } from 'lucide-react';
import QRCode from 'qrcode';
import { getLanAddresses, createMobileSession } from '../api';

interface ShareDialogProps {
  roomCode: string;
  onClose: () => void;
}

type Phase =
  | { kind: 'loading' }
  | { kind: 'error'; message: string }
  | { kind: 'ready'; url: string; qrDataUrl: string; expiresAt: number };

export function ShareDialog({ roomCode, onClose }: ShareDialogProps) {
  const [phase, setPhase] = useState<Phase>({ kind: 'loading' });
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    let cancelled = false;

    async function init() {
      try {
        const [addresses, session] = await Promise.all([
          getLanAddresses(),
          createMobileSession('coco-gg'),
        ]);

        if (cancelled) return;

        if (addresses.length === 0) {
          setPhase({
            kind: 'error',
            message: 'No LAN address found. Connect the desktop to a Wi-Fi network and try again.',
          });
          return;
        }

        const url = `http://${addresses[0]}:2014/plugins/coco-gg/?mode=mobile&t=${session.token}&room=${roomCode}`;
        const qrDataUrl = await QRCode.toDataURL(url, {
          errorCorrectionLevel: 'M',
          margin: 2,
          scale: 6,
        });

        if (cancelled) return;

        setPhase({ kind: 'ready', url, qrDataUrl, expiresAt: session.expiresAt });
      } catch (err) {
        if (!cancelled) {
          setPhase({
            kind: 'error',
            message: err instanceof Error ? err.message : String(err),
          });
        }
      }
    }

    init();
    return () => {
      cancelled = true;
    };
  }, [roomCode]);

  const handleCopy = () => {
    if (phase.kind !== 'ready') return;
    navigator.clipboard.writeText(phase.url).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div className="bg-white rounded-xl shadow-xl p-6 max-w-lg w-full">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-slate-900">Invite players</h2>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close"
            className="text-slate-400 hover:text-slate-700 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="flex flex-col items-center gap-1 mb-4">
          <span className="text-xs uppercase tracking-wide text-slate-500">Room</span>
          <span className="text-3xl font-bold tracking-widest text-slate-900">{roomCode}</span>
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
            <p className="text-sm text-slate-600">Scan the QR code on your phone. Expires in 1 hour.</p>

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
          </div>
        )}
      </div>
    </div>
  );
}
