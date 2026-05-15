import type { RoomStatus } from '../api';

interface RoomCardProps {
  room: RoomStatus;
  onShareQR: (code: string) => void;
  onDestroy: (code: string) => void;
  onKick: (code: string, playerId: string) => void;
  onStartGame: (code: string) => void;
}

export function RoomCard(props: RoomCardProps) {
  const { room, onShareQR, onDestroy, onKick, onStartGame } = props;

  const seconds = Math.max(0, Math.floor(Date.now() / 1000 - room.createdAt));
  const age =
    seconds < 60
      ? 'just now'
      : seconds < 3600
        ? `${Math.floor(seconds / 60)}m ago`
        : `${Math.floor(seconds / 3600)}h ago`;

  const playerCount = room.players.length;
  const playerLabel = `${playerCount} player${playerCount === 1 ? '' : 's'}`;

  const canStart = room.phase === 'lobby' && playerCount >= 2;
  const startTitle =
    room.phase !== 'lobby'
      ? `Game already ${room.phase}`
      : playerCount < 2
        ? 'Need at least 2 players'
        : 'Start the game';

  return (
    <div className="bg-white border border-slate-200 rounded p-2 mb-1">
      <div className="flex items-center gap-2">
        <h3 className="text-sm font-mono font-semibold text-slate-900">Room {room.code}</h3>
        <span className="text-[10px] uppercase tracking-wide px-1.5 py-0.5 rounded bg-slate-100 text-slate-600">
          {room.phase}
        </span>
      </div>
      <p className="text-xs text-slate-500">
        {playerLabel} · {age}
      </p>
      <div className="mt-3">
        {room.players.length === 0 ? (
          <p className="text-xs text-slate-500 italic mb-1">(waiting for players)</p>
        ) : (
          <ul className="mb-1 space-y-0.5">
            {room.players.map((p) => (
              <li key={p.id} className="flex items-center justify-between text-sm">
                <span className="text-slate-700">{p.name}</span>
                <button
                  type="button"
                  onClick={() => onKick(room.code, p.id)}
                  aria-label={`Kick ${p.name}`}
                  className="rounded px-1.5 py-0.5 text-xs text-red-600 hover:bg-red-50"
                >
                  Kick
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>
      <div className="mt-1.5 flex justify-end gap-1.5">
        <button
          type="button"
          onClick={() => onStartGame(room.code)}
          disabled={!canStart}
          title={startTitle}
          className="px-2 py-1 rounded bg-slate-900 text-xs font-medium text-white transition-colors hover:bg-slate-700 disabled:bg-slate-400 disabled:cursor-not-allowed"
        >
          Start game
        </button>
        <button
          type="button"
          onClick={() => onShareQR(room.code)}
          aria-label={`Show QR code for room ${room.code}`}
          className="px-2 py-1 rounded bg-slate-900 text-xs font-medium text-white transition-colors hover:bg-slate-700"
        >
          QR (Invite Players)
        </button>
        <button
          type="button"
          onClick={() => onDestroy(room.code)}
          aria-label={`Destroy room ${room.code}`}
          className="px-2 py-1 rounded text-xs font-medium text-red-600 hover:bg-red-50"
        >
          Destroy
        </button>
      </div>
    </div>
  );
}
