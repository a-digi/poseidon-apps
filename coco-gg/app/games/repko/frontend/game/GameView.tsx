import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { ClientAction, GameEvent, StackPick, StateMsg, Tile, UnitClass } from '../types';
import { Board } from './Board';
import {
  ActionPanel,
  reachableSources,
  sortSourcesByPreference,
  type SubAction,
} from './ActionPanel';
import { PlayerStatus } from './PlayerStatus';
import { ResourcePanel } from './ResourcePanel';
import { CivilizationPicker } from './CivilizationPicker';
import { DiplomacyBanner } from './DiplomacyBanner';
import { EventBanner } from './EventBanner';
import { RankingPanel } from './RankingPanel';
import { pathThroughOwnedTiles, reachableForAction } from './coords';
import { UNIT_CLASS } from './units';
import { ArmyInspectorPopover } from './ArmyInspectorPopover';

export interface Animation {
  id: string;
  kind: 'claim' | 'conquest' | 'repulsed' | 'tie';
  q: number;
  r: number;
  expiresAt: number;
}

const ANIMATION_DURATION_MS = 1500;

interface GameViewProps {
  state: StateMsg | null;
  myPlayerId: string | null;
  events: GameEvent[];
  onAction: (action: ClientAction) => void;
  onLeave: () => void;
}

function tileKey(q: number, r: number): string {
  return `${q},${r}`;
}

function findTile(state: StateMsg, q: number, r: number): Tile | undefined {
  return (state.board?.tiles ?? []).find((t) => t.q === q && t.r === r);
}

