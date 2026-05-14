import React from 'react';
import { DEFAULT_IMAGE, formatTime } from '../MusicPlayer';
import { PlayerTrack } from '@/components/Music/Player/PlayerContext.tsx';
import { useTranslation } from 'react-i18next';

export const playModeOptions = [
  {
    value: 'playlist',
    icon: (
      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <rect x="3" y="6" width="18" height="2" rx="1" />
        <rect x="3" y="12" width="14" height="2" rx="1" />
        <rect x="3" y="18" width="10" height="2" rx="1" />
      </svg>
    ),
  },
  {
    value: 'repeat',
    icon: (
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
  },
  {
    value: 'shuffle',
    icon: (
      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path d="M16 3h5v5" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
        <path d="M4 20l16-16" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
        <path d="M21 16v5h-5" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
        <path d="M15 15l6 6" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
      </svg>
    ),
  },
];

export type PlayerAvaMinimalProps = {
  audioRef: React.RefObject<HTMLAudioElement>;
  audioSrc: string;
  handleTimeUpdate: () => void;
  handleLoadedMetadata: () => void;
  handleEnded: () => void;
  handleNext: () => void;
  currentTrack: PlayerTrack | null;
  cover: string | null;
  handleCopyFileUrl: () => void;
  progress: number;
  duration: number;
  copied: boolean;
  handleProgress: (e: React.ChangeEvent<HTMLInputElement>) => void;
  handleProgressCommit: () => void;
  handlePlayPause: () => void;
  isPlaying: boolean;
  modeDropdownRef: React.RefObject<HTMLDivElement>;
  setShowModeDropdown: React.Dispatch<React.SetStateAction<boolean>>;
  playMode: 'playlist' | 'repeat' | 'shuffle';
  showModeDropdown: boolean;
  setPlayMode: React.Dispatch<React.SetStateAction<'playlist' | 'repeat' | 'shuffle'>>;
  volume: number;
  handleVolume: (e: React.ChangeEvent<HTMLInputElement>) => void;
};

const AvaMinimalPlayer: React.FC<PlayerAvaMinimalProps> = ({
  audioRef,
  audioSrc,
  handleTimeUpdate,
  handleLoadedMetadata,
  handleEnded,
  handleNext,
  currentTrack,
  cover,
  progress,
  duration,
  copied,
  handleProgress,
  handleProgressCommit,
  handlePlayPause,
  isPlaying,
  modeDropdownRef,
  setShowModeDropdown,
  playMode,
  showModeDropdown,
  setPlayMode,
  volume,
  handleVolume,
}) => {
  const { t } = useTranslation();

  return (
    <div className="w-full min-w-[400px] max-w-[400px] mx-auto rounded-2xl shadow-2xl bg-gradient-to-br from-gray-50 via-white to-gray-100 p-8 flex flex-col items-center gap-6 border border-gray-200">
      {/* Audio-Element: alle relevanten Props werden verwendet */}
      <audio
        ref={audioRef}
        src={audioSrc}
        onTimeUpdate={handleTimeUpdate}
        onLoadedMetadata={handleLoadedMetadata}
        onEnded={handleEnded}
        className="hidden"
      />
      <div className="flex flex-col items-center w-full">
        <div className="relative mb-6">
          <img
            src={currentTrack?.picture || cover || DEFAULT_IMAGE}
            alt="Album Art"
            className="w-40 h-40 rounded-2xl object-cover shadow-xl border-4 border-white"
          />
        </div>
        <div className="flex flex-col items-center w-full mb-4">
          <span className="font-bold text-md text-gray-800 tracking-wide text-center">
            {currentTrack?.title || '-'}
          </span>
          {currentTrack && (currentTrack as any).artist && (
            <span className="text-base text-gray-600 font-normal mt-1 text-center">
              {(currentTrack as any).artist}
            </span>
          )}
        </div>
        <div className="w-full flex flex-col gap-2 mb-2">
          <div className="flex items-center justify-between text-xs text-gray-500 mb-1 font-mono">
            <span>{formatTime(progress)}</span>
            <span>{formatTime(duration)}</span>
          </div>
          <input
            type="range"
            min={0}
            max={duration}
            value={progress}
            onChange={handleProgress}
            onMouseUp={handleProgressCommit}
            onTouchEnd={handleProgressCommit}
            className="w-full accent-blue-600 h-2 rounded-full bg-gray-300"
            disabled={duration === 0}
            style={{ appearance: 'none' }}
          />
        </div>
        {copied && (
          <div className="text-xs text-green-600 mt-1">{t('player.urlCopied', 'URL kopiert!')}</div>
        )}
        <div className="flex items-center justify-center gap-8 w-full mt-4">
          <button
            className="bg-white/80 rounded-full w-12 h-12 flex items-center justify-center shadow hover:bg-blue-100 text-blue-600 transition border border-blue-200"
            aria-label="Previous"
            onClick={handleNext}
            disabled={!currentTrack}
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <polygon points="16,18 8,12 16,6" />
            </svg>
          </button>
          <button
            onClick={currentTrack ? handlePlayPause : undefined}
            className={`rounded-full w-16 h-16 flex items-center justify-center shadow-xl border-2 border-blue-200 transition text-3xl ${!currentTrack ? 'bg-gray-200 text-gray-400 cursor-not-allowed opacity-60' : 'bg-blue-600 text-white hover:bg-blue-700'}`}
            aria-label={isPlaying ? 'Pause' : 'Play'}
            disabled={!currentTrack}
          >
            {isPlaying ? (
              <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <rect x="6" y="5" width="4" height="14" rx="1" />
                <rect x="14" y="5" width="4" height="14" rx="1" />
              </svg>
            ) : (
              <svg className="w-8 h-8" fill="currentColor" viewBox="0 0 24 24">
                <polygon points="8,5 20,12 8,19" />
              </svg>
            )}
          </button>
          <button
            className="bg-white/80 rounded-full w-12 h-12 flex items-center justify-center shadow hover:bg-blue-100 text-blue-600 transition border border-blue-200"
            aria-label="Next"
            onClick={currentTrack ? handleNext : undefined}
            disabled={!currentTrack}
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <polygon points="8,6 16,12 8,18" />
            </svg>
          </button>
        </div>
        <div className="flex items-center justify-between w-full mt-6 gap-4">
          <div className="flex items-center gap-2">
            <svg
              className="w-6 h-6 text-gray-400"
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
              className="w-32 accent-blue-600 h-1 rounded-full bg-gray-300"
              style={{ appearance: 'none' }}
            />
          </div>
          <div className="flex flex-col items-center relative" ref={modeDropdownRef}>
            <div className="relative">
              <button
                className="rounded-full bg-white/80 p-2 flex items-center justify-center hover:bg-blue-100 transition shadow border border-blue-200"
                aria-label={t('player.mode') || 'Mode'}
                onClick={() => setShowModeDropdown((v) => !v)}
                tabIndex={0}
              >
                {playModeOptions.find((opt) => opt.value === playMode)?.icon}
              </button>
              {showModeDropdown && (
                <div
                  className="absolute left-1/2 -translate-x-1/2 mt-2 z-10 flex flex-col bg-white border border-blue-200 rounded shadow-lg min-w-[40px]"
                  onMouseEnter={() => setShowModeDropdown(true)}
                  onMouseLeave={() => setShowModeDropdown(false)}
                >
                  {playModeOptions.map((opt) => (
                    <button
                      key={opt.value}
                      className={`flex items-center justify-center p-2 hover:bg-blue-100 transition ${playMode === opt.value ? 'text-green-600' : 'text-blue-600'}`}
                      onClick={() => {
                        setPlayMode(opt.value as 'playlist' | 'repeat' | 'shuffle');
                        setShowModeDropdown(false);
                      }}
                      aria-label={opt.value}
                      tabIndex={0}
                    >
                      {opt.icon}
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AvaMinimalPlayer;
