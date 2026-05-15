import { useCallback, useMemo, useState } from 'react';
import type { ClientAction, StateMsg, Tile } from '../types';
import { Board } from './Board';
import { ActionPanel, type ActionMode } from './ActionPanel';
import { PlayerStatus } from './PlayerStatus';
import { ResourcePanel } from './ResourcePanel';
import { CivilizationPicker } from './CivilizationPicker';
import { DiplomacyBanner } from './DiplomacyBanner';
import { reachableForAction } from './coords';

interface GameViewProps {
  state: StateMsg | null;
  myPlayerId: string | null;
  onAction: (action: ClientAction) => void;
  onLeave: () => void;
}

function tileKey(q: number, r: number): string {
  return `${q},${r}`;
}

function findTile(state: StateMsg, q: number, r: number): Tile | undefined {
  return (state.board?.tiles ?? []).find((t) => t.q === q && t.r === r);
}

export function GameView({ state, myPlayerId, onAction, onLeave }: GameViewProps) {
  const [mode, setMode] = useState<ActionMode>('none');
  const [selectedTile, setSelectedTile] = useState<{ q: number; r: number } | null>(null);
  const [selectedSource, setSelectedSource] = useState<{ q: number; r: number } | null>(null);
  const [selectedUnits, setSelectedUnits] = useState<Set<number>>(new Set());

  const isMyTurn =
    state !== null &&
    state.currentTurn?.playerId === myPlayerId &&
    myPlayerId !== null;

  console.info('[repko/GameView] render', {
    phase: state?.phase ?? '(null)',
    hasBoard: state?.board !== undefined,
    hasYou: state?.you !== undefined,
    players: state?.players.length ?? 0,
    myPlayerId,
  });

  const resetActionState = useCallback(() => {
    setMode('none');
    setSelectedSource(null);
    setSelectedUnits(new Set());
  }, []);

  const handleModeChange = useCallback((next: ActionMode) => {
    setMode(next);
    setSelectedSource(null);
    setSelectedUnits(new Set());
  }, []);

  const toggleUnit = useCallback((idx: number) => {
    setSelectedUnits((prev) => {
      const next = new Set(prev);
      if (next.has(idx)) next.delete(idx);
      else next.add(idx);
      return next;
    });
  }, []);

  const neutralTileKeys = useMemo(() => {
    if (state === null) return new Set<string>();
    const s = new Set<string>();
    for (const t of state.board?.tiles ?? []) {
      if (t.ownerId === '') s.add(tileKey(t.q, t.r));
    }
    return s;
  }, [state]);

  const reachableKeys = useMemo(() => {
    if (state === null || myPlayerId === null) return new Set<string>();
    if (mode !== 'move' && mode !== 'attack' && mode !== 'offer_diplomacy' && mode !== 'buy') {
      return new Set<string>();
    }
    const tiles = state.board?.tiles ?? [];
    const tileByKey = new Map<string, Tile>();
    for (const t of tiles) tileByKey.set(tileKey(t.q, t.r), t);

    if (mode === 'buy') {
      const owned = tiles.filter((t) => t.ownerId === myPlayerId);
      const out = new Set<string>();
      for (const src of owned) {
        const reach = reachableForAction(state, myPlayerId, { q: src.q, r: src.r });
        for (const r of reach) {
          const target = tileByKey.get(tileKey(r.q, r.r));
          if (target !== undefined && target.ownerId === '') out.add(tileKey(r.q, r.r));
        }
      }
      return out;
    }

    if (selectedSource === null) return new Set<string>();
    const reach = reachableForAction(state, myPlayerId, selectedSource);
    const out = new Set<string>();
    for (const r of reach) {
      const target = tileByKey.get(tileKey(r.q, r.r));
      if (target === undefined) continue;
      if (mode === 'move') {
        if (target.ownerId === myPlayerId) out.add(tileKey(r.q, r.r));
      } else if (mode === 'attack') {
        if (target.ownerId !== myPlayerId) out.add(tileKey(r.q, r.r));
      } else if (mode === 'offer_diplomacy') {
        if (target.ownerId !== myPlayerId && target.ownerId !== '') out.add(tileKey(r.q, r.r));
      }
    }
    return out;
  }, [state, myPlayerId, mode, selectedSource]);

  const handleStartingTilePick = useCallback(
    (q: number, r: number) => {
      if (state === null || !isMyTurn) return;
      const tile = findTile(state, q, r);
      if (tile === undefined || tile.ownerId !== '') return;
      onAction({ type: 'pick_starting_tile', q, r });
    },
    [state, isMyTurn, onAction],
  );

  const handlePlayingTileClick = useCallback(
    (q: number, r: number) => {
      if (state === null || !isMyTurn || myPlayerId === null) return;
      const tile = findTile(state, q, r);
      if (tile === undefined) return;

      if (mode === 'none' || mode === 'recruit' || mode === 'upgrade') {
        setSelectedTile({ q, r });
        return;
      }

      if (mode === 'buy') {
        if (reachableKeys.has(tileKey(q, r)) && tile.ownerId === '') {
          onAction({ type: 'buy_tile', q, r });
          resetActionState();
        }
        return;
      }

      if (mode === 'move' || mode === 'attack' || mode === 'offer_diplomacy') {
        if (selectedSource === null) {
          if (tile.ownerId !== myPlayerId) return;
          if ((mode === 'move' || mode === 'attack') && tile.garrison.length === 0) return;
          setSelectedSource({ q, r });
          setSelectedUnits(new Set());
          return;
        }

        if (tile.ownerId === myPlayerId && (q !== selectedSource.q || r !== selectedSource.r)) {
          if (mode === 'move') {
            if (reachableKeys.has(tileKey(q, r))) {
              onAction({
                type: 'move',
                fromQ: selectedSource.q,
                fromR: selectedSource.r,
                toQ: q,
                toR: r,
                unitIndices: [...selectedUnits],
              });
              resetActionState();
              return;
            }
          }
          if (mode === 'attack' || mode === 'offer_diplomacy') {
            if (tile.garrison.length === 0 && mode === 'attack') return;
            setSelectedSource({ q, r });
            setSelectedUnits(new Set());
            return;
          }
          setSelectedSource({ q, r });
          setSelectedUnits(new Set());
          return;
        }

        if (!reachableKeys.has(tileKey(q, r))) return;

        if (mode === 'move') {
          if (tile.ownerId !== myPlayerId) return;
          onAction({
            type: 'move',
            fromQ: selectedSource.q,
            fromR: selectedSource.r,
            toQ: q,
            toR: r,
            unitIndices: [...selectedUnits],
          });
          resetActionState();
          return;
        }
        if (mode === 'attack') {
          if (tile.ownerId === myPlayerId) return;
          onAction({
            type: 'attack',
            fromQ: selectedSource.q,
            fromR: selectedSource.r,
            toQ: q,
            toR: r,
            unitIndices: [...selectedUnits],
          });
          resetActionState();
          return;
        }
        if (mode === 'offer_diplomacy') {
          if (tile.ownerId === myPlayerId || tile.ownerId === '') return;
          onAction({ type: 'offer_diplomacy', q, r });
          resetActionState();
          return;
        }
      }
    },
    [state, isMyTurn, myPlayerId, mode, selectedSource, selectedUnits, reachableKeys, onAction, resetActionState],
  );

  if (state === null) {
    return (
      <div className="fixed inset-0 flex items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-slate-200 border-t-slate-700" />
      </div>
    );
  }

  if (state.phase === 'lobby') {
    return <Lobby state={state} onLeave={onLeave} />;
  }

  if (state.phase === 'civ_pick') {
    const myCommittedCivId =
      myPlayerId !== null
        ? state.players.find((p) => p.id === myPlayerId)?.civilizationId ?? ''
        : '';

    if (myCommittedCivId === '') {
      return <CivilizationPicker state={state} myPlayerId={myPlayerId} onAction={onAction} />;
    }

    const totalPlayers = state.players.length;
    const committedCount = state.players.filter((p) => p.civilizationId !== '').length;
    const remaining = totalPlayers - committedCount;
    const playerWord = remaining === 1 ? 'player' : 'players';

    return (
      <div className="fixed inset-0 flex flex-col bg-slate-50">
        <header className="border-b border-slate-200 bg-white px-4 py-2 text-xs text-slate-600">
          Waiting for {remaining} of {totalPlayers} {playerWord} to pick their civilization…
        </header>
        <main className="relative flex-1 overflow-hidden">
          <Board state={state} myPlayerId={myPlayerId} />
          <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
            <div className="rounded-md bg-white/90 px-3 py-2 text-xs text-slate-700 shadow">
              Preparing the field…
            </div>
          </div>
        </main>
      </div>
    );
  }

  if (state.phase === 'tile_pick') {
    const current = state.players.find((p) => p.id === state.currentTurn?.playerId) ?? null;
    return (
      <div className="fixed inset-0 flex flex-col bg-slate-100">
        <header className="flex items-center gap-2 border-b border-slate-200 bg-white px-3 py-2">
          <button
            type="button"
            onClick={onLeave}
            className="text-xs text-slate-500 hover:text-slate-900"
          >
            ← Leave
          </button>
          <span className="text-sm font-medium text-slate-800">
            {isMyTurn ? 'Pick your starting tile' : `Waiting for ${current?.name ?? '…'} to pick…`}
          </span>
        </header>
        <main className="relative flex-1 overflow-hidden">
          <Board
            state={state}
            myPlayerId={myPlayerId}
            startingTileHighlights={isMyTurn ? neutralTileKeys : new Set()}
            onTileClick={handleStartingTilePick}
          />
        </main>
      </div>
    );
  }

  if (state.phase === 'game_over') {
    const winner = state.players.find((p) => p.id === state.winnerId) ?? null;
    return (
      <div className="fixed inset-0 flex flex-col bg-slate-100">
        <header className="flex items-center gap-2 border-b border-slate-200 bg-white px-3 py-2">
          <button
            type="button"
            onClick={onLeave}
            className="text-xs text-slate-500 hover:text-slate-900"
          >
            ← Leave
          </button>
          <span className="text-sm font-medium text-slate-800">Game over</span>
        </header>
        <main className="relative flex-1 overflow-hidden">
          <Board state={state} myPlayerId={myPlayerId} />
          <div className="absolute inset-0 flex items-center justify-center bg-slate-900/40">
            <div className="flex flex-col items-center gap-3 rounded-lg bg-white px-6 py-4 shadow-lg">
              <p className="text-lg font-semibold text-slate-900">
                {winner !== null ? `${winner.name} wins!` : 'No winner'}
              </p>
              <button
                type="button"
                onClick={onLeave}
                className="rounded bg-slate-900 px-3 py-1.5 text-xs font-medium text-white hover:bg-slate-700"
              >
                Leave
              </button>
            </div>
          </div>
        </main>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 flex flex-col bg-slate-100">
      <header className="flex items-center gap-2 border-b border-slate-200 bg-white px-3 py-2">
        <button
          type="button"
          onClick={onLeave}
          className="text-xs text-slate-500 hover:text-slate-900"
        >
          ← Leave
        </button>
        <PlayerStatus state={state} myPlayerId={myPlayerId} />
      </header>

      <DiplomacyBanner state={state} myPlayerId={myPlayerId} onAction={onAction} />

      <main className="relative flex-1 overflow-hidden">
        <Board
          state={state}
          myPlayerId={myPlayerId}
          reachableHighlights={reachableKeys}
          selectedSource={selectedSource}
          onTileClick={handlePlayingTileClick}
        />
      </main>

      <div className="border-t border-slate-200 bg-white px-3 py-2">
        <ResourcePanel resources={state.you?.resources ?? {}} />
      </div>

      <ActionPanel
        state={state}
        myPlayerId={myPlayerId}
        isMyTurn={isMyTurn}
        mode={mode}
        onModeChange={handleModeChange}
        selectedTile={selectedTile}
        selectedSourceForMove={selectedSource}
        selectedUnitIndices={selectedUnits}
        onUnitIndexToggle={toggleUnit}
        onAction={onAction}
      />
    </div>
  );
}

interface LobbyProps {
  state: StateMsg;
  onLeave: () => void;
}

function Lobby({ state, onLeave }: LobbyProps) {
  return (
    <div className="fixed inset-0 flex flex-col bg-white">
      <header className="flex items-center gap-2 border-b border-slate-200 px-3 py-2">
        <button
          type="button"
          onClick={onLeave}
          className="text-xs text-slate-500 hover:text-slate-900"
        >
          ← Leave
        </button>
        <span className="text-sm font-medium">Waiting for game to start…</span>
      </header>
      <main className="flex flex-1 flex-col items-center justify-center gap-4 p-4">
        <p className="text-xs text-slate-500">{state.players.length} player(s) joined.</p>
        <ul className="space-y-1">
          {state.players.map((p) => (
            <li key={p.id} className="flex items-center gap-2 text-sm">
              <span className="h-3 w-3 rounded-full" style={{ background: p.color }} />
              {p.name}
            </li>
          ))}
        </ul>
        <p className="mt-4 text-xs italic text-slate-400">
          The host will start the game from the dashboard.
        </p>
      </main>
    </div>
  );
}

export default GameView;
