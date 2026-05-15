import { useMemo, useState } from 'react';
import type { ClientAction, StateMsg, Tile, Unit, UnitType } from '../types';

export type ActionMode =
  | 'none'
  | 'recruit'
  | 'upgrade'
  | 'move'
  | 'attack'
  | 'offer_diplomacy'
  | 'buy';

interface ActionPanelProps {
  state: StateMsg;
  myPlayerId: string | null;
  isMyTurn: boolean;
  mode: ActionMode;
  onModeChange: (mode: ActionMode) => void;
  selectedTile: { q: number; r: number } | null;
  selectedSourceForMove: { q: number; r: number } | null;
  selectedUnitIndices: Set<number>;
  onUnitIndexToggle: (index: number) => void;
  onAction: (action: ClientAction) => void;
}

export const UNIT_COSTS: Record<UnitType, { gold: number; iron: number; food: number }> = {
  infantry: { gold: 5, iron: 1, food: 0 },
  cavalry: { gold: 10, iron: 2, food: 0 },
  artillery: { gold: 20, iron: 3, food: 0 },
};

const UNIT_BASE_POWER: Record<UnitType, number> = {
  infantry: 1,
  cavalry: 2,
  artillery: 3,
};

const UNIT_ICON: Record<UnitType, string> = {
  infantry: '⚔️',
  cavalry: '🐎',
  artillery: '💥',
};

const MODES: ReadonlyArray<{ id: ActionMode; label: string }> = [
  { id: 'recruit', label: 'Recruit' },
  { id: 'upgrade', label: 'Upgrade' },
  { id: 'move', label: 'Move' },
  { id: 'attack', label: 'Attack' },
  { id: 'offer_diplomacy', label: 'Diplomacy' },
  { id: 'buy', label: 'Buy' },
];

function findTile(state: StateMsg, q: number, r: number): Tile | undefined {
  return (state.board?.tiles ?? []).find((t) => t.q === q && t.r === r);
}

function upgradeCostFor(unit: Unit): { gold: number; iron: number; food: number } | null {
  if (unit.level >= 3) return null;
  const base = UNIT_COSTS[unit.type];
  const mult = unit.level === 2 ? 2 : 1;
  return { gold: base.gold * mult, iron: base.iron * mult, food: base.food * mult };
}

interface ModeButtonProps {
  active: boolean;
  disabled: boolean;
  label: string;
  onClick: () => void;
}

function ModeButton({ active, disabled, label, onClick }: ModeButtonProps) {
  const activeCls = active
    ? 'bg-slate-900 text-white'
    : 'bg-white text-slate-700 hover:bg-slate-100';
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className={`rounded border border-slate-200 px-2 py-1 text-xs font-medium disabled:cursor-not-allowed disabled:bg-slate-300 disabled:text-slate-500 ${activeCls}`}
    >
      {label}
    </button>
  );
}

interface RecruitRowProps {
  unit: UnitType;
  count: number;
  onIncrement: () => void;
  onDecrement: () => void;
}

function RecruitRow({ unit, count, onIncrement, onDecrement }: RecruitRowProps) {
  const cost = UNIT_COSTS[unit];
  return (
    <div className="flex items-center justify-between gap-2 rounded border border-slate-200 bg-white px-2 py-1">
      <div className="flex items-center gap-2">
        <span className="text-base">{UNIT_ICON[unit]}</span>
        <span className="text-xs font-medium capitalize text-slate-800">{unit}</span>
        <span className="text-[10px] text-slate-500">
          {cost.gold}g · {cost.iron}i
        </span>
      </div>
      <div className="flex items-center gap-1">
        <button
          type="button"
          onClick={onDecrement}
          disabled={count === 0}
          className="rounded border border-slate-300 bg-white px-1.5 py-0.5 text-xs disabled:cursor-not-allowed disabled:bg-slate-200 disabled:text-slate-400"
        >
          −
        </button>
        <span className="w-5 text-center text-xs font-bold text-slate-900">{count}</span>
        <button
          type="button"
          onClick={onIncrement}
          className="rounded border border-slate-300 bg-white px-1.5 py-0.5 text-xs hover:bg-slate-100"
        >
          +
        </button>
      </div>
    </div>
  );
}

