import { useEffect, useState } from 'react';
import { File, FileText, ImageOff, Trash2 } from 'lucide-react';
import { type Attachment, type AttachmentBytes, getAttachmentBytes } from '../../api';

interface AttachmentThumbnailProps {
  attachment: Attachment;
  onClick: () => void;
  onDelete: () => void;
}

type ImageState =
  | { kind: 'loading' }
  | { kind: 'error' }
  | { kind: 'loaded'; dataURL: string };

function ImageThumbnail({ attachment }: { attachment: Attachment }) {
  const [state, setState] = useState<ImageState>({ kind: 'loading' });

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

  if (state.kind === 'loading') {
    return (
      <div className="flex items-center justify-center w-full h-full">
        <div className="w-5 h-5 border-2 border-slate-300 border-t-blue-500 rounded-full animate-spin" />
      </div>
    );
  }

  if (state.kind === 'error') {
    return (
      <div className="flex items-center justify-center w-full h-full">
        <ImageOff className="w-6 h-6 text-slate-400" />
      </div>
    );
  }

  return (
    <img
      src={state.dataURL}
      alt={attachment.originalFilename}
      className="w-full h-full object-cover"
    />
  );
}

function FileThumbnail({ attachment }: { attachment: Attachment }) {
  const isPdf = attachment.mime === 'application/pdf';
  const Icon = isPdf ? FileText : File;

  return (
    <div className="flex flex-col items-center justify-center w-full h-full gap-1 px-1">
      <Icon className="w-8 h-8 text-slate-400" />
      <span className="text-xs text-slate-500 text-center truncate w-full px-1">
        {attachment.originalFilename}
      </span>
    </div>
  );
}

export function AttachmentThumbnail({ attachment, onClick, onDelete }: AttachmentThumbnailProps) {
  const isImage = attachment.mime.startsWith('image/');

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    onDelete();
  };

  return (
    <div
      className="relative aspect-square rounded-lg bg-slate-100 border border-slate-200 overflow-hidden cursor-pointer group"
      onClick={onClick}
    >
      {isImage ? (
        <ImageThumbnail attachment={attachment} />
      ) : (
        <FileThumbnail attachment={attachment} />
      )}

      <div className="absolute bottom-0 inset-x-0 bg-black/50 px-2 py-1 text-white text-xs truncate opacity-0 group-hover:opacity-100 transition-opacity">
        {attachment.originalFilename}
      </div>

      <button
        type="button"
        aria-label="Delete"
        onClick={handleDelete}
        className="absolute top-1 right-1 opacity-0 group-hover:opacity-100 transition-opacity p-1 rounded bg-black/40 hover:bg-red-600 text-white"
      >
        <Trash2 className="w-3.5 h-3.5" />
      </button>
    </div>
  );
}
