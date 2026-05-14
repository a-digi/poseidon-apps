import React, { useEffect, useRef, useState } from 'react';
import { getUploadSession, mobileUploadFile, type Attachment } from '../api';

type PageState = 'loading' | 'invalid' | 'ready' | 'success';

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export function MobileUploaderPage() {
  const sp = new URLSearchParams(window.location.search);
  const token = sp.get('t') ?? '';

  const [pageState, setPageState] = useState<PageState>('loading');
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const [uploaded, setUploaded] = useState<Attachment | null>(null);

  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (!token) {
      setPageState('invalid');
      return;
    }
    getUploadSession(token)
      .then(({ session }) => {
        if (session.status !== 'active') {
          setPageState('invalid');
        } else {
          setPageState('ready');
        }
      })
      .catch(() => setPageState('invalid'));
  }, [token]);

  function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0] ?? null;
    setSelectedFile(file);
    setUploadError(null);
  }

  function handleUpload() {
    if (!selectedFile) return;
    setUploading(true);
    setUploadError(null);

    const reader = new FileReader();
    reader.onload = () => {
      const dataUrl = reader.result as string;
      const base64 = dataUrl.replace(/^data:[^;]+;base64,/, '');
      mobileUploadFile(token, selectedFile.name, selectedFile.type, base64)
        .then((attachment) => {
          setUploaded(attachment);
          setPageState('success');
        })
        .catch((err: unknown) => {
          setUploadError(err instanceof Error ? err.message : 'Upload failed');
        })
        .finally(() => setUploading(false));
    };
    reader.onerror = () => {
      setUploadError('Failed to read file');
      setUploading(false);
    };
    reader.readAsDataURL(selectedFile);
  }

  function handleReset() {
    setSelectedFile(null);
    setUploaded(null);
    setUploadError(null);
    setPageState('ready');
    if (fileInputRef.current) fileInputRef.current.value = '';
  }

  return (
    <div className="min-h-screen bg-gray-950 text-white">
      <div className="max-w-sm mx-auto px-4 py-8">
        {pageState === 'loading' && (
          <div className="flex justify-center items-center min-h-[40vh]">
            <div className="animate-spin border border-t-transparent rounded-full w-8 h-8" />
          </div>
        )}

        {pageState === 'invalid' && (
          <div className="flex justify-center items-center min-h-[40vh]">
            <div className="bg-gray-900 rounded-2xl p-6 text-center">
              <p className="text-gray-300 text-base">
                This upload link has expired or is invalid.
              </p>
            </div>
          </div>
        )}

        {pageState === 'ready' && (
          <div className="flex flex-col gap-4">
            <label className="flex items-center justify-center gap-2 w-full min-h-16 bg-blue-600 hover:bg-blue-500 active:bg-blue-700 rounded-xl cursor-pointer transition-colors text-white font-medium text-base">
              <span>Choose file or take photo</span>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/png,image/jpeg,image/webp,application/pdf"
                capture="environment"
                className="sr-only"
                onChange={handleFileChange}
              />
            </label>

            {selectedFile && (
              <div className="bg-gray-900 rounded-xl px-4 py-3 text-sm text-gray-300 break-all">
                <span className="font-medium text-white">{selectedFile.name}</span>
                <span className="ml-2 text-gray-500">{formatBytes(selectedFile.size)}</span>
              </div>
            )}

            {uploadError && (
              <p className="text-red-400 text-sm">{uploadError}</p>
            )}

            <button
              onClick={handleUpload}
              disabled={!selectedFile || uploading}
              className="flex items-center justify-center gap-2 w-full min-h-16 bg-green-600 hover:bg-green-500 active:bg-green-700 disabled:opacity-40 disabled:cursor-not-allowed rounded-xl transition-colors text-white font-medium text-base"
            >
              {uploading ? (
                <span className="animate-spin border border-t-transparent rounded-full w-5 h-5" />
              ) : (
                'Upload'
              )}
            </button>
          </div>
        )}

        {pageState === 'success' && uploaded && (
          <div className="flex flex-col gap-4">
            <div className="bg-gray-900 rounded-2xl p-6 text-center flex flex-col gap-2">
              <p className="text-green-400 font-semibold text-base">Uploaded successfully</p>
              <p className="text-gray-300 text-sm break-all">{uploaded.originalFilename}</p>
            </div>
            <button
              onClick={handleReset}
              className="flex items-center justify-center gap-2 w-full min-h-16 bg-blue-600 hover:bg-blue-500 active:bg-blue-700 rounded-xl transition-colors text-white font-medium text-base"
            >
              Upload another file
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
