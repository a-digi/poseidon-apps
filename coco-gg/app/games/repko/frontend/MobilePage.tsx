import { useEffect, useRef, useState } from 'react';
import { buildWsUrl } from '@shell/api';
import { Connection } from './game/Connection';
import { GameView } from './game/GameView';
import { ErrorBoundary } from './components/ErrorBoundary';
import { leaveRoom } from './api';
import type { ClientAction, ErrorMsg, StateMsg } from './types';

const MAX_NAME_LENGTH = 32;

type Phase = 'name' | 'joining' | 'playing' | 'disconnected';

const resumeKey = (room: string) => `repko_resume_${room}`;

function MobilePage() {
  const [name, setName] = useState('');
  const [room, setRoom] = useState<string | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [resume, setResume] = useState<string | null>(null);
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
    let storedResume: string | null = null;
    if (r !== null) {
      try {
        storedResume = window.localStorage.getItem(resumeKey(r));
      } catch {
        storedResume = null;
      }
    }
    console.info('[repko/MobilePage] mount', { room: r, hasToken: t !== null, hasResume: storedResume !== null });
    setRoom(r);
    setToken(t);
    setResume(storedResume);
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

  const tryResume = (roomCode: string, mobileToken: string, resume: string) => {
    if (phaseRef.current === 'joining') return;
    console.info('[repko/MobilePage] resume attempt', { room: roomCode });
    setPhase('joining');
    const wsUrl = buildWsUrl(window.location.host, 'repko', roomCode, mobileToken);
    const conn = new Connection(
      wsUrl,
      {
        onWelcome: (msg) => {
          try { window.localStorage.setItem(resumeKey(roomCode), msg.resumeToken); } catch { /* ignore */ }
          clearJoinTimeout();
          setMyPlayerId(msg.playerId);
          setName(msg.you?.name ?? '');
          setPhase('playing');
          console.info('[repko/MobilePage] welcome (resume)', { playerId: msg.playerId });
        },
        onState: (msg) => {
          console.info('[repko/MobilePage] state', { phase: msg.phase, players: msg.players.length });
          setState(msg);
        },
        onError: (msg) => {
          if (msg.message.startsWith('resume failed:')) {
            console.warn('[repko/MobilePage] resume failed', msg.message);
            try { window.localStorage.removeItem(resumeKey(roomCode)); } catch { /* ignore */ }
            setResume(null);
            resetToName('Could not resume your previous session. Please rejoin.');
            return;
          }
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
    conn.sendHelloResume(roomCode, resume);
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

  useEffect(() => {
    if (room === null || token === null || resume === null) return;
    if (phaseRef.current !== 'name') return;
    tryResume(room, token, resume);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [room, token, resume]);

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
      const stored = room !== null ? (() => {
        try { return window.localStorage.getItem(resumeKey(room)); } catch { return null; }
      })() : null;
      if (room === null || token === null || stored === null) {
        resetToName('Disconnected from the room.');
        return;
      }
      setResume(stored);
      setPhase('disconnected');
    }
  };

  const handleRejoin = () => {
    if (room === null || token === null || resume === null) {
      resetToName('Missing session data. Please rejoin manually.');
      return;
    }
    tryResume(room, token, resume);
  };

  const handleLeaveFromDisconnect = async () => {
    if (room === null || resume === null) {
      resetToName('You left the room.');
      return;
    }
    try {
      await leaveRoom(room, resume);
    } catch (err) {
      console.warn('[repko/MobilePage] leaveRoom failed', err);
    }
    try { window.localStorage.removeItem(resumeKey(room)); } catch { /* ignore */ }
    setResume(null);
    resetToName('You left the room.');
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
          try { window.localStorage.setItem(resumeKey(room), msg.resumeToken); } catch { /* ignore */ }
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
          if (msg.message.startsWith('resume failed:')) {
            try { window.localStorage.removeItem(resumeKey(room)); } catch { /* ignore */ }
            setResume(null);
            resetToName('Could not resume your previous session. Please rejoin.');
            return;
          }
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
    if (room !== null) {
      try { window.localStorage.removeItem(resumeKey(room)); } catch { /* ignore */ }
    }
    connectionRef.current?.sendLeave();
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

  if (phase === 'disconnected') {
    return (
      <div className="fixed inset-0 flex flex-col items-center justify-center gap-6 bg-white px-6 text-slate-900">
        <div className="flex flex-col items-center gap-2">
          <span className="text-3xl">🔌</span>
          <h1 className="text-xl font-semibold">Connection lost</h1>
          <p className="text-sm text-center text-slate-600">
            Your game is still active. Rejoin to continue, or leave to release your tiles.
          </p>
        </div>
        <div className="flex w-full max-w-sm flex-col gap-3">
          <button
            type="button"
            onClick={handleRejoin}
            className="w-full rounded-md bg-slate-900 px-4 py-3 text-base font-medium text-white hover:bg-slate-700"
          >
            Rejoin
          </button>
          <button
            type="button"
            onClick={handleLeaveFromDisconnect}
            className="w-full rounded-md border border-slate-300 bg-white px-4 py-3 text-base font-medium text-slate-700 hover:bg-slate-100"
          >
            Leave
          </button>
        </div>
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