interface RecruitSubPanelProps {
  tile: Tile | undefined;
  isMine: boolean;
  onSubmit: (unit: UnitType, count: number) => void;
}

function RecruitSubPanel({ tile, isMine, onSubmit }: RecruitSubPanelProps) {
  const [counts, setCounts] = useState<Record<UnitType, number>>({
    infantry: 0,
    cavalry: 0,
    artillery: 0,
  });

  const increment = (u: UnitType) => setCounts((c) => ({ ...c, [u]: c[u] + 1 }));
  const decrement = (u: UnitType) => setCounts((c) => ({ ...c, [u]: Math.max(0, c[u] - 1) }));

  const totalCount = counts.infantry + counts.cavalry + counts.artillery;
  const canSubmit = tile !== undefined && isMine && totalCount > 0;

  const handleSubmit = () => {
    if (!canSubmit || tile === undefined) return;
    const units: UnitType[] = ['infantry', 'cavalry', 'artillery'];
    for (const u of units) {
      if (counts[u] > 0) onSubmit(u, counts[u]);
    }
    setCounts({ infantry: 0, cavalry: 0, artillery: 0 });
  };

  return (
    <div className="flex flex-col gap-2">
      <p className="text-[10px] text-slate-500">
        {tile === undefined
          ? 'Tap a tile on the board first.'
          : isMine
            ? `Recruit on tile (${tile.q},${tile.r})`
            : 'Selected tile is not yours.'}
      </p>
      <div className="flex flex-col gap-1">
        <RecruitRow
          unit="infantry"
          count={counts.infantry}
          onIncrement={() => increment('infantry')}
          onDecrement={() => decrement('infantry')}
        />
        <RecruitRow
          unit="cavalry"
          count={counts.cavalry}
          onIncrement={() => increment('cavalry')}
          onDecrement={() => decrement('cavalry')}
        />
        <RecruitRow
          unit="artillery"
          count={counts.artillery}
          onIncrement={() => increment('artillery')}
          onDecrement={() => decrement('artillery')}
        />
      </div>
      <button
        type="button"
        onClick={handleSubmit}
        disabled={!canSubmit}
        className="rounded bg-slate-900 px-3 py-1.5 text-xs font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
      >
        Recruit
      </button>
    </div>
  );
}

interface UpgradeSubPanelProps {
  tile: Tile | undefined;
  isMine: boolean;
  onUpgrade: (unitIndex: number) => void;
}

function UpgradeSubPanel({ tile, isMine, onUpgrade }: UpgradeSubPanelProps) {
  if (tile === undefined || !isMine) {
    return (
      <p className="text-[10px] text-slate-500">
        {tile === undefined ? 'Tap a tile on the board first.' : 'Selected tile is not yours.'}
      </p>
    );
  }
  if (tile.garrison.length === 0) {
    return <p className="text-[10px] text-slate-500">No units to upgrade on this tile.</p>;
  }
  return (
    <div className="flex flex-col gap-1">
      <p className="text-[10px] text-slate-500">Upgrade on tile ({tile.q},{tile.r})</p>
      <div className="flex flex-wrap gap-1">
        {tile.garrison.map((u, idx) => {
          const cost = upgradeCostFor(u);
          const disabled = cost === null;
          return (
            <button
              type="button"
              key={`${idx}-${u.type}-${u.level}`}
              onClick={() => onUpgrade(idx)}
              disabled={disabled}
              className="flex items-center gap-1 rounded border border-slate-200 bg-white px-2 py-1 text-xs disabled:cursor-not-allowed disabled:bg-slate-200 disabled:text-slate-400 hover:bg-slate-100"
            >
              <span>{UNIT_ICON[u.type]}</span>
              <span>L{u.level}</span>
              {cost !== null && (
                <span className="text-[10px] text-slate-500">
                  → L{u.level + 1} ({cost.gold}g·{cost.iron}i)
                </span>
              )}
              {cost === null && <span className="text-[10px] text-slate-400">MAX</span>}
            </button>
          );
        })}
      </div>
    </div>
  );
}

