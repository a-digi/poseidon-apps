import React, { useMemo } from 'react';
import { Music, Play } from 'lucide-react';
import { usePlayer } from './PlayerContext';

interface Mp3Item {
  id: string;
  title: string;
  url: string;
  artist?: string;
  album?: string;
  length?: number;
  picture?: string;
}

interface PlaylistProps {
  playlistItems: Mp3Item[];
}

const formatDuration = (seconds?: number): string => {
  if (!seconds || seconds <= 0) return '—';
  const m = Math.floor(seconds / 60);
  const s = Math.floor(seconds % 60);
  return `${m}:${s.toString().padStart(2, '0')}`;
};

const sumDuration = (items: Mp3Item[]): number =>
  items.reduce((acc, item) => acc + (item.length ?? 0), 0);

const Playlist: React.FC<PlaylistProps> = ({ playlistItems }) => {
  const { setTrack, currentTrack } = usePlayer();
  const total = useMemo(() => sumDuration(playlistItems), [playlistItems]);

  if (!playlistItems.length) {
    return (
      <div className="bg-white rounded-lg border border-slate-200 py-16 text-center text-sm text-slate-500 italic">
        This playlist is empty.
      </div>
    );
  }

  const trackWord = playlistItems.length === 1 ? 'track' : 'tracks';

  return (
    <div className="bg-white rounded-lg border border-slate-200 overflow-hidden shadow-sm">
      <header className="sticky top-0 z-10 bg-white/95 backdrop-blur border-b border-slate-200 px-4 py-3 flex items-baseline gap-3">
        <h2 className="font-semibold text-slate-900">Now playing</h2>
        <span className="text-xs text-slate-500">
          {playlistItems.length} {trackWord}
          {total > 0 && ` · ${formatDuration(total)}`}
        </span>
      </header>
      <ul className="divide-y divide-slate-100">
        {playlistItems.map((item) => {
          const isActive = currentTrack?.id === item.id;
          return (
            <li
              key={item.id}
              onClick={() => setTrack(item)}
              className={`group relative flex items-center gap-4 px-4 py-3 cursor-pointer transition-colors ${
                isActive ? 'bg-blue-50' : 'hover:bg-slate-50'
              }`}
              aria-current={isActive ? 'true' : undefined}
            >
              {isActive && (
                <span className="absolute left-0 top-0 bottom-0 w-1 bg-blue-600" aria-hidden />
              )}
              <div className="relative w-14 h-14 shrink-0 rounded-md overflow-hidden bg-slate-100 ring-1 ring-slate-200/80 shadow-sm">
                {item.picture ? (
                  <img src={item.picture} alt="" className="w-full h-full object-cover" />
                ) : (
                  <div className="w-full h-full flex items-center justify-center text-slate-400">
                    <Music className="w-6 h-6" />
                  </div>
                )}
                {!isActive && (
                  <div className="absolute inset-0 bg-black/45 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                    <Play className="w-5 h-5 text-white" fill="white" />
                  </div>
                )}
              </div>
              <div className="flex-1 min-w-0">
                <p
                  className={`truncate text-sm ${
                    isActive ? 'text-blue-700 font-semibold' : 'text-slate-900 font-medium'
                  }`}
                  title={item.title}
                >
                  {item.title}
                </p>
                {item.artist && (
                  <p className="text-xs text-slate-500 truncate">{item.artist}</p>
                )}
              </div>
              <span className="text-xs text-slate-500 tabular-nums shrink-0">
                {formatDuration(item.length)}
              </span>
            </li>
          );
        })}
      </ul>
    </div>
  );
};

export default Playlist;
