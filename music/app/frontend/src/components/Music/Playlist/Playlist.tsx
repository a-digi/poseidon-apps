import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Response, ResponseStatus } from '@/backend/Response';
import ModalDialog from '@/components/Modal/ModalDialog';
import { useToast } from '@/components/Notification/ToastProvider';
import ListView from './View/ListView';
import { useTopBar } from '@/components/TopBar/TopBarContext';
import CreateButton from '@/components/ui/Button/CreateButton';

const wails = (window.go as { main?: { App?: any } })?.main?.App;

export type DigitalItem = {
  id: string;
  title: string;
  url: string;
  artist?: string;
  album?: string;
  genre?: string;
  year?: number;
  track?: number;
  length?: number;
  picture?: string;
  mimeType?: string;
};

export type Playlist = DigitalItem[];

export type PlaylistIndex = {
  id: string;
  name: string;
  items?: DigitalItem[];
};

const initialPlaylists: PlaylistIndex[] = [];

const Playlist: React.FC = () => {
  const { t } = useTranslation();
  const { addToast } = useToast();
  const { setTitle } = useTopBar();
  const [playlists, setPlaylists] = useState<PlaylistIndex[]>(initialPlaylists);
  const [selectedPlaylistId, setSelectedPlaylistId] = useState<string | null>(null);
  const [newName, setNewName] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editName, setEditName] = useState('');
  const [showAdd, setShowAdd] = useState(false);
  const [optionsId, setOptionsId] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string>('');

  useEffect(() => {
    setTitle({
      text: t('playlist.title'),
      icon: (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="w-6 h-6 text-green-700"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 19V6l12-2v13"
          />
          <circle cx="6" cy="18" r="3" />
          <circle cx="18" cy="16" r="3" />
        </svg>
      ),
    });
  }, [t, setTitle]);

  useEffect(() => {
    if (wails?.ListPlaylists) {
      wails
        .ListPlaylists()
        .then((data: string) => {
          let parsed: Response;
          try {
            parsed = JSON.parse(data);
          } catch (e) {
            addToast({ message: t('playlist.parseError'), type: 'error' });
            return;
          }
          if (parsed.status === ResponseStatus.Success) {
            setPlaylists(Array.isArray(parsed.message) ? parsed.message : []);
          } else {
            addToast({ message: parsed.message || t('playlist.backendError'), type: 'error' });
          }
        })
        .catch(() => {
          addToast({ message: t('playlist.listError'), type: 'error' });
        });
    }
  }, [addToast, t]);

  const handleAdd = async () => {
    if (newName.trim() && wails?.CreatePlaylist) {
      wails
        .CreatePlaylist(newName.trim())
        .then((created: string) => {
          let parsed: Response = {} as Response;
          console.log(created);
          try {
            parsed = JSON.parse(created) as Response;
          } catch (e) {
            addToast({ message: t('playlist.createError'), type: 'error' });
            return;
          }
          if (parsed.status === ResponseStatus.Success) {
            const playlist: PlaylistIndex = {
              id: parsed.message.id,
              name: parsed.message.name,
            };
            setPlaylists((prev) => [...prev, playlist]);
            addToast({
              message: t('playlist.createSuccess', { title: playlist.name }),
              type: 'success',
            });
          } else {
            addToast({ message: parsed.message || t('playlist.createError'), type: 'error' });
          }
          setNewName('');
        })
        .catch(() => {
          addToast({ message: t('playlist.createError'), type: 'error' });
        });
    }
  };

  const handleDelete = async (id: string) => {
    const playlist = playlists.find((p) => p.id === id);
    const playlistName = playlist?.name || '';
    if (wails?.DeletePlaylist) {
      try {
        const ok = await wails.DeletePlaylist(id);
        if (ok) {
          setPlaylists(playlists.filter((p) => p.id !== id));
          addToast({
            message: t('playlist.deleteSuccessWithTitle', { title: playlistName }),
            type: 'success',
          });
        } else {
          addToast({ message: t('playlist.deleteError'), type: 'error' });
        }
      } catch (err) {
        addToast({ message: t('playlist.deleteError'), type: 'error' });
      }
    }
  };

  const handleEdit = (id: string, name: string) => {
    setEditingId(id);
    setEditName(name);
  };

  const handleEditSave = async (id: string) => {
    if (wails?.EditPlaylist) {
      try {
        const ok = await wails.EditPlaylist(id, editName);
        if (ok) setPlaylists(playlists.map((p) => (p.id === id ? { ...p, name: editName } : p)));
        setEditingId(null);
        setEditName('');
      } catch (err) {
        addToast({ message: t('playlist.editError'), type: 'error' });
      }
    }
  };

  const handleSelectPlaylist = async (id: string) => {
    setSelectedPlaylistId(id);
    if (wails?.GetPlaylistByID) {
      try {
        const data = await wails.GetPlaylistByID(id);
        let parsed: Response;
        try {
          parsed = JSON.parse(data);
        } catch (e) {
          addToast({ message: t('playlist.loadError'), type: 'error' });
          return;
        }
        if (parsed.status === ResponseStatus.Success) {
          const items = Array.isArray(parsed.message.items) ? parsed.message.items : [];
          setPlaylists((prev) =>
            prev.map((p) => (p.id === id ? { ...p, items: items as DigitalItem[] } : p))
          );
        } else {
          addToast({ message: parsed.message, type: 'error' });
        }
      } catch (err) {
        addToast({ message: t('playlist.loadError'), type: 'error' });
      }
    } else {
      addToast({ message: t('playlist.getByIdUnavailable'), type: 'error' });
    }
  };

  const handleAddItem = async (playlistId: string, file?: File, filePath?: string) => {
    let itemTitle = '';
    if (filePath && wails?.AddPlaylistItem) {
      try {
        const fileName = filePath.split(/[\\/]/).pop() || filePath;
        itemTitle = fileName;
        const item: DigitalItem = { id: Date.now().toString(), title: fileName, url: filePath };
        const ok: string = await wails.AddPlaylistItem(playlistId, item);
        const response = JSON.parse(ok) as Response;
        if (response.status === ResponseStatus.Success) {
          addToast({
            message: t('playlist.addMp3SuccessWithTitle', { title: itemTitle }),
            type: 'success',
          });
          window.dispatchEvent(new CustomEvent('playlist:items-changed', { detail: { playlistId } }));
          await handleSelectPlaylist(playlistId);
        } else {
          addToast({ message: response.message || t('playlist.addMp3Error'), type: 'error' });
        }
      } catch (err) {
        addToast({ message: t('playlist.addMp3Error'), type: 'error' });
      }
      return;
    }
    if (file && file.name && wails?.AddPlaylistItem) {
      try {
        itemTitle = file.name;
        const item: DigitalItem = { id: Date.now().toString(), title: file.name, url: file.name };
        const ok = await wails.AddPlaylistItem(playlistId, item);
        const response = JSON.parse(ok) as Response;
        if (response.status === ResponseStatus.Success) {
          addToast({
            message: t('playlist.addMp3SuccessWithTitle', { title: itemTitle }),
            type: 'success',
          });
          window.dispatchEvent(new CustomEvent('playlist:items-changed', { detail: { playlistId } }));
          await handleSelectPlaylist(playlistId);
        } else {
          addToast({ message: response.message || t('playlist.addMp3Error'), type: 'error' });
        }
      } catch (err) {
        addToast({ message: t('playlist.addMp3Error'), type: 'error' });
      }
    }
  };

  const handleDeleteItem = async (playlistId: string, itemId: string) => {
    const playlist = playlists.find((p) => p.id === playlistId);
    const item = playlist?.items?.find((i) => i.id === itemId);
    const itemTitle = item?.title || '';
    if (wails?.DeletePlaylistItem) {
      try {
        const result = await wails.DeletePlaylistItem(playlistId, itemId);
        let response: Response;
        try {
          response = JSON.parse(result);
        } catch (e) {
          addToast({ message: t('playlist.deleteMp3Error'), type: 'error' });
          return;
        }
        if (response.status === ResponseStatus.Success) {
          addToast({
            message: t('playlist.deleteMp3SuccessWithTitle', { title: itemTitle }),
            type: 'success',
          });
          window.dispatchEvent(new CustomEvent('playlist:items-changed', { detail: { playlistId } }));
          setPlaylists((prev) =>
            prev.map((p) =>
              p.id === playlistId
                ? { ...p, items: p.items ? p.items.filter((item) => item.id !== itemId) : [] }
                : p
            )
          );
        } else {
          addToast({ message: response.message || t('playlist.deleteMp3Error'), type: 'error' });
        }
      } catch (err) {
        addToast({ message: t('playlist.deleteMp3Error'), type: 'error' });
      }
    }
  };

  return (
    <div className="w-full mx-auto p-6">
      <ModalDialog
        open={showAdd}
        onClose={() => setShowAdd(false)}
        title={t('playlist.addTitle', 'Neue Playlist erstellen')}
      >
        <form
          onSubmit={async (e) => {
            e.preventDefault();
            await handleAdd();
            setShowAdd(false);
          }}
          className="w-full p-4 flex gap-2"
        >
          <input
            type="text"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            placeholder={t('playlist.new')}
            className="border rounded px-2 py-1 flex-1"
            required
            autoFocus
          />
          <CreateButton type="submit" label={t('playlist.add')} />
        </form>
      </ModalDialog>
      <ListView
        playlists={playlists}
        selectedPlaylistId={selectedPlaylistId}
        onSelect={handleSelectPlaylist}
        onAddItem={handleAddItem}
        onDelete={handleDelete}
        onEdit={handleEdit}
        onEditSave={handleEditSave}
        editingId={editingId}
        editName={editName}
        setEditName={setEditName}
        optionsId={optionsId}
        setOptionsId={setOptionsId}
        t={t}
        handleDeleteItem={handleDeleteItem}
        addToast={addToast}
        addNewItemComponent={
          <CreateButton
            className="ml-auto flex items-center gap-1"
            style={{ minWidth: 'auto' }}
            onClick={() => setShowAdd((v) => !v)}
            aria-label={t('playlist.add')}
            label={t('playlist.add')}
          />
        }
      />
      <ModalDialog
        open={!!errorMessage}
        onClose={() => setErrorMessage('')}
        errorMessage={errorMessage}
        title={t('fileDialog.invalidTypeTitle')}
      />
    </div>
  );
};

export default Playlist;
