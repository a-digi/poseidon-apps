import React, { createContext, useContext, useState, useCallback } from 'react';

export interface PlayerTrack {
  id: string;
  title: string;
  url: string;
  artist?: string;
  album?: string;
  length?: number;
  picture?: string;
}

interface PlayerContextType {
  currentTrack: PlayerTrack | null;
  setTrack: (track: PlayerTrack) => void;
  isPlaying: boolean;
  setIsPlaying: (playing: boolean) => void;
  playlist: PlayerTrack[];
  setPlaylist: (tracks: PlayerTrack[]) => void;
}

const PlayerContext = createContext<PlayerContextType | undefined>(undefined);

export const PlayerProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [currentTrack, setCurrentTrack] = useState<PlayerTrack | null>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [playlist, setPlaylist] = useState<PlayerTrack[]>([]);

  const setTrack = useCallback((track: PlayerTrack) => {
    setCurrentTrack(track);
    setIsPlaying(true);
  }, []);

  return (
    <PlayerContext.Provider
      value={{ currentTrack, setTrack, isPlaying, setIsPlaying, playlist, setPlaylist }}
    >
      {children}
    </PlayerContext.Provider>
  );
};

export function usePlayer() {
  const ctx = useContext(PlayerContext);
  if (!ctx) throw new Error('usePlayer must be used within a PlayerProvider');
  return ctx;
}