interface UnitChecklistProps {
  garrison: Unit[];
  selected: Set<number>;
  onToggle: (idx: number) => void;
}

function UnitChecklist({ garrison, selected, onToggle }: UnitChecklistProps) {
  if (garrison.length === 0) {
    return <p className="text-[10px] text-slate-500">No units on source tile.</p>;
  }
  return (
    <div className="flex flex-wrap gap-1">
      {garrison.map((u, idx) => {
        const isSelected = selected.has(idx);
        const power = UNIT_BASE_POWER[u.type] * u.level;
        return (
          <button
            type="button"
            key={`${idx}-${u.type}-${u.level}`}
            onClick={() => onToggle(idx)}
            className={`flex items-center gap-1 rounded border px-2 py-1 text-xs ${
              isSelected
                ? 'border-amber-500 bg-amber-100 text-amber-900'
                : 'border-slate-200 bg-white text-slate-700 hover:bg-slate-100'
            }`}
          >
            <span>{UNIT_ICON[u.type]}</span>
            <span>L{u.level}</span>
            <span className="text-[10px] text-slate-500">⚡{power}</span>
          </button>
        );
      })}
    </div>
  );
}

interface MoveAttackSubPanelProps {
  kind: 'move' | 'attack';
  source: { q: number; r: number } | null;
  sourceTile: Tile | undefined;
  selected: Set<number>;
  onToggle: (idx: number) => void;
  onCancel: () => void;
}

function MoveAttackSubPanel({
  kind,
  source,
  sourceTile,
  selected,
  onToggle,
  onCancel,
}: MoveAttackSubPanelProps) {
  return (
    <div className="flex flex-col gap-1">
      {source === null && (
        <p className="text-[10px] text-slate-500">
          Tap one of your tiles with units to choose a source.
        </p>
      )}
      {source !== null && sourceTile !== undefined && (
        <>
          <p className="text-[10px] text-slate-500">
            Source: ({source.q},{source.r}). Pick units, then tap a{' '}
            {kind === 'move' ? 'friendly destination' : 'target enemy/neutral tile'}.
          </p>
          <UnitChecklist garrison={sourceTile.garrison} selected={selected} onToggle={onToggle} />
        </>
      )}
      <button
        type="button"
        onClick={onCancel}
        className="self-start rounded border border-slate-300 bg-white px-2 py-1 text-xs text-slate-700 hover:bg-slate-100"
      >
        Cancel
      </button>
    </div>
  );
}

interface DiplomacySubPanelProps {
  source: { q: number; r: number } | null;
  onCancel: () => void;
}

function DiplomacySubPanel({ source, onCancel }: DiplomacySubPanelProps) {
  return (
    <div className="flex flex-col gap-1">
      <p className="text-[10px] text-slate-500">
        {source === null
          ? 'Tap one of your tiles to use as the source.'
          : `Source: (${source.q},${source.r}). Tap an enemy tile within 2 hexes to offer diplomacy.`}
      </p>
      <button
        type="button"
        onClick={onCancel}
        className="self-start rounded border border-slate-300 bg-white px-2 py-1 text-xs text-slate-700 hover:bg-slate-100"
      >
        Cancel
      </button>
    </div>
  );
}

interface BuySubPanelProps {
  selectedTile: Tile | undefined;
  onCancel: () => void;
}

