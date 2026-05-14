import { useState } from 'react';
import { Smartphone } from 'lucide-react';
import { useIsMobile } from '../../lib/useIsMobile';
import { AttachmentDialog } from './AttachmentDialog';
import { DeviceUploadButton } from './DeviceUploadButton';
import { InvoiceAttachments } from './InvoiceAttachments';

interface InvoiceAttachmentPanelProps {
  invoiceId: string;
}

const phoneTileClass =
  'px-3 py-2.5 rounded-lg border border-dashed border-slate-200 text-slate-400 hover:border-blue-400 hover:text-blue-500 hover:bg-blue-50 transition-colors flex items-center gap-1.5';

export function InvoiceAttachmentPanel({ invoiceId }: InvoiceAttachmentPanelProps) {
  const [refreshKey, setRefreshKey] = useState(0);
  const [showPhoneDialog, setShowPhoneDialog] = useState(false);
  const isMobile = useIsMobile();

  const bump = () => setRefreshKey(k => k + 1);

  return (
    <>
      <InvoiceAttachments invoiceId={invoiceId} refreshKey={refreshKey} />
      <div className="flex items-start gap-2">
        <DeviceUploadButton invoiceId={invoiceId} onUploaded={bump} />
        {!isMobile && (
          <button
            type="button"
            onClick={() => setShowPhoneDialog(true)}
            aria-label="Upload from phone"
            title="Upload from phone"
            className={phoneTileClass}
          >
            <Smartphone className="w-5 h-5" />
            <span className="text-xs font-medium">Phone</span>
          </button>
        )}
      </div>
      {showPhoneDialog && (
        <AttachmentDialog
          invoiceId={invoiceId}
          onClose={() => { setShowPhoneDialog(false); bump(); }}
        />
      )}
    </>
  );
}
