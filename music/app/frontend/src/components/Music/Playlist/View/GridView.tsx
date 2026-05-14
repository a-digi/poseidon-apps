import React from 'react';
import type { PlaylistIndex } from '../Playlist';
import DigitalItemGridView from './DigitalItemGridView';

interface GridViewProps {
  playlists: PlaylistIndex[];
  t: any;
  handleDeleteItem: (playlistId: string, itemId: string) => void;
}

const GridView: React.FC<GridViewProps> = ({ playlists, t, handleDeleteItem }) => (
  <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
    {playlists.map((playlist) => (
      <div key={playlist.id} className="border rounded p-3 flex flex-col gap-2 cursor-pointer">
        <span className="font-semibold text-lg mb-2">{playlist.name}</span>
        {playlist.items && playlist.items.length > 0 && (
          <DigitalItemGridView
            items={playlist.items}
            playlistId={playlist.id}
            t={t}
            onDelete={handleDeleteItem}
            showDeleteButton={true}
          />
        )}
      </div>
    ))}
  </div>
);

export default GridView;