function BuySubPanel({ selectedTile, onCancel }: BuySubPanelProps) {
  const cost =
    selectedTile !== undefined && selectedTile.ownerId === '' ? selectedTile.yield * 20 : null;
  return (
    <div className="flex flex-col gap-1">
      <p className="text-[10px] text-slate-500">
        Tap a neutral tile within 2 hexes of one of your tiles.
        {cost !== null && ` Cost preview: ${cost}g`}
      </p>
      <button
        type="button"
        onClick={onCancel}
        className="self-start rounded border border-slate-300 bg-white px-2 py-1 text-xs text-slate-700 hover:bg-slate-100"
      >
        Cancel
      </button>
    </div>
  );
}

export function ActionPanel({
  state,
  myPlayerId,
  isMyTurn,
  mode,
  onModeChange,
  selectedTile,
  selectedSourceForMove,
  selectedUnitIndices,
  onUnitIndexToggle,
  onAction,
}: ActionPanelProps) {
  const currentPlayer = useMemo(
    () => state.players.find((p) => p.id === state.currentTurn?.playerId) ?? null,
    [state.players, state.currentTurn],
  );

  if (!isMyTurn) {
    const name = currentPlayer?.name ?? '…';
    return (
      <div className="border-t border-slate-200 bg-white px-3 py-2 text-xs text-slate-500">
        Waiting for {name}…
      </div>
    );
  }

  const tile = selectedTile !== null ? findTile(state, selectedTile.q, selectedTile.r) : undefined;
  const isMine = tile !== undefined && myPlayerId !== null && tile.ownerId === myPlayerId;
  const sourceTile =
    selectedSourceForMove !== null
      ? findTile(state, selectedSourceForMove.q, selectedSourceForMove.r)
      : undefined;

  const handleEndTurn = () => {
    onAction({ type: 'end_turn' });
    onModeChange('none');
  };

  const handleRecruit = (unit: UnitType, count: number) => {
    if (selectedTile === null) return;
    onAction({
      type: 'recruit',
      q: selectedTile.q,
      r: selectedTile.r,
      unit,
      count,
    });
  };

  const handleUpgrade = (unitIndex: number) => {
    if (selectedTile === null) return;
    onAction({
      type: 'upgrade',
      q: selectedTile.q,
      r: selectedTile.r,
      unitIndex,
    });
  };

  const cancelToNone = () => onModeChange('none');

  return (
    <div className="flex flex-col gap-2 border-t border-slate-200 bg-slate-50 px-3 py-2">
      <div className="flex flex-wrap gap-2">
        {MODES.map((m) => (
          <ModeButton
            key={m.id}
            active={mode === m.id}
            disabled={false}
            label={m.label}
            onClick={() => onModeChange(mode === m.id ? 'none' : m.id)}
          />
        ))}
        <button
          type="button"
          onClick={handleEndTurn}
          className="ml-auto rounded bg-emerald-600 px-3 py-1 text-xs font-medium text-white hover:bg-emerald-700"
        >
          End Turn
        </button>
      </div>

      {mode === 'recruit' && (
        <RecruitSubPanel tile={tile} isMine={isMine} onSubmit={handleRecruit} />
      )}
      {mode === 'upgrade' && (
        <UpgradeSubPanel tile={tile} isMine={isMine} onUpgrade={handleUpgrade} />
      )}
      {(mode === 'move' || mode === 'attack') && (
        <MoveAttackSubPanel
          kind={mode}
          source={selectedSourceForMove}
          sourceTile={sourceTile}
          selected={selectedUnitIndices}
          onToggle={onUnitIndexToggle}
          onCancel={cancelToNone}
        />
      )}
      {mode === 'offer_diplomacy' && (
        <DiplomacySubPanel source={selectedSourceForMove} onCancel={cancelToNone} />
      )}
      {mode === 'buy' && <BuySubPanel selectedTile={tile} onCancel={cancelToNone} />}
    </div>
  );
}

export default ActionPanel;
