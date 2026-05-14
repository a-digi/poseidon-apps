import React, { useState, useRef, useEffect, useCallback } from 'react';
import Dropdown from '@/components/ui/Dropdown';
import { useTranslation } from 'react-i18next';
import Playlist from '@/components/Music/Player/Playlist';
import { Response, ResponseStatus } from '@/backend/Response';
import { usePlayer } from './PlayerContext';
import AvaPlayer from './Skin/Ava';
import MinimalBarPlayer from './Skin/MinimalBar';
import { useTopBar } from '@/components/TopBar/TopBarContext';
import { useSidebar } from '@/components/Layout/Sidebar/SidebarContext';
import { Card, CardsContainer } from '@/components/Card';
import CreateButton from '@/components/ui/Button/CreateButton';
import { useNavigate, useLocation } from 'react-router-dom';
import { releaseMinimalMode, startDrag } from '@/lib/hostBridge';

const wails = window.go?.main?.App;

export const DEFAULT_IMAGE =
  'https://upload.wikimedia.org/wikipedia/commons/thumb/a/ac/No_image_available.svg/480px-No_image_available.svg.png';

export interface Mp3Item {
  id: string;
  title: string;
  url: string;
  picture?: string;
}

interface PlaylistIndex {
  id: string;
  name: string;
}

export const formatTime = (seconds: number) => {
  const min = Math.floor(seconds / 60);
  const sec = Math.floor(seconds % 60);
  return `${min}:${sec < 10 ? '0' : ''}${sec}`;
};

interface MusicPlayerProps {
  mode: 'full' | 'minimal';
}

