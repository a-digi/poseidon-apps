import { useMemo, useState } from 'react';
import type { StateMsg } from '../types';
import { playerScore } from './scoring';
import { ScoreBreakdownModal } from './ScoreBreakdownModal';

interface RankingPanelProps {
  state: StateMsg;
  myPlayerId: string | null;
}

export function RankingPanel({ state, myPlayerId }: RankingPanelProps) {
  const [open, setOpen] = useState(false);
  const [modalPlayerId, setModalPlayerId] = useState<string | null>(null);

  const rows = useMemo(() => {
    return state.players
      .filter((p) => p.leftGame !== true)
      .map((p) => ({ player: p, score: playerScore(state, p.id) }))
      .sort((a, b) => b.score - a.score);
  }, [state]);

  if (open === false) {
    return (
      <button
        type="button"
        aria-label="Show ranking"
        onClick={() => setOpen(true)}
        className="absolute right-0 top-1/2 z-30 flex h-20 w-8 -translate-y-1/2 items-center justify-center rounded-l-md bg-slate-800 text-xs font-bold text-white shadow-md hover:bg-slate-700"
      >
        ‹
      </button>
    );
  }

  return (
    <>
      <aside className="absolute inset-y-0 right-0 z-30 flex w-64 flex-col bg-slate-900 text-white shadow-2xl">
        <header className="flex items-center justify-between border-b border-slate-700 px-3 py-2">
          <h2 className="text-sm font-bold">Ranking</h2>
          <button
            type="button"
            aria-label="Hide ranking"
            onClick={() => setOpen(false)}
            className="rounded px-2 py-0.5 text-slate-300 hover:bg-slate-800 hover:text-white"
          >
            ›
          </button>
        </header>
        <ol className="flex-1 overflow-y-auto">
          {rows.map((row, idx) => {
            const isMe = row.player.id === myPlayerId;
            return (
              <li
                key={row.player.id}
                className={`flex items-center gap-2 border-b border-slate-800 px-3 py-2 text-sm ${isMe ? 'bg-slate-800' : ''} ${row.player.eliminated ? 'opacity-60' : ''}`}
              >
                <span className="w-6 text-xs text-slate-400">#{idx + 1}</span>
                <span
                  className="inline-block h-3 w-3 shrink-0 rounded-full"
                  style={{ background: row.player.color }}
                  aria-hidden
                />
                <span className="min-w-0 flex-1 truncate">{row.player.name}</span>
                <button
                  type="button"
                  onClick={() => setModalPlayerId(row.player.id)}
                  className="rounded px-1.5 py-0.5 font-bold text-amber-300 hover:bg-slate-800 hover:underline"
                  aria-label={`Show score breakdown for ${row.player.name}`}
                >
                  {row.score}
                </button>
              </li>
            );
          })}
        </ol>
      </aside>
      {modalPlayerId !== null && (
        <ScoreBreakdownModal
          state={state}
          playerId={modalPlayerId}
          onClose={() => setModalPlayerId(null)}
        />
      )}
    </>
  );
}
