import React from 'react';

export interface BrowserEntry {
  id: string;
  name: string;
  tag: 'folder' | 'file';
  path: string;
  size?: number;
  modified?: string;
  downloaded?: boolean;
  targetFolder?: string;
}

interface StorageBrowserProps {
  entries: BrowserEntry[];
  breadcrumb: string[];
  onFolderClick: (entry: BrowserEntry) => void;
  onBreadcrumbClick: (index: number) => void;
  onFileDownload: (entry: BrowserEntry) => void;
  onShowDownloadPath: (entry: BrowserEntry) => void;
  onDeleteItem: (entry: BrowserEntry) => void;
  deletingIds: Set<string>;
}

const FolderIcon = () => (
  <svg className="w-6 h-6 text-yellow-500 mr-2" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
    <path d="M3 7a2 2 0 012-2h4l2 2h8a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2V7z" strokeLinejoin="round" />
  </svg>
);

const FileIcon = () => (
  <svg className="w-6 h-6 text-blue-500 mr-2" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
    <path d="M6 2a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8.828a2 2 0 00-.586-1.414l-4.828-4.828A2 2 0 0014.172 2H6z" strokeLinejoin="round" />
  </svg>
);

const DownloadIcon = () => (
  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
    <path d="M12 16v-8m0 8l-4-4m4 4l4-4M4 20h16" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
);

const formatSize = (bytes?: number) => {
  if (!bytes) return '';
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1048576) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1048576).toFixed(1)} MB`;
};

export const StorageBrowser: React.FC<StorageBrowserProps> = ({
  entries,
  breadcrumb,
  onFolderClick,
  onBreadcrumbClick,
  onFileDownload,
  onShowDownloadPath,
  onDeleteItem,
  deletingIds,
}) => {
  return (
    <div className="bg-white p-2 max-w-full mx-auto mt-4">
      <nav className="mb-4 flex items-center text-sm text-blue-700 font-medium flex-wrap">
        <span
          className={`cursor-pointer hover:underline ${breadcrumb.length === 0 ? 'font-bold' : ''}`}
          onClick={() => onBreadcrumbClick(-1)}
        >
          Root
        </span>
        {breadcrumb.map((name, idx) => (
          <React.Fragment key={`${idx}-${name}`}>
            <span className="mx-2 text-blue-300">/</span>
            <span
              className={`cursor-pointer hover:underline ${idx === breadcrumb.length - 1 ? 'font-bold' : ''}`}
              onClick={() => onBreadcrumbClick(idx)}
            >
              {name}
            </span>
          </React.Fragment>
        ))}
      </nav>
      {entries.length === 0 ? (
        <div className="text-gray-400 py-8 text-center">This folder is empty.</div>
      ) : (
        <div className="divide-y divide-blue-50">
          {entries.map((entry) => {
            const isDeleting = deletingIds.has(entry.id);
            const rowClass = isDeleting
              ? 'opacity-40 pointer-events-none cursor-not-allowed grayscale'
              : `hover:bg-blue-50 ${entry.tag === 'folder' ? 'cursor-pointer font-semibold' : ''}`;
            return (
              <div
                key={entry.id}
                className={`flex items-center px-4 py-2 rounded transition-colors ${rowClass}`}
                onClick={() => {
                  if (isDeleting) return;
                  if (entry.tag === 'folder') onFolderClick(entry);
                }}
              >
                {entry.tag === 'folder' ? <FolderIcon /> : <FileIcon />}
                {entry.downloaded && (
                  <span title="Downloaded" className="mr-1 text-green-600 inline-flex items-center" aria-label="Downloaded">
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2.5" viewBox="0 0 24 24">
                      <path d="M5 13l4 4L19 7" strokeLinecap="round" strokeLinejoin="round" />
                    </svg>
                  </span>
                )}
                <span className="flex-1 truncate">
                  {entry.name}
                  {isDeleting && <span className="ml-2 text-xs text-gray-500 italic">deleting…</span>}
                </span>
                <span className="text-xs text-gray-500 mr-3">{formatSize(entry.size)}</span>
                {entry.tag === 'file' && entry.downloaded && (
                  <button
                    type="button"
                    className="p-1 mr-1 rounded text-green-600 hover:bg-green-100 hover:text-green-700 transition-colors"
                    aria-label={`Show download path for ${entry.name}`}
                    title={entry.targetFolder ? `Downloaded to: ${entry.targetFolder}` : 'Downloaded'}
                    onClick={(e) => {
                      e.stopPropagation();
                      onShowDownloadPath(entry);
                    }}
                  >
                    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                      <circle cx="12" cy="12" r="9" />
                      <path d="M8 12l3 3 5-6" strokeLinecap="round" strokeLinejoin="round" />
                    </svg>
                  </button>
                )}
                {entry.tag === 'file' && (
                  <button
                    type="button"
                    className="p-1 rounded text-blue-500 hover:bg-blue-100 hover:text-blue-700 transition-colors"
                    aria-label={`Download ${entry.name}`}
                    onClick={(e) => {
                      e.stopPropagation();
                      onFileDownload(entry);
                    }}
                  >
                    <DownloadIcon />
                  </button>
                )}
                <button
                  type="button"
                  className="p-1 ml-1 rounded text-red-500 hover:bg-red-100 hover:text-red-700 transition-colors"
                  aria-label={`Delete ${entry.name}`}
                  title={entry.tag === 'folder' ? 'Delete folder (recursive)' : 'Delete file'}
                  onClick={(e) => {
                    e.stopPropagation();
                    onDeleteItem(entry);
                  }}
                >
                  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M3 6h18" />
                    <path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
                    <path d="M6 6l1 14a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2l1-14" />
                    <path d="M10 11v6" />
                    <path d="M14 11v6" />
                  </svg>
                </button>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};

export default StorageBrowser;
