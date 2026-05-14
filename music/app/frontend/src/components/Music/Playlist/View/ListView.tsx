import React, { useState, useRef, useEffect } from 'react';
import type { PlaylistIndex } from '../Playlist';
import DigitalItemGridView from './DigitalItemGridView';
import DigitalItemListView from '@/components/Music/Playlist/View/DigitalItemListView.tsx';
import ListViewModeButton from '@/components/ui/Button/ListViewModeButton';
import GridViewModeButton from '@/components/ui/Button/GridViewModeButton';
import CreateButton from '@/components/ui/Button/CreateButton.tsx';

// Typdefinition für window.go, um TypeScript-Fehler zu vermeiden
// @ts-ignore
declare global {
  interface Window {
    go?: any;
  }
}

interface ListViewProps {
  playlists: PlaylistIndex[];
  selectedPlaylistId: string | null;
  onSelect: (playlistId: string, item: any) => void;
  onAddItem: (playlistId: string, file?: File, filePath?: string) => void;
  onDelete: (id: string) => void;
  onEdit: (id: string, name: string) => void;
  onEditSave: (id: string) => void;
  editingId: string | null;
  editName: string;
  setEditName: (name: string) => void;
  optionsId: string | null;
  setOptionsId: (id: string | null) => void;
  t: any;
  handleDeleteItem: (playlistId: string, itemId: string) => void;
  addToast: any;
  addNewItemComponent?: React.ReactNode;
}

const ListView: React.FC<ListViewProps> = ({
  playlists,
  selectedPlaylistId,
  onSelect,
  onAddItem,
  onDelete,
  onEdit,
  onEditSave,
  editingId,
  editName,
  setEditName,
  optionsId,
  setOptionsId,
  t,
  handleDeleteItem,
  addToast,
  addNewItemComponent,
}) => {
  const [viewMode, setViewMode] = useState<'list' | 'grid'>('list');
  const selectedPlaylist = playlists.find((p) => p.id === selectedPlaylistId);

  // Click-Outside-Handler für Dropdown
  const dropdownRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      // Prüfe, ob der Klick im Dropdown oder einem seiner Kinder war
      const target = event.target as HTMLElement;
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(target) &&
        !target.closest('.dropdown-menu')
      ) {
        setOptionsId(null);
      }
    }
    if (optionsId) {
      document.addEventListener('mousedown', handleClickOutside);
    }
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [optionsId, setOptionsId]);

  return (
    <>
      <div className="flex items-center justify-between mb-10">
        <div className="flex gap-2 items-center">
          <ListViewModeButton
            active={viewMode === 'list'}
            rounded="left"
            onClick={() => setViewMode('list')}
            aria-pressed={viewMode === 'list'}
          />
          <GridViewModeButton
            active={viewMode === 'grid'}
            onClick={() => setViewMode('grid')}
            aria-pressed={viewMode === 'grid'}
          />
        </div>
        {addNewItemComponent && <div className="flex items-center">{addNewItemComponent}</div>}
      </div>
      <ul className="space-y-4">
        {playlists.map((playlist) => (
          <li key={playlist.id} className="border rounded p-3 flex flex-col gap-2">
            <div className="flex justify-between items-center relative">
              <div
                onClick={() => {
                  onSelect(playlist.id, null);
                }}
                className="flex cursor-pointer items-center gap-2"
              >
                <span className="font-semibold text-lg">{playlist.name}</span>
              </div>

              <div className="flex items-center gap-2 ml-auto relative">
                <CreateButton
                  label={t('playlist.addItem', 'Add item')}
                  className="flex items-center gap-1 px-2 py-1 min-w-0"
                  style={{ minWidth: '0', height: '28px' }}
                  onClick={async (e) => {
                    e.stopPropagation();
                    if (window.go?.main?.App?.OpenFileDialog) {
                      try {
                        const filePath = await window.go.main.App.OpenFileDialog();
                        if (filePath) {
                          onAddItem(playlist.id, undefined, filePath);
                          onSelect(playlist.id, playlist.items?.[0]);
                        }
                      } catch (err) {
                        addToast({ message: t('playlist.addMp3Error'), type: 'error' });
                      }
                    }
                  }}
                  aria-label={t('playlist.addMp3')}
                />
                <div className="relative" ref={dropdownRef}>
                  <button
                    className="text-gray-500 hover:text-blue-600 px-2 py-1 rounded text-2xl"
                    onClick={(e) => {
                      e.stopPropagation();
                      setOptionsId(optionsId === playlist.id ? null : playlist.id);
                    }}
                    aria-label={t('playlist.options')}
                  >
                    &#8942;
                  </button>
                  {optionsId === playlist.id && editingId !== playlist.id && (
                    <div
                      className="absolute right-0 mt-2 w-32 bg-white border border-gray-200 rounded shadow-lg z-50 flex flex-col dropdown-menu"
                      tabIndex={-1}
                    >
                      <button
                        onClick={() => {
                          console.log('Edit Playlist');
                          setOptionsId(null);
                          onEdit(playlist.id, playlist.name);
                        }}
                        className="px-4 py-2 text-left hover:bg-yellow-100 text-yellow-900 rounded-t"
                      >
                        {t('playlist.edit')}
                      </button>
                      <button
                        onClick={() => {
                          console.log('Delete Playlist');
                          setOptionsId(null);
                          onDelete(playlist.id);
                        }}
                        className="px-4 py-2 text-left hover:bg-red-100 text-red-700 rounded-b"
                      >
                        {t('playlist.delete')}
                      </button>
                    </div>
                  )}
                </div>
              </div>
            </div>
            {editingId === playlist.id && (
              <div className="flex gap-2 mt-2">
                <input
                  type="text"
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  className="border rounded px-2 py-1 flex-1"
                />
                <button
                  onClick={() => onEditSave(playlist.id)}
                  className="bg-green-600 text-white px-3 py-1 rounded flex items-center justify-center"
                  aria-label={t('playlist.save')}
                >
                  {/* Save-Icon (Checkmark) */}
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="w-5 h-5"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M5 13l4 4L19 7"
                    />
                  </svg>
                </button>
                <button
                  onClick={() => {
                    setEditName('');
                    setOptionsId(null);
                    onEdit('', '');
                  }}
                  className="bg-gray-300 px-3 py-1 rounded flex items-center justify-center"
                  aria-label={t('playlist.cancel')}
                >
                  {/* Cancel-Icon (X) */}
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="w-5 h-5"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M6 18L18 6M6 6l12 12"
                    />
                  </svg>
                </button>
              </div>
            )}
            <div className="flex items-center gap-2"></div>
            {/* Zeige nur die Items der ausgewählten Playlist an */}
            {selectedPlaylistId === playlist.id &&
              selectedPlaylist?.items &&
              selectedPlaylist.items.length > 0 &&
              (viewMode === 'list' ? (
                <DigitalItemListView
                  items={selectedPlaylist.items}
                  playlistId={playlist.id}
                  onDelete={handleDeleteItem}
                  t={t}
                  showDeleteButton={true}
                  onItemClick={(item) => onSelect(playlist.id, item)}
                />
              ) : (
                <DigitalItemGridView
                  items={selectedPlaylist.items}
                  playlistId={playlist.id}
                  t={t}
                  onDelete={handleDeleteItem}
                  showDeleteButton={true}
                  onItemClick={(item) => onSelect(playlist.id, item)}
                />
              ))}
          </li>
        ))}
      </ul>
    </>
  );
};

export default ListView;
