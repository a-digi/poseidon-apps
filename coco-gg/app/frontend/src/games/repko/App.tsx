import { useEffect, useState } from 'react';
import { RepkoLogo } from './RepkoLogo';
import {
  createRoom,
  listRooms,
  destroyRoom,
  kickPlayer,
  startGame,
  type RoomStatus,
  type RoomsStats,
} from './api';
import type { ServerConfig } from '../../shell/api';
import { ShareDialog } from './components/ShareDialog';
import { RoomCard } from './components/RoomCard';
import { ErrorBoundary } from './components/ErrorBoundary';

interface AppProps {
  config?: ServerConfig;
}

function App({ config }: AppProps) {
  const [rooms, setRooms] = useState<RoomStatus[]>([]);
  const [stats, setStats] = useState<RoomsStats>({ activeRooms: 0, totalPlayers: 0, activeGames: 0 });
  const [shareCode, setShareCode] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  async function refreshRooms() {
    try {
      const result = await listRooms();
      console.info('[repko/App] listRooms ok', { rooms: result.rooms.length, stats: result.stats });
      setRooms(result.rooms);
      setStats(result.stats);
    } catch (err) {
      console.error('[repko/App] listRooms failed', err);
      throw err;
    }
  }

  async function handleCreateRoom() {
    try {
      const { code } = await createRoom();
      setShareCode(code);
      refreshRooms();
    } catch (err) {
      setErrorMessage(err instanceof Error ? err.message : String(err));
    }
  }

  async function handleDestroyRoom(code: string) {
    try {
      await destroyRoom(code);
      if (shareCode === code) setShareCode(null);
      refreshRooms();
    } catch (err) {
      setErrorMessage(err instanceof Error ? err.message : String(err));
    }
  }

  async function handleKick(code: string, playerId: string) {
    try {
      await kickPlayer(code, playerId);
      refreshRooms();
    } catch (err) {
      setErrorMessage(err instanceof Error ? err.message : String(err));
    }
  }

  async function handleStartGame(code: string) {
    try {
      await startGame(code);
      refreshRooms();
    } catch (err) {
      setErrorMessage(err instanceof Error ? err.message : String(err));
    }
  }

  useEffect(() => {
    console.info('[repko/App] mount');
    let cancelled = false;

    async function tick() {
      if (cancelled) return;
      try {
        await refreshRooms();
      } catch (err) {
        setErrorMessage(err instanceof Error ? err.message : String(err));
      }
    }

    tick();
    const id = setInterval(tick, 2000);
    return () => {
      cancelled = true;
      clearInterval(id);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const subline = `${stats.activeRooms} active room${stats.activeRooms === 1 ? '' : 's'} · ${stats.totalPlayers} player${stats.totalPlayers === 1 ? '' : 's'} online · ${stats.activeGames} in play`;

  return (
    <ErrorBoundary label="repko/App">
      <div>
      <div className="mb-2 flex items-center gap-2">
        <RepkoLogo className="h-7 w-7 shrink-0 text-slate-800" />
        <div>
          <h2 className="text-lg font-bold text-slate-900 leading-tight">Repko</h2>
          <p className="text-xs text-slate-500">{subline}</p>
        </div>
      </div>

      {errorMessage !== null && (
        <div className="mb-2 px-3 py-2 bg-red-50 border border-red-200 rounded text-xs text-red-700">
          {errorMessage}
        </div>
      )}

      <div className="flex gap-2 mb-3">
        <button
          type="button"
          onClick={handleCreateRoom}
          className="px-3 py-1.5 rounded bg-slate-900 text-white text-xs font-medium transition-colors hover:bg-slate-700"
        >
          + Create Instance
        </button>
      </div>

      {rooms.length === 0 && (
        <p className="text-sm text-slate-500 italic">
          No instances yet — click "Create Instance" to get started.
        </p>
      )}

      {rooms.map((room) => (
        <RoomCard
          key={room.code}
          room={room}
          onShareQR={setShareCode}
          onDestroy={handleDestroyRoom}
          onKick={handleKick}
          onStartGame={handleStartGame}
        />
      ))}

      {shareCode !== null && rooms.some((r) => r.code === shareCode) && (
        <ShareDialog roomCode={shareCode} onClose={() => setShareCode(null)} config={config} />
      )}
      </div>
    </ErrorBoundary>
  );
}

export default App;
