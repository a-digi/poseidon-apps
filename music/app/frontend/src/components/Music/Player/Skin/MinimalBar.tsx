import React from 'react';
import { formatTime } from '../MusicPlayer';
import { PlayerTrack } from '@/components/Music/Player/PlayerContext.tsx';
import { useTranslation } from 'react-i18next';

export type PlayerMinimalBarProps = {
  handleNext: () => void;
  handlePrev: () => void;
  currentTrack: PlayerTrack | null;
  progress: number;
  duration: number;
  copied: boolean;
  handleProgress: (e: React.ChangeEvent<HTMLInputElement>) => void;
  handleProgressCommit: () => void;
  handlePlayPause: () => void;
  isPlaying: boolean;
  playMode: 'playlist' | 'repeat' | 'shuffle';
  setPlayMode: React.Dispatch<React.SetStateAction<'playlist' | 'repeat' | 'shuffle'>>;
  volume: number;
  handleVolume: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onClose?: () => void;
  onDragStart?: (offsetX: number, offsetY: number) => void;
};

const playModeIcons = {
  playlist: (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <rect x="3" y="6" width="18" height="2" rx="1" />
      <rect x="3" y="12" width="14" height="2" rx="1" />
      <rect x="3" y="18" width="10" height="2" rx="1" />
    </svg>
  ),
  repeat: (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path d="M17 1l4 4-4 4" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
      <path
        d="M3 11V9a4 4 0 014-4h14"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path d="M7 23l-4-4 4-4" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
      <path
        d="M21 13v2a4 4 0 01-4 4H3"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  ),
  shuffle: (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path d="M16 3h5v5" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M4 20l16-16" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M21 16v5h-5" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M15 15l6 6" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  ),
};

const MinimalBarPlayer: React.FC<PlayerMinimalBarProps> = ({
  handleNext,
  handlePrev,
  currentTrack,
  progress,
  duration,
  handleProgress,
  handleProgressCommit,
  handlePlayPause,
  isPlaying,
  playMode,
  setPlayMode,
  volume,
  handleVolume,
  onClose,
  onDragStart,
}) => {
  const { t } = useTranslation();

  return (
    <div className="w-[350px] max-w-[400px] min-h-[40px] flex flex-col items-center px-2 py-1 bg-white/90 shadow-lg border border-gray-200 gap-1 overflow-hidden">
      {/* Title row with drag handle and close button */}
      <div className="flex items-center w-full gap-1">
        {onDragStart && (
          <button
            className="flex-shrink-0 cursor-grab active:cursor-grabbing text-gray-400 hover:text-gray-600 p-1 rounded"
            aria-label="Drag"
            onMouseDown={(e) => {
              e.preventDefault();
              onDragStart(e.clientX, e.clientY);
            }}
          >
            <svg className="w-4 h-4" viewBox="0 0 16 16" fill="currentColor">
              <circle cx="5" cy="4" r="1.2" />
              <circle cx="11" cy="4" r="1.2" />
              <circle cx="5" cy="8" r="1.2" />
              <circle cx="11" cy="8" r="1.2" />
              <circle cx="5" cy="12" r="1.2" />
              <circle cx="11" cy="12" r="1.2" />
            </svg>
          </button>
        )}
        <div className="flex-1 flex flex-col items-center min-w-0">
          <span className="font-semibold text-sm text-gray-900 truncate text-center w-full">
            {currentTrack?.title || '-'}
          </span>
          {currentTrack && 'artist' in currentTrack && currentTrack.artist && (
            <span className="text-xs text-gray-500 truncate text-center w-full">
              {currentTrack.artist}
            </span>
          )}
        </div>
        {onClose && (
          <button
            className="flex-shrink-0 text-gray-400 hover:text-gray-700 p-1 rounded hover:bg-gray-100 transition"
            aria-label="Close"
            onClick={onClose}
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        )}
      </div>

      {/* Progress bar */}
      <div className="flex items-center w-full gap-2">
        <span className="text-xs font-mono text-gray-500 w-10 text-right">
          {formatTime(progress)}
        </span>
        <input
          type="range"
          min={0}
          max={duration}
          value={progress}
          onChange={handleProgress}
          onMouseUp={handleProgressCommit}
          onTouchEnd={handleProgressCommit}
          className="flex-1 accent-blue-600 h-1 rounded-full bg-gray-200"
          disabled={duration === 0}
          style={{ appearance: 'none' }}
        />
        <span className="text-xs font-mono text-gray-500 w-10 text-left">
          {formatTime(duration)}
        </span>
      </div>

      {/* Playback controls */}
      <div className="flex items-center justify-between w-full gap-2 mt-1">
        <button
          className="rounded-full w-7 h-7 flex items-center justify-center hover:bg-blue-100 text-blue-600 transition border border-blue-200"
          aria-label="Previous"
          onClick={handlePrev}
          disabled={!currentTrack}
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <polygon points="16,18 8,12 16,6" />
          </svg>
        </button>
        <button
          onClick={currentTrack ? handlePlayPause : undefined}
          className={`rounded-full w-8 h-8 flex items-center justify-center shadow border-2 border-blue-200 transition text-xl ${!currentTrack ? 'bg-gray-200 text-gray-400 cursor-not-allowed opacity-60' : 'bg-blue-600 text-white hover:bg-blue-700'}`}
          aria-label={isPlaying ? 'Pause' : 'Play'}
          disabled={!currentTrack}
        >
          {isPlaying ? (
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <rect x="6" y="5" width="4" height="14" rx="1" />
              <rect x="14" y="5" width="4" height="14" rx="1" />
            </svg>
          ) : (
            <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
              <polygon points="8,5 20,12 8,19" />
            </svg>
          )}
        </button>
        <button
          className="rounded-full w-7 h-7 flex items-center justify-center hover:bg-blue-100 text-blue-600 transition border border-blue-200"
          aria-label="Next"
          onClick={currentTrack ? handleNext : undefined}
          disabled={!currentTrack}
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <polygon points="8,6 16,12 8,18" />
          </svg>
        </button>
        <div className="flex items-center gap-1 min-w-[60px] ml-2">
          <svg
            className="w-4 h-4 text-gray-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path d="M5 9v6h4l5 5V4l-5 5H5z" />
            <path d="M15.54 8.46a5 5 0 010 7.07" />
            <path d="M19 12a9 9 0 00-4.5-7.79" />
          </svg>
          <input
            type="range"
            min={0}
            max={1}
            step={0.01}
            value={volume}
            onChange={handleVolume}
            className="w-12 accent-blue-600 h-1 rounded-full bg-gray-200"
            style={{ appearance: 'none' }}
          />
        </div>
        <button
          className="rounded-full p-1 flex items-center justify-center hover:bg-blue-100 transition border border-blue-200 ml-2"
          aria-label={t('player.mode') || 'Mode'}
          onClick={() => {
            const next =
              playMode === 'playlist' ? 'repeat' : playMode === 'repeat' ? 'shuffle' : 'playlist';
            setPlayMode(next as 'playlist' | 'repeat' | 'shuffle');
          }}
          tabIndex={0}
        >
          {playModeIcons[playMode]}
        </button>
      </div>
    </div>
  );
};

export default MinimalBarPlayer;