const MusicPlayerInner: React.FC<MusicPlayerProps> = ({ mode }) => {
  const { t } = useTranslation();
  const { setTitle } = useTopBar();
  const { setWide, wide } = useSidebar();
  const [isPlaying, setIsPlaying] = useState(false);
  const [progress, setProgress] = useState(0);
  const [duration, setDuration] = useState(0);
  const [volume, setVolume] = useState(0.7);
  const [isDragging, setIsDragging] = useState(false);
  const [cover, setCover] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const [playMode, setPlayMode] = useState<'playlist' | 'repeat' | 'shuffle'>('playlist');
  const [showModeDropdown, setShowModeDropdown] = useState(false);
  const modeDropdownRef = useRef<HTMLDivElement>(null);

  const [directFile, setDirectFile] = useState<File | null>(null);

  useEffect(() => {
    if (!showModeDropdown) return;
    function handleClick(e: MouseEvent) {
      if (modeDropdownRef.current && !modeDropdownRef.current.contains(e.target as Node)) {
        setShowModeDropdown(false);
      }
    }
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, [showModeDropdown]);

  const [playlists, setPlaylists] = useState<PlaylistIndex[]>([]);
  const [selectedPlaylistId, setSelectedPlaylistId] = useState<string>('');
  const [playlistItems, setPlaylistItems] = useState<Mp3Item[]>([]);
  const { currentTrack, setTrack } = usePlayer();
  const [audioSrc, setAudioSrc] = useState<string>('');
  const enforcePlayTimeout = useRef<number | null>(null);
  const recordedTrackIdRef = useRef<string | null>(null);
  const audioRef = useRef<HTMLAudioElement>(null);
  const restoreItemIdRef = useRef<string>('');
  const loadedRef = useRef<boolean>(false);
  const skipAutoplayRef = useRef<boolean>(false);

  useEffect(() => {
    if (!wails?.GetPlayerState) {
      loadedRef.current = true;
      return;
    }
    wails.GetPlayerState().then((data: string) => {
      try {
        const parsed = JSON.parse(data) as Response;
        if (parsed.status === ResponseStatus.Success) {
          const state = parsed.message as {
            selectedPlaylistId: string;
            playMode: 'playlist' | 'repeat' | 'shuffle';
            currentItemId: string;
          };
          if (state.selectedPlaylistId) {
            setSelectedPlaylistId(state.selectedPlaylistId);
          }
          if (state.playMode === 'playlist' || state.playMode === 'repeat' || state.playMode === 'shuffle') {
            setPlayMode(state.playMode);
          }
          restoreItemIdRef.current = state.currentItemId ?? '';
        }
      } catch {
        /* ignore — fall back to defaults */
      }
      loadedRef.current = true;
    }).catch(() => {
      loadedRef.current = true;
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (selectedPlaylistId && wails?.GetPlaylistByID) {
      wails.GetPlaylistByID(selectedPlaylistId).then((data: string) => {
        try {
          const parsed = JSON.parse(data);
          if (parsed.status === ResponseStatus.Success) {
            setPlaylistItems(parsed.message.items);
            const restoreId = restoreItemIdRef.current;
            restoreItemIdRef.current = '';
            if (restoreId !== '') {
              skipAutoplayRef.current = true;
            }
            const restored = restoreId
              ? parsed.message.items.find((it: Mp3Item) => it.id === restoreId)
              : undefined;
            if (restored) {
              setTrack(restored);
            } else if (parsed.message.items[0]) {
              setTrack(parsed.message.items[0]);
            }
            setProgress(0);
            setDuration(0);
            setIsPlaying(false);
          } else {
            setPlaylistItems([]);
            setProgress(0);
            setDuration(0);
            setIsPlaying(false);
          }
        } catch {
          setPlaylistItems([]);
          setProgress(0);
          setDuration(0);
          setIsPlaying(false);
        }
      });
    } else {
      setPlaylistItems([]);
      setProgress(0);
      setDuration(0);
      setIsPlaying(false);
    }
  }, [selectedPlaylistId, setTrack]);

  useEffect(() => {
    if (!loadedRef.current) return;
    if (!wails?.SavePlayerState) return;
    wails.SavePlayerState({
      selectedPlaylistId,
      playMode,
      currentItemId: currentTrack?.id ?? '',
    }).catch(() => {
      /* best-effort; UI shouldn't surface a save failure */
    });
  }, [selectedPlaylistId, playMode, currentTrack?.id]);

  useEffect(() => {
    if (!selectedPlaylistId) return;
    const handler = (event: Event) => {
      const detail = (event as CustomEvent<{ playlistId?: string }>).detail;
      if (detail?.playlistId !== selectedPlaylistId) return;
      if (!wails?.GetPlaylistByID) return;
      wails.GetPlaylistByID(selectedPlaylistId).then((data: string) => {
        try {
          const parsed = JSON.parse(data);
          if (parsed.status === ResponseStatus.Success) {
            setPlaylistItems(parsed.message.items ?? []);
          }
        } catch {
          /* ignore parse failure; keep current list */
        }
      });
    };
    window.addEventListener('playlist:items-changed', handler);
    return () => window.removeEventListener('playlist:items-changed', handler);
  }, [selectedPlaylistId]);

  useEffect(() => {
    if (directFile && (!currentTrack || currentTrack.id !== 'direct')) {
      setTrack({ id: 'direct', title: directFile.name, url: URL.createObjectURL(directFile) });
    }
  }, [directFile, currentTrack, setTrack]);

  useEffect(() => {
    async function fetchCover() {
      if (!currentTrack?.url) {
        setCover(null);
        return;
      }
      try {
        if (/^https?:/.test(currentTrack.url)) {
          const response = await fetch(currentTrack.url);
          const blob = await response.blob();
          const { parseBlob } = await import('music-metadata-browser');
          const metadata = await parseBlob(blob);
          const picture = metadata.common.picture?.[0];
          if (picture) {
            const base64 = URL.createObjectURL(new Blob([picture.data], { type: picture.format }));
            setCover(base64);
          } else {
            setCover(null);
          }
        } else {
          setCover(null);
        }
      } catch {
        setCover(null);
      }
    }
    fetchCover();
  }, [currentTrack]);

  const enforcePlayCurrentSelection = useCallback(
    (retries = 0) => {
      if (retries > 5) {
        return;
      }
      if (!audioRef.current) {
        if (enforcePlayTimeout.current !== null) {
          clearTimeout(enforcePlayTimeout.current);
        }
        enforcePlayTimeout.current = setTimeout(() => {
          enforcePlayCurrentSelection(retries + 1);
        }, 200);
        return;
      }
      if (currentTrack) {
        setIsPlaying(true);
      }
    },
    [currentTrack]
  );

  useEffect(() => {
    recordedTrackIdRef.current = null;
    if (!currentTrack?.url) {
      setAudioSrc('');
      return;
    }
    if (/^https?:/.test(currentTrack.url)) {
      setAudioSrc(currentTrack.url);
    } else if (wails?.GetAudioDataUrl && currentTrack.id !== 'direct') {
      wails
        .GetAudioDataUrl(currentTrack.url)
        .then((dataUrl: string) => {
          setAudioSrc(dataUrl);
        })
        .catch(() => setAudioSrc(''));
    } else if (currentTrack.id === 'direct') {
      setAudioSrc(currentTrack.url);
    } else {
      setAudioSrc('');
    }

    setProgress(0);
    setDuration(0);
    if (skipAutoplayRef.current) {
      skipAutoplayRef.current = false;
      setIsPlaying(false);
      return;
    }
    enforcePlayCurrentSelection(0);
  }, [currentTrack, enforcePlayCurrentSelection]);

  useEffect(() => {
    if (isPlaying && audioRef.current && audioSrc && currentTrack) {
      // eslint-disable-next-line @typescript-eslint/no-empty-function
      audioRef.current.play().catch(() => {});
    }
  }, [isPlaying, audioSrc, currentTrack]);

  const handlePlay = useCallback(() => {
    if (!currentTrack) return;
    if (recordedTrackIdRef.current === currentTrack.id) return;
    recordedTrackIdRef.current = currentTrack.id;
    const effectiveItemId =
      currentTrack.id === 'direct'
        ? `file:${currentTrack.title}`
        : currentTrack.id;
    wails?.RecordPlay?.({
      itemId: effectiveItemId,
      playlistId: selectedPlaylistId || undefined,
      title: currentTrack.title,
      artist: currentTrack.artist ?? '',
      album: currentTrack.album ?? '',
    })?.catch?.(() => {});
  }, [currentTrack, selectedPlaylistId]);

  const handlePlayPause = () => {
    setIsPlaying((prev) => !prev);
    if (audioRef.current) {
      if (!isPlaying) {
        audioRef.current.play();
      } else {
        audioRef.current.pause();
      }
    }
  };

  const handleNext = () => {
    if (!playlistItems.length || !currentTrack) return;
    if (playMode === 'shuffle') {
      let nextIdx = Math.floor(Math.random() * playlistItems.length);
      if (playlistItems.length > 1) {
        while (playlistItems[nextIdx].id === currentTrack.id) {
          nextIdx = Math.floor(Math.random() * playlistItems.length);
        }
      }
      setTrack(playlistItems[nextIdx]);
      setProgress(0);
      return;
    }
    const idx = playlistItems.findIndex((item) => item.id === currentTrack.id);
    let nextIdx = idx + 1;
    if (nextIdx >= playlistItems.length) {
      nextIdx = 0;
    }
    setTrack(playlistItems[nextIdx]);
    setProgress(0);
  };

  const handleEnded = () => {
    if (playMode === 'repeat') {
      if (audioRef.current) {
        audioRef.current.currentTime = 0;
        audioRef.current.play();
      }
    } else {
      handleNext();
    }
  };

  const handleProgress = (e: React.ChangeEvent<HTMLInputElement>) => {
    setIsDragging(true);
    setProgress(Number(e.target.value));
  };

  const handleProgressCommit = () => {
    setIsDragging(false);
    if (audioRef.current) {
      audioRef.current.currentTime = progress;
    }
  };

  const handleVolume = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = Number(e.target.value);
    setVolume(value);
    if (audioRef.current) {
      audioRef.current.volume = value;
    }
  };

  const handleTimeUpdate = () => {
    if (audioRef.current && !isDragging) {
      setProgress(audioRef.current.currentTime);
    }
  };

  const handleLoadedMetadata = () => {
    if (audioRef.current) {
      setDuration(audioRef.current.duration);
    }
  };

  const handleCopyFileUrl = () => {
    if (currentTrack?.url) {
      navigator.clipboard.writeText(currentTrack.url).then(() => {
        setCopied(true);
        setTimeout(() => setCopied(false), 1500);
      });
    }
  };

  const handlePrev = useCallback(() => {}, []);

  const handleDirectFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files && e.target.files[0];
    if (file) {
      setDirectFile(file);
      setTrack({ id: 'direct', title: file.name, url: URL.createObjectURL(file) });
      setSelectedPlaylistId('');
      enforcePlayCurrentSelection(0);
    }
  };

  useEffect(() => {
    setTitle({
      text: t('menu.music.player'),
      icon: (
        <svg
          className="w-5 h-5 mb-1 text-blue-600"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2" fill="none" />
          <polygon points="10,8 16,12 10,16" fill="currentColor" />
        </svg>
      ),
    });
    return () => setTitle('');
  }, [t, setTitle]);

  const navigate = useNavigate();
  const location = useLocation();
  const [showMusicPlayerUI, setShowMusicPlayerUI] = useState(false);

  useEffect(() => {
    if (location.pathname.startsWith('/music/player')) {
      setShowMusicPlayerUI(true);
      setWide(false);
      if (wails?.ListPlaylists) {
        wails.ListPlaylists().then((data: string) => {
          try {
            const parsed = JSON.parse(data);
            if (parsed.status === 'success' && Array.isArray(parsed.message)) {
              setPlaylists(parsed.message);
            }
          } catch {
            /* empty */
          }
        });
      }
    } else {
      setShowMusicPlayerUI(false);
    }
  }, [location.pathname, setWide]);

  const audio = (
    <audio
      ref={audioRef}
      src={audioSrc}
      onPlay={handlePlay}
      onTimeUpdate={handleTimeUpdate}
      onLoadedMetadata={handleLoadedMetadata}
      onEnded={handleEnded}
      className="hidden"
    />
  );

  return (
    <>
      {audio}
      {mode === 'minimal' ? (
        <MinimalBarPlayer
          handleNext={handleNext}
          handlePrev={handlePrev}
          currentTrack={currentTrack}
          progress={progress}
          duration={duration}
          copied={copied}
          handleProgress={handleProgress}
          handleProgressCommit={handleProgressCommit}
          handlePlayPause={handlePlayPause}
          isPlaying={isPlaying}
          playMode={playMode}
          setPlayMode={setPlayMode}
          volume={volume}
          handleVolume={handleVolume}
          onClose={releaseMinimalMode}
          onDragStart={startDrag}
        />
      ) : (
        <>
      {showMusicPlayerUI && (
        <CardsContainer columns="1fr 1fr 1fr" gap="1.5rem" className="mb-6 mt-0">
          <Card
            color="darkGradient"
            shadow="md"
            centered
            icon={
              <svg
                className="w-5 h-5 text-blue-100"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <rect
                  x="4"
                  y="6"
                  width="16"
                  height="12"
                  rx="3"
                  stroke="currentColor"
                  strokeWidth="2"
                  fill="none"
                />
                <path
                  d="M8 10h8M8 14h4"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                />
              </svg>
            }
            title={t('playlist.label')}
          >
            {playlists.length > 0 ? (
              <div className="w-full flex justify-center">
                <Dropdown
                  label={t('playlist.label')}
                  items={playlists.map((p) => ({ value: p.id, label: p.name }))}
                  selectedValue={selectedPlaylistId}
                  onSelect={(value) => {
                    setSelectedPlaylistId(value);
                    setDirectFile(null);
                  }}
                />
              </div>
            ) : (
              <div className="mt-1 text-sm flex flex-col gap-1 items-center">
                <span>{t('playlist.noPlaylists', 'Noch keine Playlist vorhanden.')}</span>
                <CreateButton
                  className="inline-flex items-center gap-2 px-2 py-1 text-sm font-semibold"
                  onClick={() => navigate('/music/playlist')}
                  label={t('playlist.add', 'Playlist erstellen')}
                  leftIcon={
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <rect
                        x="4"
                        y="6"
                        width="16"
                        height="12"
                        rx="3"
                        stroke="currentColor"
                        strokeWidth="2"
                        fill="none"
                      />
                      <path
                        d="M8 10h8M8 14h4"
                        stroke="currentColor"
                        strokeWidth="2"
                        strokeLinecap="round"
                      />
                    </svg>
                  }
                />
              </div>
            )}
          </Card>
          <Card
            color="darkGradient"
            shadow="md"
            centered
            icon={
              <svg
                className="w-5 h-5 text-gray-10"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  d="M12 4v16m8-8H4"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            }
            title={t('storage.chooseTargetFolder', 'Datei öffnen')}
          >
            <CreateButton
              type="button"
              className="px-3 py-1 text-sm font-semibold"
              onClick={() => document.getElementById('direct-file-input')?.click()}
              label={t('storage.chooseTargetFolder', 'Datei auswählen')}
            />
            <input
              id="direct-file-input"
              type="file"
              accept="audio/*"
              style={{ display: 'none' }}
              onChange={handleDirectFileChange}
            />
            {directFile && (
              <span className="text-xs truncate max-w-xs mt-2">{directFile.name}</span>
            )}
          </Card>
          {selectedPlaylistId && (
            <Card
              color="darkGradient"
              shadow="md"
              centered
              icon={
                <svg
                  className="w-5 h-5 text-white"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <rect
                    x="4"
                    y="6"
                    width="16"
                    height="12"
                    rx="3"
                    stroke="currentColor"
                    strokeWidth="2"
                    fill="none"
                  />
                  <path
                    d="M8 10h8M8 14h4"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                  />
                </svg>
              }
              title={t('playlist.selected', 'Ausgewählte Playlist')}
            >
              <span className="text-lg font-bold text-white">
                {playlists.find((p) => p.id === selectedPlaylistId)?.name || ''}
              </span>
              <span className="text-2xl font-bold text-white mt-2">{playlistItems.length}</span>
              <span className="text-xs text-gray-200 mt-1">
                {t('playlist.items', 'Songs in Playlist')}
              </span>
            </Card>
          )}
        </CardsContainer>
      )}
      {showMusicPlayerUI && (
        <div className="w-full flex flex-col items-center justify-start">
          <div
            className={`w-full ${wide ? 'max-w-[1000px]' : 'max-w-[1140px]'} flex flex-col items-stretch`}
          >
            <div
              className="flex flex-row auto-cols-auto gap-8 items-start"
              style={{ height: '550px' }}
            >
              <div className="flex-shrink-0">
                {currentTrack && (
                  selectedPlaylistId && playlistItems.length > 0 && !directFile ? (
                    <AvaPlayer
                      handleNext={handleNext}
                      currentTrack={currentTrack}
                      cover={cover}
                      progress={progress}
                      duration={duration}
                      copied={copied}
                      handleProgress={handleProgress}
                      handleProgressCommit={handleProgressCommit}
                      handlePlayPause={handlePlayPause}
                      isPlaying={isPlaying}
                      modeDropdownRef={modeDropdownRef}
                      setShowModeDropdown={setShowModeDropdown}
                      playMode={playMode}
                      showModeDropdown={showModeDropdown}
                      setPlayMode={setPlayMode}
                      volume={volume}
                      handleVolume={handleVolume}
                      handleCopyFileUrl={handleCopyFileUrl}
                    />
                  ) : (
                    <MinimalBarPlayer
                      handleNext={handleNext}
                      handlePrev={handlePrev}
                      currentTrack={currentTrack}
                      progress={progress}
                      duration={duration}
                      copied={copied}
                      handleProgress={handleProgress}
                      handleProgressCommit={handleProgressCommit}
                      handlePlayPause={handlePlayPause}
                      isPlaying={isPlaying}
                      playMode={playMode}
                      setPlayMode={setPlayMode}
                      volume={volume}
                      handleVolume={handleVolume}
                    />
                  )
                )}
              </div>
              <div className="flex-1 min-w-0 h-full overflow-y-auto">
                {selectedPlaylistId && playlistItems.length > 0 && !directFile && (
                  <Playlist playlistItems={playlistItems} />
                )}
              </div>
            </div>
          </div>
        </div>
      )}
        </>
      )}
    </>
  );
};

const MusicPlayer: React.FC<MusicPlayerProps> = ({ mode }) => <MusicPlayerInner mode={mode} />;

export default MusicPlayer;
