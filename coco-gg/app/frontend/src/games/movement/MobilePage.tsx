import { useEffect, useRef, useState } from 'react';
import { buildWsUrl } from '../../shell/api';
import { Connection } from './game/net/Connection';
import { PhaserGame } from './game/PhaserGame';
import type { ErrorMsg } from './types';

const MAX_NAME_LENGTH = 32;

type Phase = 'name' | 'joining' | 'playing';

function MobilePage() {
  const [name, setName] = useState('');
  const [room, setRoom] = useState<string | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [phase, setPhase] = useState<Phase>('name');
  const [connection, setConnection] = useState<Connection | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const connectionRef = useRef<Connection | null>(null);
  const phaseRef = useRef<Phase>(phase);
  const joinTimeoutRef = useRef<number | null>(null);

  useEffect(() => {
    const sp = new URLSearchParams(window.location.search);
    setRoom(sp.get('room'));
    try {
      setToken(window.localStorage.getItem('coco_gg_mobile_token'));
    } catch {
      setToken(null);
    }
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

  const handleJoin = () => {
    if (!canJoin || room === null || token === null) return;
    setErrorMessage(null);
    setPhase('joining');
    const wsUrl = buildWsUrl(window.location.host, 'movement', room, token);
    const conn = new Connection(
      wsUrl,
      {
        onWelcome: () => {
          clearJoinTimeout();
          setPhase('playing');
        },
        onError: (msg) => {
          clearJoinTimeout();
          setErrorMessage(msg.message);
          conn.disconnect();
          setPhase('name');
        },
        onClose: () => {
          clearJoinTimeout();
          if (phaseRef.current === 'joining') {
            setErrorMessage('Connection lost before the server answered. Check that the plugin is running on the desktop.');
            setPhase('name');
          }
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
        setErrorMessage('Timed out — the server did not respond in 8 seconds.');
        setPhase('name');
      }
    }, 8000);
    connectionRef.current = conn;
    setConnection(conn);
  };

  const handleAbort = () => {
    clearJoinTimeout();
    connectionRef.current?.disconnect();
    setErrorMessage('Cancelled.');
    setPhase('name');
  };

  const handleLeave = () => {
    connectionRef.current?.disconnect();
    setConnection(null);
    setErrorMessage('You left the room.');
    setPhase('name');
  };

  const handleKickError = (msg: ErrorMsg) => {
    connectionRef.current?.disconnect();
    setConnection(null);
    setErrorMessage(msg.message || 'You were kicked from the room.');
    setPhase('name');
  };

  const handlePlayingClose = () => {
    // Guard so this doesn't fire on top of an already-handled kick: when
    // handleKickError calls disconnect, the socket onclose follows
    // asynchronously, by which time phaseRef is already 'name'.
    if (phaseRef.current !== 'playing') return;
    connectionRef.current?.disconnect();
    setConnection(null);
    setErrorMessage('Disconnected from the room.');
    setPhase('name');
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
      <div className="fixed inset-0 bg-slate-100">
        <PhaserGame
          connection={connection}
          mode="mobile"
          onError={handleKickError}
          onClose={handlePlayingClose}
        />
        <button
          type="button"
          onClick={handleLeave}
          aria-label="Leave room"
          className="absolute top-3 left-3 z-10 rounded-full bg-white/90 px-3 py-2 text-sm font-medium text-slate-700 shadow-md hover:bg-white"
        >
          ← Leave
        </button>
      </div>
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
        <h1 className="text-xl font-semibold">Join Coco GG</h1>
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
