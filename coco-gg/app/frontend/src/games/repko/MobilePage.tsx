import { useEffect, useRef, useState } from 'react';
import { buildWsUrl } from '../../shell/api';
import { Connection } from './game/Connection';
import { GameView } from './game/GameView';
import { ErrorBoundary } from './components/ErrorBoundary';
import type { ClientAction, ErrorMsg, StateMsg } from './types';

const MAX_NAME_LENGTH = 32;

type Phase = 'name' | 'joining' | 'playing';

function MobilePage() {
  const [name, setName] = useState('');
  const [room, setRoom] = useState<string | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [phase, setPhase] = useState<Phase>('name');
  const [connection, setConnection] = useState<Connection | null>(null);
  const [state, setState] = useState<StateMsg | null>(null);
  const [myPlayerId, setMyPlayerId] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const connectionRef = useRef<Connection | null>(null);
  const phaseRef = useRef<Phase>(phase);
  const joinTimeoutRef = useRef<number | null>(null);

  useEffect(() => {
    const sp = new URLSearchParams(window.location.search);
    const r = sp.get('room');
    let t: string | null = null;
    try {
      t = window.localStorage.getItem('coco_gg_mobile_token');
    } catch {
      t = null;
    }
    console.info('[repko/MobilePage] mount', { room: r, hasToken: t !== null });
    setRoom(r);
    setToken(t);
  }, []);

  useEffect(() => {
    connectionRef.current = connection;
  }, [connection]);

  useEffect(() => {
    phaseRef.current = phase;
  }, [phase]);

  const clearJoinTimeout = () => {
    if (joinTimeoutRef.current !== null) {
      window.clearTimeout(joinTimeoutRef.current);
      joinTimeoutRef.current = null;
    }
  };

  useEffect(() => {
    return () => {
      clearJoinTimeout();
      connectionRef.current?.disconnect();
    };
  }, []);

  const canJoin = name.trim().length > 0 && room !== null && token !== null;

  const resetToName = (message: string | null) => {
    connectionRef.current?.disconnect();
    setConnection(null);
    setState(null);
    setMyPlayerId(null);
    setErrorMessage(message);
    setPhase('name');
  };

  const handleError = (msg: ErrorMsg) => {
    clearJoinTimeout();
    if (phaseRef.current === 'joining') {
      resetToName(msg.message);
      return;
    }
    if (phaseRef.current === 'playing') {
      const looksLikeKick = /kick/i.test(msg.message);
      if (looksLikeKick) {
        resetToName(msg.message || 'You were kicked from the room.');
        return;
      }
      setErrorMessage(msg.message);
    }
  };

  const handleClose = () => {
    clearJoinTimeout();
    if (phaseRef.current === 'joining') {
      resetToName('Connection lost before the server answered. Check that the plugin is running on the desktop.');
      return;
    }
    if (phaseRef.current === 'playing') {
      resetToName('Disconnected from the room.');
    }
  };

  const handleJoin = () => {
    if (!canJoin || room === null || token === null) return;
    setErrorMessage(null);
    setState(null);
    setMyPlayerId(null);
    setPhase('joining');
    const wsUrl = buildWsUrl(window.location.host, 'repko', room, token);
    console.info('[repko/MobilePage] join', { wsUrl });
    const conn = new Connection(
      wsUrl,
      {
        onWelcome: (msg) => {
          console.info('[repko/MobilePage] welcome', { playerId: msg.playerId });
          clearJoinTimeout();
          setMyPlayerId(msg.playerId);
          setPhase('playing');
        },
        onState: (msg) => {
          console.info('[repko/MobilePage] state', {
            phase: msg.phase,
            players: msg.players.length,
            hasBoard: msg.board !== undefined,
            hasYou: msg.you !== undefined,
            currentTurn: msg.currentTurn ?? null,
          });
          setState(msg);
        },
        onError: (msg) => {
          console.warn('[repko/MobilePage] error', msg);
          handleError(msg);
        },
        onClose: () => {
          console.warn('[repko/MobilePage] close (phase=' + phaseRef.current + ')');
          handleClose();
        },
      },
      { disableAutoReconnect: true },
    );
    conn.connect();
    conn.sendHello(room, name.trim());
    joinTimeoutRef.current = window.setTimeout(() => {
      joinTimeoutRef.current = null;
      if (phaseRef.current === 'joining') {
        conn.disconnect();
        resetToName('Timed out — the server did not respond in 8 seconds.');
      }
    }, 8000);
    connectionRef.current = conn;
    setConnection(conn);
  };

  const handleAbort = () => {
    clearJoinTimeout();
    resetToName('Cancelled.');
  };

  const handleLeave = () => {
    resetToName('You left the room.');
  };

  const handleAction = (action: ClientAction) => {
    connectionRef.current?.sendAction(action);
  };

  if (token === null && phase === 'name') {
    return (
      <div className="fixed inset-0 flex flex-col bg-white text-slate-900">
        <header className="border-b border-slate-200 px-6 py-4">
          <h1 className="text-xl font-semibold">Join Coco GG</h1>
        </header>
        <main className="flex flex-1 items-center justify-center px-6 py-8">
          <p className="max-w-sm text-center text-sm text-slate-600">
            Open this page from the QR code on the desktop.
          </p>
        </main>
      </div>
    );
  }

  if (phase === 'playing' && connection !== null) {
    return (
      <ErrorBoundary label="repko/MobilePage">
        <GameView
          state={state}
          myPlayerId={myPlayerId}
          onAction={handleAction}
          onLeave={handleLeave}
        />
      </ErrorBoundary>
    );
  }

  if (phase === 'joining') {
    return (
      <div className="fixed inset-0 flex flex-col items-center justify-center bg-white gap-4">
        <div className="w-10 h-10 border-4 border-slate-200 border-t-slate-700 rounded-full animate-spin" />
        <button
          type="button"
          onClick={handleAbort}
          className="text-sm text-slate-500 underline hover:text-slate-700"
        >
          Abort
        </button>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 flex flex-col bg-white text-slate-900">
      <header className="border-b border-slate-200 px-6 py-4">
        <h1 className="text-xl font-semibold">Join Repko</h1>
      </header>

      <main className="flex flex-1 flex-col items-center justify-center gap-6 px-6 py-8">
        <div className="flex flex-col items-center gap-1">
          <span className="text-xs uppercase tracking-wide text-slate-500">Room</span>
          <span className="text-3xl font-bold tracking-widest text-slate-900">
            {room ?? '(unknown room)'}
          </span>
        </div>

        {errorMessage !== null && (
          <div className="w-full max-w-sm rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
            {errorMessage}
          </div>
        )}

        <div className="flex w-full max-w-sm flex-col gap-2">
          <label htmlFor="mobile-name" className="text-sm font-medium text-slate-700">
            Your name
          </label>
          <input
            id="mobile-name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            autoFocus
            maxLength={MAX_NAME_LENGTH}
            placeholder="Your name"
            className="w-full rounded-md border border-slate-300 px-3 py-3 text-base focus:border-slate-500 focus:outline-none focus:ring-1 focus:ring-slate-500"
          />
        </div>

        <button
          type="button"
          onClick={handleJoin}
          disabled={!canJoin}
          className="w-full max-w-sm rounded-md bg-slate-900 px-4 py-3 text-base font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          Join game
        </button>
      </main>
    </div>
  );
}

export default MobilePage;
export { MobilePage };
