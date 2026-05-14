import { useEffect, useRef, useState } from 'react';
import { buildWsUrl, createMobileSession, getStatus, startPlugin, stopPlugin } from './api';
import { Connection } from './game/net/Connection';
import { PhaserGame } from './game/PhaserGame';
import { ShareDialog } from './components/ShareDialog';

const MAX_NAME_LENGTH = 32;

type Phase = 'idle' | 'starting' | 'hosting';

const PHASE_LABEL: Record<Phase, string> = {
  idle: 'idle',
  starting: 'starting…',
  hosting: 'hosting',
};

function App() {
  const [name, setName] = useState('');
  const [phase, setPhase] = useState<Phase>('idle');
  const [roomCode, setRoomCode] = useState<string | null>(null);
  const [connection, setConnection] = useState<Connection | null>(null);
  const [showShareDialog, setShowShareDialog] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const connectionRef = useRef<Connection | null>(null);

  useEffect(() => {
    connectionRef.current = connection;
  }, [connection]);

  useEffect(() => {
    return () => {
      connectionRef.current?.disconnect();
      // Closing the tab without clicking Stop must not leak a server process.
      stopPlugin('coco-gg').catch(() => {});
    };
  }, []);

  const canStart = name.trim().length > 0 && phase === 'idle';

  const cleanup = () => {
    connectionRef.current?.disconnect();
    // Best-effort: a stop failure shouldn't block the UI; the plugin will be cleaned up at app shutdown.
    stopPlugin('coco-gg').catch(() => {});
    connectionRef.current = null;
    setConnection(null);
    setRoomCode(null);
    setShowShareDialog(false);
    setPhase('idle');
  };

  const handleStart = async () => {
    setErrorMessage(null);
    setPhase('starting');
    try {
      await startPlugin('coco-gg');

      for (let i = 0; i < 10; i++) {
        const s = await getStatus('coco-gg');
        if (s.running) break;
        await new Promise((res) => setTimeout(res, 500));
        if (i === 9) throw new Error('plugin did not start within 5s');
      }

      const session = await createMobileSession('coco-gg');
      const wsUrl = buildWsUrl('localhost:2014', '', session.token);
      const conn = new Connection(wsUrl, {
        onWelcome: (msg) => {
          setRoomCode(msg.room);
          setPhase('hosting');
          setShowShareDialog(true);
        },
        onError: (msg) => {
          setErrorMessage(msg.message);
          cleanup();
        },
        onClose: () => {
          /* PhaserGame swaps callbacks once mounted; until then we just drop. */
        },
      });
      conn.connect();
      conn.sendHello('', name.trim());
      connectionRef.current = conn;
      setConnection(conn);
    } catch (err) {
      setErrorMessage(err instanceof Error ? err.message : String(err));
      setPhase('idle');
    }
  };

  const handleStop = () => {
    cleanup();
  };

  return (
    <div className="min-h-screen bg-white text-slate-900">
      <header className="flex items-center justify-between border-b border-slate-200 px-6 py-4">
        <h1 className="text-xl font-semibold">Coco GG</h1>
        <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-medium text-slate-600">
          {PHASE_LABEL[phase]}
        </span>
      </header>

      <main className="mx-auto flex max-w-3xl flex-col gap-6 px-6 py-8">
        {errorMessage !== null && (
          <div className="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
            {errorMessage}
          </div>
        )}

        {phase === 'idle' && (
          <>
            <div className="flex flex-col gap-2">
              <label htmlFor="player-name" className="text-sm font-medium text-slate-700">
                Your name
              </label>
              <input
                id="player-name"
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                maxLength={MAX_NAME_LENGTH}
                placeholder="Your name"
                className="w-full max-w-sm rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-slate-500 focus:outline-none focus:ring-1 focus:ring-slate-500"
              />
            </div>

            <div>
              <button
                type="button"
                onClick={handleStart}
                disabled={!canStart}
                className="rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
              >
                Start game
              </button>
            </div>
          </>
        )}

        {phase === 'starting' && (
          <div className="flex items-center gap-3 text-sm text-slate-600">
            <div className="w-5 h-5 border-2 border-slate-200 border-t-slate-600 rounded-full animate-spin" />
            Starting…
          </div>
        )}

        {phase === 'hosting' && connection !== null && (
          <div className="flex flex-col gap-4">
            <div className="flex items-center gap-3">
              <button
                type="button"
                onClick={() => setShowShareDialog(true)}
                className="rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white"
              >
                Share
              </button>
              <button
                type="button"
                onClick={handleStop}
                className="rounded-md border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50"
              >
                Stop
              </button>
              {roomCode !== null && (
                <span className="text-sm text-slate-600">
                  Room <span className="font-mono font-semibold text-slate-900">{roomCode}</span>
                </span>
              )}
            </div>

            <PhaserGame connection={connection} mode="desktop" />
          </div>
        )}
      </main>

      {showShareDialog && roomCode !== null && (
        <ShareDialog roomCode={roomCode} onClose={() => setShowShareDialog(false)} />
      )}
    </div>
  );
}

export default App;
