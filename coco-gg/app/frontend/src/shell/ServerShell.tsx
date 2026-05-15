import { useEffect, useState, Suspense, lazy } from 'react';
import { Play, Square } from 'lucide-react';
import { startPlugin, stopPlugin, getStatus } from './api';
import { GAMES } from './games-registry';
import { GameCard } from './GameCard';

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

interface ActiveGameSlotProps {
  gameId: string;
}

function ActiveGameSlot({ gameId }: ActiveGameSlotProps) {
  const descriptor = GAMES.find((g) => g.id === gameId);
  if (descriptor === undefined) {
    return <p className="text-sm text-slate-500 italic">Unknown game.</p>;
  }
  const GameApp = lazy(descriptor.loadApp);
  return (
    <Suspense
      fallback={
        <div className="flex items-center justify-center py-12">
          <div className="w-8 h-8 border-4 border-slate-200 border-t-blue-500 rounded-full animate-spin" />
        </div>
      }
    >
      <GameApp />
    </Suspense>
  );
}

function ServerShell() {
  const [phase, setPhase] = useState<Phase>('idle');
  const [activeGameId, setActiveGameId] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  async function handleStart() {
    setPhase('starting');
    setErrorMessage(null);
    try {
      await startPlugin();
      for (let i = 0; i < 10; i++) {
        const s = await getStatus();
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

  async function handleStop() {
    await stopPlugin().catch(() => { });
    setPhase('idle');
    setActiveGameId(null);
    setErrorMessage(null);
  }

  useEffect(() => {
    return () => {
      stopPlugin().catch(() => { });
    };
  }, []);

  return (
    <div className="min-h-screen bg-slate-50">
      <header className="bg-white border-b border-slate-200 px-3 py-2 flex items-center gap-3">
        {activeGameId !== null && (
          <button
            type="button"
            onClick={() => setActiveGameId(null)}
            className="text-sm text-slate-600 hover:text-slate-900"
          >
            ← Games
          </button>
        )}
        <div className="flex-1">
          <h1 className="text-xl font-bold text-slate-900">Coco GG · Server Dashboard</h1>
          <p className="text-xs text-slate-500">
            {phase === 'idle'
              ? 'Server is not running.'
              : phase === 'starting'
                ? 'Starting plugin…'
                : phase === 'live'
                  ? 'Server running.'
                  : 'Server error — try Stop and Start again.'}
          </p>
        </div>
        <span className={statusPillClasses(phase)}>{statusPillLabel(phase)}</span>
        {phase === 'idle' && (
          <button
            type="button"
            onClick={handleStart}
            className="flex items-center gap-1.5 rounded-md bg-slate-900 px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-slate-700"
          >
            <Play className="h-4 w-4" />
            Start server
          </button>
        )}
        {phase === 'live' && (
          <button
            type="button"
            onClick={handleStop}
            className="flex items-center gap-1.5 rounded-md border border-red-200 bg-red-50 px-3 py-1.5 text-xs font-medium text-red-700 transition-colors hover:bg-red-100"
          >
            <Square className="h-4 w-4 fill-red-600 text-red-600" />
            Stop server
          </button>
        )}
      </header>

      <main className="max-w-full p-2">
        {errorMessage !== null && (
          <div className="mb-2 px-3 py-2 bg-red-50 border border-red-200 rounded text-xs text-red-700">
            {errorMessage}
          </div>
        )}

        {phase !== 'live' && (
          <p className="text-sm text-slate-500 italic">Start the server to play.</p>
        )}

        {phase === 'live' &&
          activeGameId === null &&
          (GAMES.length === 0 ? (
            <p className="text-sm text-slate-500 italic">No games available.</p>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
              {GAMES.map((g, i) => (
                <GameCard key={g.id} game={g} index={i} onOpen={() => setActiveGameId(g.id)} />
              ))}
            </div>
          ))}

        {phase === 'live' && activeGameId !== null && <ActiveGameSlot gameId={activeGameId} />}
      </main>
    </div>
  );
}

export default ServerShell;
