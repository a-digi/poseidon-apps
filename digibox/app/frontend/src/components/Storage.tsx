import React, { useEffect, useState } from 'react';
import { pluginApi } from '@/api/pluginApi';
import type { AccountInfo, StorageEntry } from '@/api/types';
import { useToast } from '@/contexts/ToastContext';
import { StorageBrowser, type BrowserEntry } from './StorageBrowser';
import PromptDialog from './Modal/PromptDialog';
import ConfirmationDialog from './Modal/ConfirmationDialog';

interface StorageProps {
  tokenId: string;
  account: AccountInfo;
  onBack: () => void;
}

interface FolderState {
  path: string;
  breadcrumb: { name: string; path: string }[];
}

const ROOT: FolderState = { path: '', breadcrumb: [] };

const toBrowserEntry = (e: StorageEntry, idx: number): BrowserEntry => ({
  id: e.id ?? `${e.path ?? e.name}-${idx}`,
  name: e.name,
  tag: e.tag === 'folder' ? 'folder' : 'file',
  path: e.path ?? '',
  downloaded: e.downloaded === true,
  targetFolder: e.targetFolder,
});

const PROVIDER_LABELS: Record<string, string> = {
  dropbox: 'Dropbox',
  googledrive: 'Google Drive',
};

const Storage: React.FC<StorageProps> = ({ tokenId, account, onBack }) => {
  const [folder, setFolder] = useState<FolderState>(ROOT);
  const [entries, setEntries] = useState<BrowserEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);
  const [deletingIds, setDeletingIds] = useState<Set<string>>(new Set());
  const [promptOpen, setPromptOpen] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState<BrowserEntry | null>(null);
  const { addToast } = useToast();

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    pluginApi
      .listFiles(tokenId, folder.path)
      .then((res) => {
        if (cancelled) return;
        setEntries((res.entries ?? []).map(toBrowserEntry));
      })
      .catch((e: unknown) => {
        if (cancelled) return;
        setError(e instanceof Error ? e.message : 'Failed to load files');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [tokenId, folder.path, refreshKey]);

  const handleFolderClick = (entry: BrowserEntry) => {
    setFolder((prev) => ({
      path: entry.path,
      breadcrumb: [...prev.breadcrumb, { name: entry.name, path: entry.path }],
    }));
  };

  const handleBreadcrumbClick = (index: number) => {
    if (index === -1) {
      setFolder(ROOT);
      return;
    }
    setFolder((prev) => {
      const newCrumb = prev.breadcrumb.slice(0, index + 1);
      return { path: newCrumb[newCrumb.length - 1]?.path ?? '', breadcrumb: newCrumb };
    });
  };

  const handleShowDownloadPath = (entry: BrowserEntry) => {
    if (entry.targetFolder) {
      addToast({ message: `Downloaded to: ${entry.targetFolder}`, type: 'info' });
    }
  };

  const handleCreateFolder = () => setPromptOpen(true);

  const submitCreateFolder = async (name: string) => {
    const target = `${folder.path}/${name}`;
    try {
      await pluginApi.createFolder(tokenId, target);
      addToast({ message: `Folder created: ${target}`, type: 'success' });
      setRefreshKey((k) => k + 1);
    } catch (err) {
      addToast({
        message: err instanceof Error ? err.message : 'Create folder failed',
        type: 'error',
      });
    }
  };

  const handleUploadFile = async () => {
    try {
      const sourcePath = await pluginApi.openFileDialog();
      if (!sourcePath) return;
      const filename = sourcePath.split('/').pop() ?? sourcePath;
      const target = `${folder.path}/${filename}`;
      const res = await pluginApi.uploadFile(tokenId, target, sourcePath);
      addToast({ message: `Uploaded ${filename} to ${res.path}`, type: 'success' });
      setRefreshKey((k) => k + 1);
    } catch (err) {
      addToast({
        message: err instanceof Error ? err.message : 'Upload failed',
        type: 'error',
      });
    }
  };

  const handleDeleteItem = (entry: BrowserEntry) => {
    if (deletingIds.has(entry.id)) return;
    setConfirmDelete(entry);
  };

  const submitDeleteItem = async (entry: BrowserEntry) => {
    setDeletingIds((prev) => {
      const next = new Set(prev);
      next.add(entry.id);
      return next;
    });
    try {
      await pluginApi.deleteItem(tokenId, entry.path);
      addToast({ message: `Deleted ${entry.name}`, type: 'success' });
      setRefreshKey((k) => k + 1);
    } catch (err) {
      addToast({
        message: err instanceof Error ? err.message : 'Delete failed',
        type: 'error',
      });
    } finally {
      setDeletingIds((prev) => {
        const next = new Set(prev);
        next.delete(entry.id);
        return next;
      });
    }
  };

  const handleFileDownload = async (entry: BrowserEntry) => {
    try {
      const targetDir = await pluginApi.openDirectoryDialog();
      if (!targetDir) return;
      const res = await pluginApi.downloadFile(tokenId, entry.path, targetDir);
      addToast({
        message: `Downloaded ${entry.name} to ${res.path}`,
        type: 'success',
      });
      setRefreshKey((k) => k + 1);
    } catch (err) {
      addToast({
        message: err instanceof Error ? err.message : 'Download failed',
        type: 'error',
      });
    }
  };

  return (
    <section className="w-full">
      <div className="sticky top-0 z-10 bg-white border-b border-slate-200 px-4 py-2 flex items-center gap-3 shadow-sm">
        <div className="flex items-center gap-2">
          {account.provider === 'dropbox' ? (
            <svg viewBox="0 0 32 32" className="w-6 h-6 shrink-0" fill="none">
              <rect width="32" height="32" rx="5" fill="#0061FF" />
              <path d="M8 10l8 5 8-5-8-5-8 5zm0 8l8 5 8-5-8-5-8 5z" fill="#fff" />
            </svg>
          ) : (
            <svg viewBox="0 0 32 32" className="w-6 h-6 shrink-0" fill="none">
              <rect width="32" height="32" rx="5" fill="#4285F4" />
              <path d="M16 8l-8 14h16L16 8z" fill="#fff" />
            </svg>
          )}
          <div className="min-w-0">
            <p className="text-sm font-semibold text-slate-900 leading-tight truncate">
              {account.displayName || account.email || PROVIDER_LABELS[account.provider] || account.provider}
            </p>
            {account.email && account.displayName && (
              <p className="text-xs text-slate-500 leading-tight truncate">{account.email}</p>
            )}
          </div>
        </div>
        <span className="ml-auto text-xs text-slate-400 shrink-0">
          {PROVIDER_LABELS[account.provider] ?? account.provider}
        </span>
      </div>
      <div className="p-4">
      <header className="flex items-center gap-3 mb-4">
        <button
          type="button"
          onClick={onBack}
          className="text-blue-700 hover:underline text-sm"
        >
          ← Back to Connections
        </button>
        <h1 className="text-2xl font-bold text-blue-800">Storage</h1>
        <div className="ml-auto flex gap-2">
          <button
            type="button"
            onClick={handleCreateFolder}
            className="px-3 py-1 text-sm rounded bg-blue-600 text-white hover:bg-blue-700"
          >
            + New Folder
          </button>
          <button
            type="button"
            onClick={handleUploadFile}
            className="px-3 py-1 text-sm rounded bg-blue-600 text-white hover:bg-blue-700"
          >
            Upload
          </button>
        </div>
      </header>
      {loading ? (
        <div className="text-blue-600 text-center py-8">Loading…</div>
      ) : error ? (
        <div className="text-red-600 text-center py-8">{error}</div>
      ) : (
        <StorageBrowser
          entries={entries}
          breadcrumb={folder.breadcrumb.map((b) => b.name)}
          onFolderClick={handleFolderClick}
          onBreadcrumbClick={handleBreadcrumbClick}
          onFileDownload={handleFileDownload}
          onShowDownloadPath={handleShowDownloadPath}
          onDeleteItem={handleDeleteItem}
          deletingIds={deletingIds}
        />
      )}
      <PromptDialog
        open={promptOpen}
        title="New folder"
        placeholder="Folder name"
        onCancel={() => setPromptOpen(false)}
        onSubmit={(name) => {
          setPromptOpen(false);
          void submitCreateFolder(name);
        }}
      />
      <ConfirmationDialog
        open={confirmDelete !== null}
        title={confirmDelete?.tag === 'folder' ? 'Delete folder?' : 'Delete file?'}
        message={
          confirmDelete
            ? `Delete "${confirmDelete.name}"?${confirmDelete.tag === 'folder' ? ' Folders are deleted recursively.' : ''} This cannot be undone.`
            : ''
        }
        confirmLabel="Delete"
        cancelLabel="Cancel"
        onClose={() => setConfirmDelete(null)}
        onConfirm={() => {
          const entry = confirmDelete;
          setConfirmDelete(null);
          if (entry) void submitDeleteItem(entry);
        }}
      />
      </div>
    </section>
  );
};

export default Storage;
