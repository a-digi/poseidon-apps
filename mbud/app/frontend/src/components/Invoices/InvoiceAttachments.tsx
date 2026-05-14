import { useEffect, useState } from 'react';
import { type Attachment, listInvoiceAttachments, deleteAttachment } from '../../api';
import { AttachmentThumbnail } from './AttachmentThumbnail';
import { AttachmentPreviewModal } from './AttachmentPreviewModal';

interface InvoiceAttachmentsProps {
  invoiceId: string;
  refreshKey: number;
  onChange?: () => void;
}

export function InvoiceAttachments({ invoiceId, refreshKey, onChange }: InvoiceAttachmentsProps) {
  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const [previewing, setPreviewing] = useState<Attachment | null>(null);

  useEffect(() => {
    listInvoiceAttachments(invoiceId)
      .then(setAttachments)
      .catch(() => setAttachments([]));
  }, [invoiceId, refreshKey]);

  const handleDelete = async (id: string) => {
    setAttachments(prev => prev.filter(a => a.id !== id));
    try {
      await deleteAttachment(id);
    } catch {
      const refetched = await listInvoiceAttachments(invoiceId).catch(() => []);
      setAttachments(refetched);
    }
    onChange?.();
  };

  if (attachments.length === 0) return null;

  return (
    <>
      <div className="flex flex-col gap-2">
        <p className="text-sm font-medium text-slate-700">Attachments ({attachments.length})</p>
        <div className="grid grid-cols-3 gap-2">
          {attachments.map(attachment => (
            <AttachmentThumbnail
              key={attachment.id}
              attachment={attachment}
              onClick={() => setPreviewing(attachment)}
              onDelete={() => handleDelete(attachment.id)}
            />
          ))}
        </div>
      </div>
      {previewing !== null && (
        <AttachmentPreviewModal attachment={previewing} onClose={() => setPreviewing(null)} />
      )}
    </>
  );
}
