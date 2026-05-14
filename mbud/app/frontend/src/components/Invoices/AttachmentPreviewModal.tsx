import { useEffect, useState } from 'react';
import { Download, File, X } from 'lucide-react';
import { type Attachment, type AttachmentBytes, getAttachmentBytes } from '../../api';

interface AttachmentPreviewModalProps {
  attachment: Attachment;
  onClose: () => void;
}

type BytesState =
  | { kind: 'loading' }
  | { kind: 'error' }
  | { kind: 'loaded'; dataURL: string };

export function AttachmentPreviewModal({ attachment, onClose }: AttachmentPreviewModalProps) {
  const [state, setState] = useState<BytesState>({ kind: 'loading' });

  useEffect(() => {
    let cancelled = false;

    getAttachmentBytes(attachment.id)
      .then((bytes: AttachmentBytes) => {
        if (!cancelled) {
          setState({ kind: 'loaded', dataURL: `data:${bytes.mime};base64,${bytes.dataBase64}` });
        }
      })
      .catch(() => {
        if (!cancelled) setState({ kind: 'error' });
      });

    return () => { cancelled = true; };
  }, [attachment.id]);

  const isImage = attachment.mime.startsWith('image/');
  const isPdf = attachment.mime === 'application/pdf';
  const dataURL = state.kind === 'loaded' ? state.dataURL : null;

  return (
    <div
      className="fixed inset-0 z-50 flex flex-col items-center justify-center bg-black/80 p-4"
      onClick={onClose}
    >
      <div
        className="flex flex-col items-center gap-3 w-full"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-end gap-2 w-full max-w-5xl">
          {dataURL && (
            <a
              href={dataURL}
              download={attachment.originalFilename}
              className="px-3 py-1.5 rounded-md bg-white/10 hover:bg-white/20 text-white text-sm flex items-center gap-1.5 transition-colors"
            >
              <Download className="w-4 h-4" />
              Download
            </a>
          )}
          <button
            type="button"
            onClick={onClose}
            className="p-1.5 rounded-md bg-white/10 hover:bg-white/20 text-white transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {state.kind === 'loading' && (
          <div className="flex items-center justify-center py-16">
            <div className="w-8 h-8 border-4 border-slate-300 border-t-blue-500 rounded-full animate-spin" />
          </div>
        )}

        {state.kind === 'error' && (
          <p className="text-white text-sm">Failed to load file.</p>
        )}

        {state.kind === 'loaded' && isImage && (
          <img
            src={state.dataURL}
            alt={attachment.originalFilename}
            className="max-w-full max-h-[85vh] object-contain rounded-lg"
          />
        )}

        {state.kind === 'loaded' && isPdf && (
          <iframe
            src={state.dataURL}
            title={attachment.originalFilename}
            className="w-[80vw] h-[85vh] rounded-lg bg-white"
          />
        )}

        {state.kind === 'loaded' && !isImage && !isPdf && (
          <div className="bg-white rounded-xl p-8 flex flex-col items-center gap-4">
            <File className="w-12 h-12 text-slate-400" />
            <span className="text-sm text-slate-700 text-center break-all">
              {attachment.originalFilename}
            </span>
            <a
              href={state.dataURL}
              download={attachment.originalFilename}
              className="px-3 py-1.5 rounded-md bg-slate-100 hover:bg-slate-200 text-slate-700 text-sm flex items-center gap-1.5 transition-colors"
            >
              <Download className="w-4 h-4" />
              Download
            </a>
          </div>
        )}
      </div>
    </div>
  );
}
