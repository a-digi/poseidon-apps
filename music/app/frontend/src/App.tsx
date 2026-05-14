import React from 'react';
import { MemoryRouter, NavLink, Navigate, Route, Routes } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Play, ListMusic, BarChart3 } from 'lucide-react';
import { PlayerProvider } from '@/components/Music/Player/PlayerContext';
import { ToastProvider } from '@/components/Notification/ToastProvider';
import { TopBarProvider, useTopBar } from '@/components/TopBar/TopBarContext';
import { SidebarProvider } from '@/components/Layout/Sidebar/SidebarContext';
import MusicPlayer from '@/components/Music/Player/MusicPlayer';
const Playlist = React.lazy(() => import('@/components/Music/Playlist/Playlist'));
const Analytics = React.lazy(() => import('@/components/Music/Analytics/Analytics').then((m) => ({ default: m.Analytics })));
import { pluginApi } from '@/api/pluginApi';
import { subscribeMode, subscribeTabActive } from '@/lib/hostBridge';

const renderTitle = (title: ReturnType<typeof useTopBar>['title']) => {
  if (!title) return null;
  if (typeof title === 'string') return <span>{title}</span>;
  return (
    <span className="flex items-center gap-2">
      {title.icon}
      <span>{title.text}</span>
    </span>
  );
};

const tabClass = ({ isActive }: { isActive: boolean }) =>
  `inline-flex items-center gap-2 px-5 py-3 text-sm border-b-2 -mb-px transition-colors ${isActive
    ? 'border-blue-600 text-blue-700 font-semibold'
    : 'border-transparent text-slate-600 hover:text-slate-900 hover:border-slate-300'
  }`;

const PluginShell: React.FC<{ children: React.ReactNode; minimal: boolean }> = ({ children, minimal }) => {
  const { title } = useTopBar();
  const { t } = useTranslation();

  if (minimal) {
    return <div className="w-full h-full">{children}</div>;
  }

  return (
    <div className="flex flex-col min-h-screen">
      <header className="bg-gradient-to-r from-slate-900 via-slate-800 to-slate-900 text-white px-6 py-3 flex items-center gap-3 shadow-md shrink-0">
        <div className="w-9 h-9 rounded-lg bg-white/10 ring-1 ring-white/10 backdrop-blur flex items-center justify-center">
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 19V6l12-2v13"
            />
            <circle cx="6" cy="18" r="3" stroke="currentColor" strokeWidth={2} fill="none" />
            <circle cx="18" cy="16" r="3" stroke="currentColor" strokeWidth={2} fill="none" />
          </svg>
        </div>
        <div className="flex-1 min-w-0">
          <div className="font-bold leading-tight">Music</div>
          <div className="text-xs text-slate-400 leading-tight">Playlists &amp; playback</div>
        </div>
        <div className="text-sm text-slate-200 font-medium truncate">{renderTitle(title)}</div>
      </header>
      <nav
        aria-label="Music sections"
        className="border-b border-slate-200 px-6 flex gap-1 shrink-0"
      >
        <NavLink to="/music/player" className={tabClass}>
          <Play className="w-4 h-4" />
          {t('menu.music.player', 'Player')}
        </NavLink>
        <NavLink to="/music/playlist" className={tabClass}>
          <ListMusic className="w-4 h-4" />
          {t('menu.music.playlist', 'Playlist')}
        </NavLink>
        <NavLink to="/music/analytics" className={tabClass}>
          <BarChart3 className="w-4 h-4" />
          {t('menu.music.analytics', 'Analytics')}
        </NavLink>
      </nav>
      <main className="flex-1 p-6 md:p-8">{children}</main>
    </div>
  );
};

const App: React.FC = () => {
  const [mode, setMode] = React.useState<'full' | 'minimal'>('full');

  React.useEffect(() => {
    pluginApi.initTables().catch(() => { });
    return subscribeMode((m) => setMode(m));
  }, []);

  React.useEffect(() => {
    const ownPluginId = new URLSearchParams(window.location.search).get('pluginId') ?? '';
    return subscribeTabActive((activePluginId) => {
      if (activePluginId === ownPluginId) setMode('full');
    });
  }, []);

  // Single stable tree — PlayerProvider and MusicPlayer are never unmounted.
  // Only PluginShell chrome and Routes are toggled based on mode so that
  // audio state and the currentTrack survive route changes in the host app.
  return (
    <MemoryRouter initialEntries={['/music/player']}>
      <PlayerProvider>
        <ToastProvider>
          <TopBarProvider>
            <SidebarProvider>
              <PluginShell minimal={mode === 'minimal'}>
                {mode === 'full' && (
                  <React.Suspense fallback={null}>
                    <Routes>
                      <Route path="/music/player" element={<></>} />
                      <Route path="/music/playlist" element={<Playlist />} />
                      <Route path="/music/analytics" element={<Analytics />} />
                      <Route path="/" element={<Navigate to="/music/player" replace />} />
                      <Route path="*" element={<Navigate to="/music/player" replace />} />
                    </Routes>
                  </React.Suspense>
                )}
                <MusicPlayer mode={mode} />
              </PluginShell>
            </SidebarProvider>
          </TopBarProvider>
        </ToastProvider>
      </PlayerProvider>
    </MemoryRouter>
  );
};

export default App;
