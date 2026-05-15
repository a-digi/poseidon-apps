import { useMemo } from 'react';
import type { StateMsg } from '../types';
import { scoreBreakdown } from './scoring';

interface ScoreBreakdownModalProps {
  state: StateMsg;
  playerId: string;
  onClose: () => void;
}

export function ScoreBreakdownModal({ state, playerId, onClose }: ScoreBreakdownModalProps) {
  const player = state.players.find((p) => p.id === playerId);
  const breakdown = useMemo(() => scoreBreakdown(state, playerId), [state, playerId]);
  if (player === undefined) return null;
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      onClick={onClose}
    >
      <div
        className="max-h-[80vh] w-full max-w-md overflow-y-auto rounded-lg bg-white p-4 shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="mb-3 flex items-center justify-between border-b border-slate-200 pb-2">
          <div className="flex items-center gap-2">
            <span
              className="inline-block h-3 w-3 rounded-full"
              style={{ background: player.color }}
              aria-hidden
            />
            <h2 className="text-base font-bold text-slate-900">{player.name}</h2>
            <span className="text-sm text-slate-500">— score {breakdown.total}</span>
          </div>
          <button
            type="button"
            aria-label="Close"
            onClick={onClose}
            className="rounded px-2 py-0.5 text-slate-500 hover:bg-slate-100 hover:text-slate-900"
          >
            ✕
          </button>
        </header>

        <section className="mb-3">
          <h3 className="mb-1 text-sm font-semibold text-slate-700">
            Tiles ({breakdown.tileScore} pts)
          </h3>
          {breakdown.tiles.length === 0 ? (
            <p className="text-xs italic text-slate-500">No tiles owned.</p>
          ) : (
            <ul className="space-y-1 text-xs">
              {breakdown.tiles.map((t) => (
                <li key={`${t.q},${t.r}`} className="flex items-center justify-between">
                  <span className="truncate">
                    🌍 {t.name ?? `(${t.q},${t.r})`}
                    <span className="ml-1 text-slate-500">yields {t.yieldTotal}</span>
                  </span>
                  <span className="font-semibold text-slate-800">{t.score}</span>
                </li>
              ))}
            </ul>
          )}
        </section>

        <section>
          <h3 className="mb-1 text-sm font-semibold text-slate-700">
            Units ({breakdown.unitScore} pts)
          </h3>
          {breakdown.units.length === 0 ? (
            <p className="text-xs italic text-slate-500">No units deployed.</p>
          ) : (
            <ul className="space-y-1 text-xs">
              {breakdown.units.map((u) => (
                <li key={`${u.type}-${u.level}`} className="flex items-center justify-between">
                  <span>
                    {u.icon} {u.name} <span className="text-slate-500">L{u.level} ×{u.count} (power {u.power})</span>
                  </span>
                  <span className="font-semibold text-slate-800">{u.subtotal}</span>
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>
    </div>
  );
}