export function GameView({ state, myPlayerId, events, onAction, onLeave }: GameViewProps) {
  const [panelHidden, setPanelHidden] = useState(false);
  const [selectedTile, setSelectedTile] = useState<{ q: number; r: number } | null>(null);
  const [subAction, setSubAction] = useState<SubAction>('inspect');
  const [selectedStackCounts, setSelectedStackCounts] = useState<Map<number, number>>(new Map());
  const [attackSourceOverride, setAttackSourceOverride] = useState<
    { q: number; r: number } | null
  >(null);
  const [marchDestination, setMarchDestination] = useState<
    { q: number; r: number } | null
  >(null);
  const [inspectedArmyId, setInspectedArmyId] = useState<string | null>(null);
  const [animations, setAnimations] = useState<Animation[]>([]);
  const eventCursorRef = useRef(0);

  useEffect(() => {
    if (events.length <= eventCursorRef.current) {
      eventCursorRef.current = events.length;
      return;
    }
    const newAnims: Animation[] = [];
    for (let i = eventCursorRef.current; i < events.length; i++) {
      const ev = events[i];
      if (ev.targetQ === undefined || ev.targetR === undefined) continue;
      let kind: Animation['kind'] | null = null;
      if (ev.kind === 'buy_tile') kind = 'claim';
      else if (ev.kind === 'attack_won') kind = 'conquest';
      else if (ev.kind === 'attack_lost') kind = 'repulsed';
      else if (ev.kind === 'attack_tie') kind = 'tie';
      if (kind === null) continue;
      newAnims.push({
        id: `${i}-${kind}-${Math.random().toString(36).slice(2, 7)}`,
        kind,
        q: ev.targetQ,
        r: ev.targetR,
        expiresAt: Date.now() + ANIMATION_DURATION_MS,
      });
    }
    eventCursorRef.current = events.length;
    if (newAnims.length === 0) return;
    setAnimations((prev) => [...prev, ...newAnims]);
  }, [events]);

  useEffect(() => {
    if (animations.length === 0) return;
    const id = window.setInterval(() => {
      const now = Date.now();
      setAnimations((prev) => prev.filter((a) => a.expiresAt > now));
    }, 250);
    return () => window.clearInterval(id);
  }, [animations.length]);

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

  const handleSubActionChange = useCallback((next: SubAction) => {
    setSubAction(next);
    setSelectedStackCounts(new Map());
    if (next !== 'attack') setAttackSourceOverride(null);
    if (next !== 'march') setMarchDestination(null);
  }, []);

  useEffect(() => {
    setMarchDestination(null);
  }, [selectedTile]);

  const handleStackCountChange = useCallback((stackIndex: number, count: number) => {
    setSelectedStackCounts((prev) => {
      const next = new Map(prev);
      if (count <= 0) next.delete(stackIndex);
      else next.set(stackIndex, count);
      return next;
    });
  }, []);

  const buildStackPicks = useCallback((counts: Map<number, number>): StackPick[] => {
    const picks: StackPick[] = [];
    for (const [stackIndex, count] of counts) {
      if (count > 0) picks.push({ stackIndex, count });
    }
    return picks;
  }, []);

  const neutralTileKeys = useMemo(() => {
    if (state === null) return new Set<string>();
    const s = new Set<string>();
    for (const t of state.board?.tiles ?? []) {
      if (t.ownerId === '') s.add(tileKey(t.q, t.r));
    }
    return s;
  }, [state]);

  const selectedTileObj = useMemo<Tile | null>(() => {
    if (state === null || selectedTile === null) return null;
    return findTile(state, selectedTile.q, selectedTile.r) ?? null;
  }, [state, selectedTile]);

  const moveDestinationKeys = useMemo(() => {
    if (
      state === null ||
      myPlayerId === null ||
      selectedTile === null ||
      subAction !== 'move'
    ) {
      return new Set<string>();
    }
    const reach = reachableForAction(state, myPlayerId, selectedTile);
    const out = new Set<string>();
    for (const r of reach) {
      const target = findTile(state, r.q, r.r);
      if (target !== undefined && target.ownerId === myPlayerId) {
        out.add(tileKey(r.q, r.r));
      }
    }
    return out;
  }, [state, myPlayerId, selectedTile, subAction]);

  const reachableHighlights = subAction === 'move' ? moveDestinationKeys : new Set<string>();

  const marchPathHighlights = useMemo<Set<string>>(() => {
    if (
      state === null ||
      myPlayerId === null ||
      selectedTile === null ||
      subAction !== 'march' ||
      marchDestination === null
    ) {
      return new Set();
    }
    const path = pathThroughOwnedTiles(state, myPlayerId, selectedTile, marchDestination);
    if (path === null) return new Set();
    return new Set(path.map((h) => tileKey(h.q, h.r)));
  }, [state, subAction, marchDestination, selectedTile, myPlayerId]);

  const attackSource = useMemo<{ q: number; r: number } | null>(() => {
    if (state === null || myPlayerId === null || selectedTileObj === null) return null;
    if (subAction !== 'attack') return null;
    if (attackSourceOverride !== null) return attackSourceOverride;
    const sources = sortSourcesByPreference(
      reachableSources(state, myPlayerId, selectedTileObj),
    );
    if (sources.length === 0) return null;
    return { q: sources[0].q, r: sources[0].r };
  }, [state, myPlayerId, selectedTileObj, subAction, attackSourceOverride]);

  const boardSelectedSource = useMemo<{ q: number; r: number } | null>(() => {
    if (subAction === 'move') return selectedTile;
    if (subAction === 'attack') return attackSource;
    return null;
  }, [subAction, selectedTile, attackSource]);

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
      console.info('[repko/GameView] tile click', { q, r, subAction });
      if (state === null || !isMyTurn || myPlayerId === null) return;
      const tile = findTile(state, q, r);
      if (tile === undefined) return;

      if (subAction === 'move' && selectedTile !== null) {
        if (moveDestinationKeys.has(tileKey(q, r))) {
          onAction({
            type: 'move',
            fromQ: selectedTile.q,
            fromR: selectedTile.r,
            toQ: q,
            toR: r,
            units: buildStackPicks(selectedStackCounts),
          });
          setSubAction('inspect');
          setSelectedStackCounts(new Map());
          return;
        }
        setSelectedTile({ q, r });
        setSubAction('inspect');
        setSelectedStackCounts(new Map());
        setAttackSourceOverride(null);
        return;
      }

      if (subAction === 'march' && selectedTile !== null) {
        if (tile.ownerId === myPlayerId && (tile.q !== selectedTile.q || tile.r !== selectedTile.r)) {
          setMarchDestination({ q, r });
          return;
        }
        return;
      }

      setSelectedTile({ q, r });
      setSubAction('inspect');
      setSelectedStackCounts(new Map());
      setAttackSourceOverride(null);
    },
    [
      state,
      isMyTurn,
      myPlayerId,
      subAction,
      selectedTile,
      moveDestinationKeys,
      selectedStackCounts,
      onAction,
      buildStackPicks,
    ],
  );

  const handleEndTurn = useCallback(() => {
    onAction({ type: 'end_turn' });
    setSelectedTile(null);
    setSubAction('inspect');
    setSelectedStackCounts(new Map());
    setAttackSourceOverride(null);
    setMarchDestination(null);
  }, [onAction]);

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
      return (
        <>
          <EventBanner events={events} myPlayerId={myPlayerId} />
          <CivilizationPicker state={state} myPlayerId={myPlayerId} onAction={onAction} />
        </>
      );
    }

    const totalPlayers = state.players.length;
    const committedCount = state.players.filter((p) => p.civilizationId !== '').length;
    const remaining = totalPlayers - committedCount;
    const playerWord = remaining === 1 ? 'player' : 'players';

    return (
      <>
        <EventBanner events={events} myPlayerId={myPlayerId} />
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
      </>
    );
  }

  if (state.phase === 'tile_pick') {
    const current = state.players.find((p) => p.id === state.currentTurn?.playerId) ?? null;
    return (
      <>
        <EventBanner events={events} myPlayerId={myPlayerId} />
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
              animations={animations}
              startingTileHighlights={isMyTurn ? neutralTileKeys : new Set()}
              onTileClick={handleStartingTilePick}
            />
            <RankingPanel state={state} myPlayerId={myPlayerId} />
          </main>
        </div>
      </>
    );
  }

  if (state.phase === 'game_over') {
    const winner = state.players.find((p) => p.id === state.winnerId) ?? null;
    return (
      <>
        <EventBanner events={events} myPlayerId={myPlayerId} />
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
            <Board state={state} myPlayerId={myPlayerId} animations={animations} />
            <RankingPanel state={state} myPlayerId={myPlayerId} />
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
      </>
    );
  }

  return (
    <>
      <EventBanner events={events} myPlayerId={myPlayerId} />
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
        <button
          type="button"
          onClick={handleEndTurn}
          disabled={!isMyTurn}
          className="rounded-md bg-slate-900 px-3 py-1.5 text-xs font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          Skip Turn
        </button>
      </header>

      <DiplomacyBanner state={state} myPlayerId={myPlayerId} onAction={onAction} />

      <main className="relative flex-1 overflow-hidden">
        <Board
          state={state}
          myPlayerId={myPlayerId}
          animations={animations}
          reachableHighlights={reachableHighlights}
          marchPathHighlights={marchPathHighlights}
          selectedSource={boardSelectedSource}
          onTileClick={handlePlayingTileClick}
          attackMode={subAction === 'attack'}
          inspectedTile={selectedTile}
          armies={state.armies}
          onArmyClick={(id) => setInspectedArmyId(id)}
        />
        <RankingPanel state={state} myPlayerId={myPlayerId} />
      </main>

      <div className="flex items-center gap-2 border-t border-slate-200 bg-white px-2 py-2">
        <div className="min-w-0 flex-1 overflow-x-auto">
          <ResourcePanel
            resources={state.you?.resources ?? {}}
            armyBreakdown={(() => {
              const breakdown: Record<UnitClass, number> = {
                infantry: 0, armor: 0, artillery: 0, air: 0, special: 0,
              };
              if (myPlayerId !== null) {
                for (const tile of state.board?.tiles ?? []) {
                  if (tile.ownerId !== myPlayerId) continue;
                  for (const stack of tile.garrison) {
                    const cls = UNIT_CLASS[stack.type];
                    if (cls === undefined) continue;
                    breakdown[cls] += stack.count;
                  }
                }
              }
              return breakdown;
            })()}
          />
        </div>
        <button
          type="button"
          aria-label={panelHidden ? 'Show controls' : 'Hide controls'}
          onClick={() => setPanelHidden((h) => !h)}
          className="shrink-0 rounded-md bg-slate-800 px-3 py-1.5 text-sm font-bold text-white hover:bg-slate-700 active:bg-slate-900"
        >
          {panelHidden ? '↑' : '↓'}
        </button>
      </div>

      {!panelHidden && (
        <ActionPanel
          state={state}
          myPlayerId={myPlayerId}
          isMyTurn={isMyTurn}
          selectedTile={selectedTileObj}
          subAction={subAction}
          onSubActionChange={handleSubActionChange}
          selectedStackCounts={selectedStackCounts}
          onStackCountChange={handleStackCountChange}
          attackSourceOverride={attackSourceOverride}
          onAttackSourceChange={setAttackSourceOverride}
          marchDestination={marchDestination}
          onMarchDestinationChange={setMarchDestination}
          onAction={onAction}
        />
      )}
      <ArmyInspectorPopover
        army={state.armies?.find((a) => a.id === inspectedArmyId) ?? null}
        state={state}
        onClose={() => setInspectedArmyId(null)}
      />
      </div>
    </>
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
