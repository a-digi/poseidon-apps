import { useRef, useState } from 'react';
import { Upload } from 'lucide-react';
import { uploadInvoiceAttachment } from '../../api';

interface DeviceUploadButtonProps {
  invoiceId: string;
  onUploaded: () => void;
}

const ACCEPTED = 'image/png,image/jpeg,image/webp,application/pdf';

function readAsBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      resolve((reader.result as string).replace(/^data:[^;]+;base64,/, ''));
    };
    reader.onerror = () => reject(new Error('failed to read'));
    reader.readAsDataURL(file);
  });
}

const tileClass =
  'px-3 py-2.5 rounded-lg border border-dashed border-slate-200 text-slate-400 hover:border-blue-400 hover:text-blue-500 hover:bg-blue-50 transition-colors cursor-pointer flex items-center gap-1.5';

export function DeviceUploadButton({ invoiceId, onUploaded }: DeviceUploadButtonProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [progress, setProgress] = useState<{ current: number; total: number } | null>(null);
  const [errors, setErrors] = useState<string[]>([]);

  const busy = progress !== null;

  async function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    const files = Array.from(e.target.files ?? []);
    if (files.length === 0) return;

    setErrors([]);
    setProgress({ current: 0, total: files.length });

    const failed: string[] = [];

    for (let i = 0; i < files.length; i++) {
      const file = files[i];
      setProgress({ current: i + 1, total: files.length });
      try {
        const base64 = await readAsBase64(file);
        await uploadInvoiceAttachment(invoiceId, file.name, file.type, base64);
      } catch {
        failed.push(file.name);
      }
    }

    setProgress(null);
    if (inputRef.current) inputRef.current.value = '';
    setErrors(failed);
    onUploaded();
  }

  return (
    <div className="flex flex-col gap-1">
      <label
        aria-label="Upload from device"
        title="Upload from device"
        className={busy ? `${tileClass} opacity-50 pointer-events-none` : tileClass}
      >
        {busy ? (
          <div className="w-5 h-5 border-2 border-slate-300 border-t-blue-500 rounded-full animate-spin" />
        ) : (
          <Upload className="w-5 h-5" />
        )}
        <span className="text-xs font-medium">Device</span>
        <input
          ref={inputRef}
          type="file"
          multiple
          accept={ACCEPTED}
          className="sr-only"
          disabled={busy}
          onChange={handleChange}
        />
      </label>
      {errors.length > 0 && (
        <p className="text-xs text-red-600">Failed: {errors.join(', ')}</p>
      )}
    </div>
  );
}
