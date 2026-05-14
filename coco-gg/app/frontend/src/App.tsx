import { useEffect, useState } from 'react';
import {
  startPlugin,
  stopPlugin,
  getStatus,
  createRoom,
  listRooms,
  destroyRoom,
  kickPlayer,
  type RoomStatus,
  type RoomsStats,
} from './api';
import { ShareDialog } from './components/ShareDialog';
import { RoomCard } from './components/RoomCard';

type Phase = 'idle' | 'starting' | 'live' | 'error';

function statusPillLabel(p: Phase): string {
  return p === 'starting' ? 'starting…' : p;
}

function statusPillClasses(p: Phase): string {
  const base = 'px-3 py-1 rounded-full text-xs font-medium';
  if (p === 'live') return `${base} bg-green-100 text-green-800`;
  if (p === 'starting') return `${base} bg-amber-100 text-amber-800`;
  if (p === 'error') return `${base} bg-red-100 text-red-800`;
  return `${base} bg-slate-100 text-slate-700`;
}

function App() {
  const [phase, setPhase] = useState<Phase>('idle');
  const [rooms, setRooms] = useState<RoomStatus[]>([]);
  const [stats, setStats] = useState<RoomsStats>({ activeRooms: 0, totalPlayers: 0 });
  const [shareCode, setShareCode] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  async function refreshRooms() {
    const result = await listRooms('coco-gg');
    setRooms(result.rooms);
    setStats(result.stats);
  }

  async function handleStart() {
    setPhase('starting');
    setErrorMessage(null);
    try {
      await startPlugin('coco-gg');
      for (let i = 0; i < 10; i++) {
        const s = await getStatus('coco-gg');
        if (s.running) break;
        await new Promise((r) => setTimeout(r, 500));
        if (i === 9) throw new Error('plugin did not start within 5s');
      }
      setPhase('live');
    } catch (err) {
      setErrorMessage(err instanceof Error ? err.message : String(err));
      setPhase('idle');
    }
  }

  async function handleCreateRoom() {
    try {
      const { code } = await createRoom('coco-gg');
      setShareCode(code);
      refreshRooms();
    } catch (err) {
      setErrorMessage(err instanceof Error ? err.message : String(err));
    }
  }

  async function handleDestroyRoom(code: string) {
    try {
      await destroyRoom('coco-gg', code);
      if (shareCode === code) setShareCode(null);
      refreshRooms();
    } catch (err) {
      setErrorMessage(err instanceof Error ? err.message : String(err));
    }
  }

  async function handleKick(code: string, playerId: string) {
    try {
      await kickPlayer('coco-gg', code, playerId);
      refreshRooms();
    } catch (err) {
      setErrorMessage(err instanceof Error ? err.message : String(err));
    }
  }

  async function handleStop() {
    await Promise.all(rooms.map((r) => destroyRoom('coco-gg', r.code).catch(() => {})));
    await stopPlugin('coco-gg').catch(() => {});
    setPhase('idle');
    setRooms([]);
    setStats({ activeRooms: 0, totalPlayers: 0 });
    setShareCode(null);
    setErrorMessage(null);
  }

  useEffect(() => {
    if (phase !== 'live') return;
    let failures = 0;
    let cancelled = false;

    async function tick() {
      if (cancelled) return;
      try {
        await refreshRooms();
        failures = 0;
      } catch {
        failures += 1;
        if (failures >= 3) {
          setErrorMessage('Lost contact with the plugin — try Stop and Start again.');
          setPhase('error');
        }
      }
    }

    tick();
    const id = setInterval(tick, 2000);
    return () => {
      cancelled = true;
      clearInterval(id);
    };
    // refreshRooms is defined at component scope and only invokes stable state
    // setters; adding it to deps would re-run the effect every render.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [phase]);

  useEffect(() => {
    return () => {
      // Closing the tab without clicking Stop must not leak a server process.
      stopPlugin('coco-gg').catch(() => {});
    };
  }, []);

  const subline =
    phase === 'live'
      ? `${stats.activeRooms} active room${stats.activeRooms === 1 ? '' : 's'} · ${stats.totalPlayers} player${stats.totalPlayers === 1 ? '' : 's'} online`
      : phase === 'idle'
        ? 'Server is not running.'
        : phase === 'starting'
          ? 'Starting plugin…'
          : 'Server error — try Stop and Start again.';

  return (
    <div className="min-h-screen bg-slate-50">
      <header className="bg-white border-b border-slate-200 px-6 py-4 flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-900">Coco GG · Server Dashboard</h1>
          <p className="text-sm text-slate-500 mt-1">{subline}</p>
        </div>
        <span className={statusPillClasses(phase)}>{statusPillLabel(phase)}</span>
      </header>

      <main className="p-6 max-w-3xl mx-auto">
        {errorMessage && (
          <div className="mb-4 px-4 py-3 bg-red-50 border border-red-200 rounded-md text-sm text-red-700">
            {errorMessage}
          </div>
        )}

        <div className="flex gap-3 mb-6">
          {phase === 'idle' && (
            <button
              type="button"
              onClick={handleStart}
              className="px-4 py-2 rounded-md bg-blue-600 text-white text-sm font-medium hover:bg-blue-700"
            >
              Start server
            </button>
          )}
          {phase === 'live' && (
            <>
              <button
                type="button"
                onClick={handleCreateRoom}
                className="px-4 py-2 rounded-md bg-blue-600 text-white text-sm font-medium hover:bg-blue-700"
              >
                + Create room
              </button>
              <button
                type="button"
                onClick={handleStop}
                className="ml-auto px-4 py-2 rounded-md border border-slate-300 bg-white text-sm text-slate-700 hover:bg-slate-50"
              >
                Stop server
              </button>
            </>
          )}
        </div>

        {phase === 'live' && rooms.length === 0 && (
          <p className="text-sm text-slate-500 italic">
            No rooms yet — click "Create room" to get started.
          </p>
        )}

        {rooms.map((room) => (
          <RoomCard
            key={room.code}
            room={room}
            onShareQR={setShareCode}
            onDestroy={handleDestroyRoom}
            onKick={handleKick}
          />
        ))}
      </main>

      {shareCode && rooms.some((r) => r.code === shareCode) && (
        <ShareDialog roomCode={shareCode} onClose={() => setShareCode(null)} />
      )}
    </div>
  );
}

export default App;
