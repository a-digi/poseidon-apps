import type { RoomStatus } from '../api';

interface RoomCardProps {
  room: RoomStatus;
  onShareQR: (code: string) => void;
  onDestroy: (code: string) => void;
  onKick: (code: string, playerId: string) => void;
}

export function RoomCard(props: RoomCardProps) {
  const { room, onShareQR, onDestroy, onKick } = props;

  const seconds = Math.max(0, Math.floor(Date.now() / 1000 - room.createdAt));
  const age =
    seconds < 60
      ? 'just now'
      : seconds < 3600
        ? `${Math.floor(seconds / 60)}m ago`
        : `${Math.floor(seconds / 3600)}h ago`;

  const playerCount = room.players.length;
  const playerLabel = `${playerCount} player${playerCount === 1 ? '' : 's'}`;

  return (
    <div className="bg-white border border-slate-200 rounded-lg p-4 mb-3">
      <h3 className="text-lg font-mono font-semibold text-slate-900">Room {room.code}</h3>
      <p className="text-sm text-slate-500 mt-1">
        {playerLabel} · {age}
      </p>
      <div className="mt-3">
        {room.players.length === 0 ? (
          <p className="text-sm text-slate-500 italic mb-3">(waiting for players)</p>
        ) : (
          <ul className="mb-3 space-y-1">
            {room.players.map((p) => (
              <li key={p.id} className="flex items-center justify-between text-sm">
                <span className="text-slate-700">{p.name}</span>
                <button
                  type="button"
                  onClick={() => onKick(room.code, p.id)}
                  aria-label={`Kick ${p.name}`}
                  className="rounded px-2 py-0.5 text-xs text-red-600 hover:bg-red-50"
                >
                  Kick
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>
      <div className="mt-4 flex justify-end gap-2">
        <button
          type="button"
          onClick={() => onShareQR(room.code)}
          aria-label={`Show QR code for room ${room.code}`}
          className="px-3 py-1.5 rounded-md border border-slate-300 bg-white text-sm text-slate-700 hover:bg-slate-50"
        >
          QR
        </button>
        <button
          type="button"
          onClick={() => onDestroy(room.code)}
          aria-label={`Destroy room ${room.code}`}
          className="px-3 py-1.5 rounded-md text-sm font-medium text-red-600 hover:bg-red-50"
        >
          Destroy
        </button>
      </div>
    </div>
  );
}
