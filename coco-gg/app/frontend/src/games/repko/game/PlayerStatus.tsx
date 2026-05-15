import type { StateMsg } from '../types';

interface PlayerStatusProps {
  state: StateMsg;
  myPlayerId: string | null;
}

export function PlayerStatus({ state, myPlayerId }: PlayerStatusProps) {
  const civs = state.civilizations ?? [];
  const currentPlayerId = state.currentTurn?.playerId ?? null;
  return (
    <div className="flex flex-1 items-center gap-1.5 overflow-x-auto">
      {state.players.map((p) => {
        const isCurrent = p.id === currentPlayerId;
        const isMe = p.id === myPlayerId;
        const ringCls = isCurrent ? 'ring-2 ring-amber-400' : 'ring-1 ring-slate-200';
        const civ = p.civilizationId !== '' ? civs.find((c) => c.id === p.civilizationId) : undefined;
        const flag = civ?.flag ?? '❔';
        return (
          <div
            key={p.id}
            className={`flex items-center gap-1.5 rounded-md bg-white px-2 py-1 ${ringCls} ${p.eliminated ? 'opacity-60' : ''}`}
          >
            <span className="text-sm leading-none">{flag}</span>
            <span className="inline-block h-3 w-3 rounded-full" style={{ background: p.color }} />
            <span className="max-w-[5rem] truncate text-xs font-medium text-slate-900">{p.name}</span>
            {isMe && (
              <span className="rounded bg-slate-900 px-1 text-[9px] font-semibold uppercase text-white">
                You
              </span>
            )}
            <span className="text-xs font-bold text-slate-700">{p.tileCount}🏳</span>
            {p.eliminated && (
              <span className="rounded bg-red-600 px-1 text-[9px] font-semibold uppercase text-white">
                ELIM
              </span>
            )}
            {isCurrent && !p.eliminated && (
              <span className="inline-block h-2 w-2 animate-pulse rounded-full bg-amber-500" title="Current turn" />
            )}
          </div>
        );
      })}
    </div>
  );
}

export default PlayerStatus;
