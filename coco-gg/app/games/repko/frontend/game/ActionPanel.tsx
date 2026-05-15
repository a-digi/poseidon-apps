import { useMemo, useState } from 'react';
import type {
  ClientAction,
  Civilization,
  GarrisonStack,
  ResourceType,
  StackPick,
  StateMsg,
  Tile,
  UnitType,
} from '../types';
import { hexDistance } from './coords';

export type SubAction =
  | 'inspect'
  | 'recruit'
  | 'upgrade'
  | 'move'
  | 'attack'
  | 'diplomacy'
  | 'buy'
  | 'upgrade_tile';

const BASE_POWER: Record<UnitType, number> = { infantry: 3, armor: 4, jet: 5 };

const UNIT_COSTS: Record<UnitType, { credits: number; steel: number; fuel: number }> = {
  infantry: { credits: 5, steel: 1, fuel: 0 },
  armor: { credits: 10, steel: 2, fuel: 0 },
  jet: { credits: 20, steel: 3, fuel: 0 },
};

const UNIT_ICON: Record<UnitType, string> = {
  infantry: '🪖',
  armor: '🚛',
  jet: '✈️',
};

const UNIT_NAME: Record<UnitType, string> = {
  infantry: 'Infantry',
  armor: 'Tank',
  jet: 'Jet',
};

const PROD_ICON: Record<string, string> = {
  credits: '💵',
  steel: '⚒',
  fuel: '⛽',
};

const PROD_NAME: Record<string, string> = {
  credits: 'Credits',
  steel: 'Steel',
  fuel: 'Fuel',
};

const RESOURCE_ICON: Record<ResourceType, string> = {
  credits: '💵',
  steel: '⚒',
  fuel: '⛽',
  none: '',
};

function productiveYields(
  yields: Partial<Record<ResourceType, number>> | undefined,
): { type: ResourceType; amount: number }[] {
  if (yields === undefined) return [];
  const out: { type: ResourceType; amount: number }[] = [];
  for (const t of ['credits', 'steel', 'fuel'] as const) {
    const a = yields[t] ?? 0;
    if (a > 0) out.push({ type: t, amount: a });
  }
  return out;
}

function stackPower(stack: GarrisonStack): number {
  return stack.count * BASE_POWER[stack.type] * stack.level;
}

function totalPower(stacks: GarrisonStack[]): number {
  return stacks.reduce((sum, s) => sum + stackPower(s), 0);
}

function totalUnits(stacks: GarrisonStack[]): number {
  return stacks.reduce((sum, s) => sum + s.count, 0);
}

function pickedPower(picks: StackPick[], garrison: GarrisonStack[]): number {
  return picks.reduce((sum, p) => {
    const s = garrison[p.stackIndex];
    if (s === undefined) return sum;
    return sum + p.count * BASE_POWER[s.type] * s.level;
  }, 0);
}

function diplomacyGoldCost(defenderGarrison: GarrisonStack[]): number {
  return totalPower(defenderGarrison) * 2;
}

function buyTileCost(tile: Tile): number {
  const total = productiveYields(tile.yields).reduce((s, y) => s + y.amount, 0);
  const cost = total * 3;
  if (cost < 3) return 3;
  if (cost > 15) return 15;
  return cost;
}

interface ButtonState {
  label: string;
  disabled: boolean;
}

function describeAttack(
  _tile: Tile,
  sources: Tile[],
): { state: ButtonState; source: Tile | null } {
  if (sources.length === 0) {
    return {
      state: {
        label: 'Attack — out of range (need an adjacent friendly tile)',
        disabled: true,
      },
      source: null,
    };
  }
  const source = sources[0];
  if (totalUnits(source.garrison) === 0) {
    return {
      state: {
        label: `Attack — no units at source (${source.q},${source.r})`,
        disabled: true,
      },
      source,
    };
  }
  const units = totalUnits(source.garrison);
  const power = totalPower(source.garrison);
  return {
    state: {
      label: `Attack — ${units} units / ${power} power (from (${source.q},${source.r}))`,
      disabled: false,
    },
    source,
  };
}

function describeBuy(tile: Tile, sources: Tile[], credits: number): ButtonState {
  const cost = buyTileCost(tile);
  if (sources.length === 0) {
    return {
      label: 'Buy this tile — out of range (need an adjacent friendly tile)',
      disabled: true,
    };
  }
  if (credits < cost) {
    return {
      label: `Buy this tile — costs ${cost}c (need ${cost - credits} more)`,
      disabled: true,
    };
  }
  return { label: `Buy this tile — costs ${cost}c`, disabled: false };
}

function describeDiplomacy(
  tile: Tile,
  sources: Tile[],
  credits: number,
  pending: boolean,
): ButtonState {
  if (pending) {
    return { label: 'Offer Diplomacy — offer pending', disabled: true };
  }
  if (sources.length === 0) {
    return {
      label: 'Offer Diplomacy — out of range (need an adjacent friendly tile)',
      disabled: true,
    };
  }
  const cost = diplomacyGoldCost(tile.garrison);
  if (credits < cost) {
    return {
      label: `Offer Diplomacy — costs ${cost}c (need ${cost - credits} more)`,
      disabled: true,
    };
  }
  return { label: `Offer Diplomacy — costs ${cost}c`, disabled: false };
}

function recruitCreditSteelCost(unit: UnitType, count: number): { credits: number; steel: number } {
  const c = UNIT_COSTS[unit];
  return { credits: c.credits * count, steel: c.steel * count };
}

function upgradeCreditSteelCost(stack: GarrisonStack): { credits: number; steel: number } | null {
  if (stack.level >= 3) return null;
  const c = UNIT_COSTS[stack.type];
  const mult = stack.level === 1 ? 1 : 2;
  return { credits: c.credits * mult, steel: c.steel * mult };
}

function findTile(state: StateMsg, q: number, r: number): Tile | undefined {
  return (state.board?.tiles ?? []).find((t) => t.q === q && t.r === r);
}

function findOwnerCiv(state: StateMsg, ownerId: string): Civilization | undefined {
  const player = state.players.find((p) => p.id === ownerId);
  if (player === undefined) return undefined;
  return (state.civilizations ?? []).find((c) => c.id === player.civilizationId);
}

function closestOwnedDistance(
  state: StateMsg,
  myPlayerId: string | null,
  tile: Tile,
): number | null {
  if (myPlayerId === null) return null;
  const owned = (state.board?.tiles ?? []).filter((t) => t.ownerId === myPlayerId);
  if (owned.length === 0) return null;
  let best = Number.POSITIVE_INFINITY;
  for (const t of owned) {
    const d = hexDistance({ q: t.q, r: t.r }, { q: tile.q, r: tile.r });
    if (d < best) best = d;
  }
  return best === Number.POSITIVE_INFINITY ? null : best;
}

function adjacentOwnedTiles(state: StateMsg, myPlayerId: string | null, target: Tile): Tile[] {
  if (myPlayerId === null) return [];
  const out: Tile[] = [];
  for (const t of state.board?.tiles ?? []) {
    if (t.ownerId !== myPlayerId) continue;
    if (t.q === target.q && t.r === target.r) continue;
    if (hexDistance({ q: t.q, r: t.r }, { q: target.q, r: target.r }) === 1) {
      out.push(t);
    }
  }
  return out;
}

function reachableSources(state: StateMsg, myPlayerId: string, target: Tile): Tile[] {
  return adjacentOwnedTiles(state, myPlayerId, target);
}

function sortSourcesByPreference(sources: Tile[]): Tile[] {
  return [...sources].sort((a, b) => {
    const ua = totalUnits(a.garrison);
    const ub = totalUnits(b.garrison);
    if (ub !== ua) return ub - ua;
    if (a.q !== b.q) return a.q - b.q;
    return a.r - b.r;
  });
}

interface ActionPanelProps {
  state: StateMsg;
  myPlayerId: string | null;
  isMyTurn: boolean;
  selectedTile: Tile | null;
  subAction: SubAction;
  onSubActionChange: (next: SubAction) => void;
  selectedStackCounts: Map<number, number>;
  onStackCountChange: (stackIndex: number, count: number) => void;
  attackSourceOverride: { q: number; r: number } | null;
  onAttackSourceChange: (src: { q: number; r: number } | null) => void;
  onAction: (action: ClientAction) => void;
}

interface InfoCardProps {
  state: StateMsg;
  myPlayerId: string | null;
  tile: Tile;
  noBorder?: boolean;
}

function rangeBadge(distance: number | null): { label: string; cls: string } {
  if (distance === null) return { label: 'No territory yet', cls: 'bg-slate-200 text-slate-700' };
  if (distance === 0) return { label: 'On your territory', cls: 'bg-emerald-100 text-emerald-800' };
  if (distance === 1) return { label: 'Adjacent', cls: 'bg-amber-100 text-amber-800' };
  return { label: 'Out of range', cls: 'bg-slate-200 text-slate-600' };
}

function TileInfoCard({ state, myPlayerId, tile, noBorder }: InfoCardProps) {
  const ownedSnapshot = (state.board?.tiles ?? [])
    .filter((t) => t.ownerId === myPlayerId)
    .map((t) => ({
      q: t.q,
      r: t.r,
      name: t.name,
      dist: hexDistance({ q: t.q, r: t.r }, { q: tile.q, r: tile.r }),
    }));
  const distances = ownedSnapshot.map((t) => t.dist);
  console.info('[repko/TileInfoCard] inspect', {
    myPlayerId,
    tile: { q: tile.q, r: tile.r, name: tile.name, ownerId: tile.ownerId },
    ownedCount: ownedSnapshot.length,
    owned: ownedSnapshot,
    closest: distances.length > 0 ? Math.min(...distances) : null,
    adjacentSources: adjacentOwnedTiles(state, myPlayerId, tile).map((s) => ({ q: s.q, r: s.r })),
  });
  const isMine = myPlayerId !== null && tile.ownerId === myPlayerId;
  const isNeutral = tile.ownerId === '';
  const buyCost = isNeutral ? buyTileCost(tile) : 0;
  const ownerCiv = !isNeutral ? findOwnerCiv(state, tile.ownerId) : undefined;
  const distance = closestOwnedDistance(state, myPlayerId, tile);
  const badge = rangeBadge(distance);

  const ownerLabel = isMine
    ? `${ownerCiv?.flag ?? '🏳'} You`
    : isNeutral
      ? 'Neutral'
      : `${ownerCiv?.flag ?? '🏳'} ${ownerCiv?.name ?? 'Unknown'}`;
  const ownerCls = isMine
    ? 'bg-emerald-100 text-emerald-800'
    : isNeutral
      ? 'bg-slate-200 text-slate-700'
      : 'bg-rose-100 text-rose-800';

  const prod = tile.production;
  const productionLine =
    prod === 'none'
      ? 'Barren (no production)'
      : `${PROD_ICON[prod] ?? ''} ${PROD_NAME[prod] ?? prod}`;

  const garrisonLine =
    tile.garrison.length === 0
      ? 'Garrison: empty'
      : tile.garrison
          .map((s) => `${UNIT_ICON[s.type]} L${s.level} ×${s.count}`)
          .join(', ');
  const garrisonTotals =
    tile.garrison.length > 0
      ? `Total: ${totalUnits(tile.garrison)} units (${totalPower(tile.garrison)} power)`
      : '';

  const yieldEntries = productiveYields(tile.yields);

  return (
    <div className={`${noBorder === true ? '' : 'border-b border-slate-200 '}px-3 py-2 text-xs text-slate-700`}>
      <div className="flex flex-wrap items-center gap-2">
        <div className="flex items-baseline gap-2">
          <span className="text-sm font-semibold text-slate-900">
            {tile.name !== undefined && tile.name !== '' ? tile.name : '—'}
          </span>
          <span className="font-mono text-xs text-slate-500">
            ({tile.q},{tile.r})
          </span>
        </div>
        <span className={`rounded px-1.5 py-0.5 text-[10px] font-medium ${ownerCls}`}>
          {ownerLabel}
        </span>
        <span className={`rounded px-1.5 py-0.5 text-[10px] font-medium ${badge.cls}`}>
          {badge.label}
        </span>
      </div>
      <p className="mt-1">{productionLine}</p>
      <p className="mt-1">
        <span className="font-medium text-slate-800">Garrison:</span> {garrisonLine}
      </p>
      {garrisonTotals !== '' && <p className="text-[11px] text-slate-500">{garrisonTotals}</p>}
      <div className="mt-1 flex items-center gap-2 text-xs text-slate-600">
        <span className="text-[10px] uppercase text-slate-400">Income</span>
        {yieldEntries.length === 0 ? (
          <span className="italic text-slate-400">Barren</span>
        ) : (
          yieldEntries.map(({ type, amount }) => (
            <span key={type} className="font-mono">
              {RESOURCE_ICON[type]} {amount}
            </span>
          ))
        )}
      </div>
      {isNeutral && (
        <p className="mt-1 text-xs text-slate-700">
          <span className="font-medium text-slate-800">Buy price:</span>{' '}
          <span className="font-mono">{buyCost}c</span>
        </p>
      )}
    </div>
  );
}

interface ActionButtonProps {
  label: string;
  disabled?: boolean;
  onClick: () => void;
}

function ActionButton({ label, disabled, onClick }: ActionButtonProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className="rounded-md bg-slate-900 px-3 py-2 text-sm font-medium text-white hover:bg-slate-700 disabled:cursor-not-allowed disabled:bg-slate-300"
    >
      {label}
    </button>
  );
}

function CancelButton({ onClick }: { onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="self-start rounded border border-slate-300 bg-white px-2 py-1 text-xs text-slate-700 hover:bg-slate-100"
    >
      Cancel
    </button>
  );
}

interface IconActionButtonProps {
  icon: string;
  label: string;
  disabled?: boolean;
  onClick: () => void;
}

function IconActionButton({ icon, label, disabled, onClick }: IconActionButtonProps) {
  return (
    <button
      type="button"
      aria-label={label}
      title={label}
      onClick={onClick}
      disabled={disabled}
      className="flex h-16 w-16 shrink-0 items-center justify-center rounded-md bg-slate-900 text-2xl text-white hover:bg-slate-700 disabled:cursor-not-allowed disabled:bg-slate-200 disabled:text-slate-400"
    >
      {icon}
    </button>
  );
}

interface Stepper {
  value: number;
  min: number;
  max: number;
  onChange: (next: number) => void;
}

function Stepper({ value, min, max, onChange }: Stepper) {
  return (
    <div className="flex items-center gap-1">
      <button
        type="button"
        onClick={() => onChange(Math.max(min, value - 1))}
        disabled={value <= min}
        className="rounded border border-slate-300 bg-white px-1.5 py-0.5 text-xs disabled:cursor-not-allowed disabled:bg-slate-200 disabled:text-slate-400"
      >
        −
      </button>
      <span className="w-7 text-center text-xs font-bold text-slate-900">{value}</span>
      <button
        type="button"
        onClick={() => onChange(Math.min(max, value + 1))}
        disabled={value >= max}
        className="rounded border border-slate-300 bg-white px-1.5 py-0.5 text-xs disabled:cursor-not-allowed disabled:bg-slate-200 disabled:text-slate-400 hover:bg-slate-100"
      >
        +
      </button>
      <button
        type="button"
        onClick={() => onChange(Math.min(max, value + 5))}
        disabled={value >= max}
        className="rounded border border-slate-300 bg-white px-1.5 py-0.5 text-xs disabled:cursor-not-allowed disabled:bg-slate-200 disabled:text-slate-400 hover:bg-slate-100"
      >
        +5
      </button>
      <button
        type="button"
        onClick={() => onChange(Math.min(max, value + 10))}
        disabled={value >= max}
        className="rounded border border-slate-300 bg-white px-1.5 py-0.5 text-xs disabled:cursor-not-allowed disabled:bg-slate-200 disabled:text-slate-400 hover:bg-slate-100"
      >
        +10
      </button>
    </div>
  );
}

interface RecruitSubPanelProps {
  tile: Tile;
  credits: number;
  steel: number;
  onConfirm: (unit: UnitType, count: number) => void;
  onCancel: () => void;
}

function RecruitSubPanel({ tile, credits, steel, onConfirm, onCancel }: RecruitSubPanelProps) {
  const [counts, setCounts] = useState<Record<UnitType, number>>({
    infantry: 0,
    armor: 0,
    jet: 0,
  });
  const setCount = (u: UnitType, next: number) =>
    setCounts((prev) => ({ ...prev, [u]: next }));

  const units: UnitType[] = ['infantry', 'armor', 'jet'];

  return (
    <div className="flex flex-col gap-2">
      <p className="text-[10px] text-slate-500">Recruit on tile ({tile.q},{tile.r})</p>
      {units.map((u) => {
        const cost = recruitCreditSteelCost(u, counts[u]);
        const unitCost = recruitCreditSteelCost(u, 1);
        const canConfirm = counts[u] > 0 && cost.credits <= credits && cost.steel <= steel;
        return (
          <div
            key={u}
            className="flex items-center justify-between gap-2 rounded border border-slate-200 bg-white px-2 py-1"
          >
            <div className="flex items-center gap-2 text-xs text-slate-800">
              <span className="text-base">{UNIT_ICON[u]}</span>
              <span>{UNIT_NAME[u]}</span>
            </div>
            <Stepper
              value={counts[u]}
              min={0}
              max={99}
              onChange={(n) => setCount(u, n)}
            />
            <span className="w-20 text-right text-[10px] text-slate-500">
              {unitCost.credits}c + {unitCost.steel}⚒
            </span>
            <button
              type="button"
              onClick={() => {
                if (!canConfirm) return;
                onConfirm(u, counts[u]);
              }}
              disabled={!canConfirm}
              className="rounded bg-slate-900 px-2 py-1 text-[11px] font-medium text-white hover:bg-slate-700 disabled:cursor-not-allowed disabled:bg-slate-300"
            >
              Confirm
            </button>
          </div>
        );
      })}
      <CancelButton onClick={onCancel} />
    </div>
  );
}

interface UpgradeSubPanelProps {
  tile: Tile;
  credits: number;
  steel: number;
  onUpgrade: (stackIndex: number) => void;
  onCancel: () => void;
}

function UpgradeSubPanel({ tile, credits, steel, onUpgrade, onCancel }: UpgradeSubPanelProps) {
  if (tile.garrison.length === 0) {
    return (
      <div className="flex flex-col gap-2">
        <p className="text-[10px] text-slate-500">No units to upgrade on this tile.</p>
        <CancelButton onClick={onCancel} />
      </div>
    );
  }
  return (
    <div className="flex flex-col gap-2">
      <p className="text-[10px] text-slate-500">Upgrade on tile ({tile.q},{tile.r})</p>
      <div className="flex flex-col gap-1">
        {tile.garrison.map((stack, idx) => {
          const cost = upgradeCreditSteelCost(stack);
          const affordable =
            cost !== null && cost.credits <= credits && cost.steel <= steel;
          const disabled = cost === null || !affordable;
          return (
            <div
              key={`${idx}-${stack.type}-${stack.level}`}
              className="flex items-center justify-between gap-2 rounded border border-slate-200 bg-white px-2 py-1"
            >
              <div className="flex items-center gap-2 text-xs text-slate-800">
                <span className="text-base">{UNIT_ICON[stack.type]}</span>
                <span>L{stack.level}</span>
                <span className="text-[10px] text-slate-500">×{stack.count}</span>
              </div>
              <button
                type="button"
                onClick={() => onUpgrade(idx)}
                disabled={disabled}
                className="rounded border border-slate-300 bg-white px-2 py-1 text-[10px] text-slate-700 hover:bg-slate-100 disabled:cursor-not-allowed disabled:bg-slate-200 disabled:text-slate-400"
              >
                {cost === null ? 'MAX' : `Upgrade — ${cost.credits}c + ${cost.steel}⚒`}
              </button>
            </div>
          );
        })}
      </div>
      <CancelButton onClick={onCancel} />
    </div>
  );
}

interface MoveSubPanelProps {
  tile: Tile;
  selectedStackCounts: Map<number, number>;
  onStackCountChange: (stackIndex: number, count: number) => void;
  onCancel: () => void;
}

function MoveSubPanel({
  tile,
  selectedStackCounts,
  onStackCountChange,
  onCancel,
}: MoveSubPanelProps) {
  let totalSelected = 0;
  let totalSelectedPower = 0;
  for (const [idx, c] of selectedStackCounts) {
    const stack = tile.garrison[idx];
    if (stack === undefined) continue;
    totalSelected += c;
    totalSelectedPower += c * BASE_POWER[stack.type] * stack.level;
  }
  return (
    <div className="flex flex-col gap-2">
      <p className="text-[10px] text-slate-500">
        Source: ({tile.q},{tile.r}). Pick units, then tap a friendly destination on the map.
      </p>
      {tile.garrison.length === 0 ? (
        <p className="text-[10px] text-slate-500">No units on this tile.</p>
      ) : (
        <div className="flex flex-col gap-1">
          {tile.garrison.map((stack, idx) => (
            <div
              key={`${idx}-${stack.type}-${stack.level}`}
              className="flex items-center justify-between gap-2 rounded border border-slate-200 bg-white px-2 py-1"
            >
              <div className="flex items-center gap-2 text-xs text-slate-800">
                <span className="text-base">{UNIT_ICON[stack.type]}</span>
                <span>L{stack.level}</span>
                <span className="text-[10px] text-slate-500">×{stack.count}</span>
              </div>
              <Stepper
                value={selectedStackCounts.get(idx) ?? 0}
                min={0}
                max={stack.count}
                onChange={(n) => onStackCountChange(idx, n)}
              />
            </div>
          ))}
        </div>
      )}
      <p className="text-[10px] text-slate-600">
        Selected: {totalSelected} units — power {totalSelectedPower}
      </p>
      <CancelButton onClick={onCancel} />
    </div>
  );
}

interface AttackSubPanelProps {
  target: Tile;
  sources: Tile[];
  sourceTile: Tile;
  selectedStackCounts: Map<number, number>;
  onStackCountChange: (stackIndex: number, count: number) => void;
  onAttackSourceChange: (src: { q: number; r: number } | null) => void;
  onConfirm: () => void;
  onCancel: () => void;
}

function AttackSubPanel({
  target,
  sources,
  sourceTile,
  selectedStackCounts,
  onStackCountChange,
  onAttackSourceChange,
  onConfirm,
  onCancel,
}: AttackSubPanelProps) {
  const [showSourceList, setShowSourceList] = useState(false);
  const defenderPower = totalPower(target.garrison);

  const picks: StackPick[] = [];
  for (const [idx, c] of selectedStackCounts) {
    if (c > 0) picks.push({ stackIndex: idx, count: c });
  }
  const attackerPower = pickedPower(picks, sourceTile.garrison);
  const canConfirm = picks.length > 0;

  return (
    <div className="flex flex-col gap-2">
      <p className="text-[10px] text-slate-700">
        <span className="font-medium">Target:</span> ({target.q},{target.r}) — Defender power: {defenderPower}
      </p>
      <div className="flex items-center gap-2 text-[10px] text-slate-700">
        <span>
          <span className="font-medium">Source:</span> ({sourceTile.q},{sourceTile.r})
        </span>
        {sources.length > 1 && (
          <button
            type="button"
            onClick={() => setShowSourceList((v) => !v)}
            className="rounded border border-slate-300 bg-white px-1.5 py-0.5 text-[10px] text-slate-700 hover:bg-slate-100"
          >
            {showSourceList ? 'hide sources' : 'change source'}
          </button>
        )}
      </div>
      {showSourceList && sources.length > 1 && (
        <div className="flex flex-wrap gap-1">
          {sources.map((s) => {
            const active = s.q === sourceTile.q && s.r === sourceTile.r;
            return (
              <button
                key={`${s.q},${s.r}`}
                type="button"
                onClick={() => {
                  onAttackSourceChange({ q: s.q, r: s.r });
                  setShowSourceList(false);
                }}
                className={`rounded border px-2 py-0.5 text-[10px] ${active ? 'border-amber-500 bg-amber-100 text-amber-900' : 'border-slate-300 bg-white text-slate-700 hover:bg-slate-100'}`}
              >
                ({s.q},{s.r}) — {totalUnits(s.garrison)}u
              </button>
            );
          })}
        </div>
      )}
      {sourceTile.garrison.length === 0 ? (
        <p className="text-[10px] text-slate-500">No units on the source tile.</p>
      ) : (
        <div className="flex flex-col gap-1">
          {sourceTile.garrison.map((stack, idx) => (
            <div
              key={`${idx}-${stack.type}-${stack.level}`}
              className="flex items-center justify-between gap-2 rounded border border-slate-200 bg-white px-2 py-1"
            >
              <div className="flex items-center gap-2 text-xs text-slate-800">
                <span className="text-base">{UNIT_ICON[stack.type]}</span>
                <span>L{stack.level}</span>
                <span className="text-[10px] text-slate-500">×{stack.count}</span>
              </div>
              <Stepper
                value={selectedStackCounts.get(idx) ?? 0}
                min={0}
                max={stack.count}
                onChange={(n) => onStackCountChange(idx, n)}
              />
            </div>
          ))}
        </div>
      )}
      <p className="text-[10px] text-slate-700">
        Attacker power: {attackerPower} | Defender power: {defenderPower}
      </p>
      <div className="flex items-center gap-2">
        <ActionButton
          label="Confirm attack"
          disabled={!canConfirm}
          onClick={() => {
            if (!canConfirm) return;
            onConfirm();
          }}
        />
        <CancelButton onClick={onCancel} />
      </div>
    </div>
  );
}

interface DiplomacySubPanelProps {
  target: Tile;
  credits: number;
  pending: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

function DiplomacySubPanel({ target, credits, pending, onConfirm, onCancel }: DiplomacySubPanelProps) {
  const defenderPower = totalPower(target.garrison);
  const cost = diplomacyGoldCost(target.garrison);
  const canConfirm = !pending && cost <= credits;
  return (
    <div className="flex flex-col gap-2">
      <p className="text-[10px] text-slate-700">
        Defender power: {defenderPower} | Cost: {defenderPower} × 2 = {cost} credits
      </p>
      <p className="text-[10px] text-slate-500">
        Note: paid when the offer is made, not refunded if declined.
      </p>
      {pending && (
        <p className="text-[10px] text-amber-700">An offer is already pending on this tile.</p>
      )}
      <div className="flex items-center gap-2">
        <ActionButton label="Confirm offer" disabled={!canConfirm} onClick={onConfirm} />
        <CancelButton onClick={onCancel} />
      </div>
    </div>
  );
}

interface IconActionCellProps {
  icon: string;
  label: string;
  caption: string;
  disabled?: boolean;
  onClick: () => void;
}

function IconActionCell({ icon, label, caption, disabled, onClick }: IconActionCellProps) {
  return (
    <div className="flex flex-col items-center gap-1">
      <IconActionButton icon={icon} label={label} disabled={disabled} onClick={onClick} />
      <span className="text-[10px] font-medium text-slate-700">{caption}</span>
    </div>
  );
}

interface ContextualActionsProps {
  state: StateMsg;
  myPlayerId: string;
  tile: Tile;
  credits: number;
  onSubActionChange: (next: SubAction) => void;
  onAttackSourceChange: (src: { q: number; r: number } | null) => void;
}

function ContextualActions({
  state,
  myPlayerId,
  tile,
  credits,
  onSubActionChange,
  onAttackSourceChange,
}: ContextualActionsProps) {
  const ownedSnapshot = (state.board?.tiles ?? [])
    .filter((t) => t.ownerId === myPlayerId)
    .map((t) => ({
      q: t.q,
      r: t.r,
      name: t.name,
      dist: hexDistance({ q: t.q, r: t.r }, { q: tile.q, r: tile.r }),
    }));
  const distances = ownedSnapshot.map((t) => t.dist);
  console.info('[repko/ContextualActions] inspect', {
    myPlayerId,
    tile: { q: tile.q, r: tile.r, name: tile.name, ownerId: tile.ownerId },
    ownedCount: ownedSnapshot.length,
    owned: ownedSnapshot,
    closest: distances.length > 0 ? Math.min(...distances) : null,
    adjacentSources: adjacentOwnedTiles(state, myPlayerId, tile).map((s) => ({ q: s.q, r: s.r })),
  });
  const isMine = tile.ownerId === myPlayerId;
  const isNeutral = tile.ownerId === '';
  const isEnemy = !isMine && !isNeutral;

  if (isMine) {
    const canUpgradeTile = tile.production !== 'none' && credits >= 10;
    const hasGarrison = tile.garrison.length > 0;
    return (
      <div className="grid grid-cols-2 gap-1">
        <IconActionButton icon="➕" label="Recruit" onClick={() => onSubActionChange('recruit')} />
        <IconActionButton
          icon="⬆️"
          label="Upgrade units"
          disabled={!hasGarrison}
          onClick={() => onSubActionChange('upgrade')}
        />
        <IconActionButton
          icon="🔀"
          label="Move units from here"
          disabled={!hasGarrison}
          onClick={() => onSubActionChange('move')}
        />
        <IconActionButton
          icon="⚙️"
          label={`Upgrade production — costs 10c${credits < 10 ? ` (need ${10 - credits} more)` : ''}`}
          disabled={!canUpgradeTile}
          onClick={() => onSubActionChange('upgrade_tile')}
        />
      </div>
    );
  }

  const sources = sortSourcesByPreference(reachableSources(state, myPlayerId, tile));

  if (isEnemy) {
    const { state: attackBtn, source } = describeAttack(tile, sources);
    const pending =
      state.pendingDiplomacy?.some((o) => o.q === tile.q && o.r === tile.r) ?? false;
    const dipBtn = describeDiplomacy(tile, sources, credits, pending);
    const defenderUnits = totalUnits(tile.garrison);
    const defenderPower = totalPower(tile.garrison);
    const attackCaption = `🛡 ${defenderUnits}u / ${defenderPower}p`;
    return (
      <div className="flex flex-row gap-2">
        <IconActionCell
          icon="⚔️"
          label={attackBtn.label}
          caption={attackCaption}
          disabled={attackBtn.disabled}
          onClick={() => {
            if (source !== null) {
              onAttackSourceChange({ q: source.q, r: source.r });
            }
            onSubActionChange('attack');
          }}
        />
        <IconActionCell
          icon="🤝"
          label={dipBtn.label}
          caption={`${diplomacyGoldCost(tile.garrison)}C`}
          disabled={dipBtn.disabled}
          onClick={() => onSubActionChange('diplomacy')}
        />
      </div>
    );
  }

  if (isNeutral) {
    const { state: attackBtn, source } = describeAttack(tile, sources);
    const buyBtn = describeBuy(tile, sources, credits);
    const defenderUnits = totalUnits(tile.garrison);
    const defenderPower = totalPower(tile.garrison);
    const attackCaption = `🛡 ${defenderUnits}u / ${defenderPower}p`;
    return (
      <div className="flex flex-row gap-2">
        <IconActionCell
          icon="⚔️"
          label={attackBtn.label}
          caption={attackCaption}
          disabled={attackBtn.disabled}
          onClick={() => {
            if (source !== null) {
              onAttackSourceChange({ q: source.q, r: source.r });
            }
            onSubActionChange('attack');
          }}
        />
        <IconActionCell
          icon="💵"
          label={buyBtn.label}
          caption={`${buyTileCost(tile)}C`}
          disabled={buyBtn.disabled}
          onClick={() => onSubActionChange('buy')}
        />
      </div>
    );
  }

  return <p className="text-xs italic text-slate-500">No actions available for this tile.</p>;
}

export function ActionPanel({
  state,
  myPlayerId,
  isMyTurn,
  selectedTile,
  subAction,
  onSubActionChange,
  selectedStackCounts,
  onStackCountChange,
  attackSourceOverride,
  onAttackSourceChange,
  onAction,
}: ActionPanelProps) {
  const currentPlayer = useMemo(
    () => state.players.find((p) => p.id === state.currentTurn?.playerId) ?? null,
    [state.players, state.currentTurn],
  );

  if (!isMyTurn) {
    const name = currentPlayer?.name ?? '…';
    return (
      <footer className="border-t border-slate-200 bg-white px-3 py-2 text-xs text-slate-500">
        Waiting for {name}…
      </footer>
    );
  }

  if (selectedTile === null || myPlayerId === null) {
    return (
      <footer className="border-t border-slate-200 bg-white px-3 py-2 text-xs text-slate-500">
        Tap a tile to begin.
      </footer>
    );
  }

  const credits = state.you?.resources?.credits ?? 0;
  const steel = state.you?.resources?.steel ?? 0;

  const resetToInspect = () => onSubActionChange('inspect');

  const sources = sortSourcesByPreference(reachableSources(state, myPlayerId, selectedTile));
  const attackSource =
    attackSourceOverride !== null
      ? findTile(state, attackSourceOverride.q, attackSourceOverride.r) ?? sources[0]
      : sources[0];

  const pendingOnThisTile =
    state.pendingDiplomacy?.some((o) => o.q === selectedTile.q && o.r === selectedTile.r) ?? false;

  const handleRecruit = (unit: UnitType, count: number) => {
    onAction({ type: 'recruit', q: selectedTile.q, r: selectedTile.r, unit, count });
    resetToInspect();
  };

  const handleUpgrade = (stackIndex: number) => {
    onAction({ type: 'upgrade', q: selectedTile.q, r: selectedTile.r, stackIndex });
    resetToInspect();
  };

  const handleAttackConfirm = () => {
    if (attackSource === undefined) return;
    const units: StackPick[] = [];
    for (const [idx, c] of selectedStackCounts) {
      if (c > 0) units.push({ stackIndex: idx, count: c });
    }
    if (units.length === 0) return;
    onAction({
      type: 'attack',
      fromQ: attackSource.q,
      fromR: attackSource.r,
      toQ: selectedTile.q,
      toR: selectedTile.r,
      units,
    });
    resetToInspect();
  };

  const handleDiplomacyConfirm = () => {
    onAction({ type: 'offer_diplomacy', q: selectedTile.q, r: selectedTile.r });
    resetToInspect();
  };

  const isNonOwned = selectedTile.ownerId !== myPlayerId;
  const isNonOwnedInspect = isNonOwned && subAction === 'inspect';
  const isOwnedInspect = !isNonOwned && subAction === 'inspect';
  const isNonOwnedSimpleConfirm = isNonOwned && (subAction === 'buy' || subAction === 'diplomacy');
  const useInlineLayout = isNonOwnedInspect || isNonOwnedSimpleConfirm || isOwnedInspect;

  const handleInlineConfirm = () => {
    if (subAction === 'buy') {
      onAction({ type: 'buy_tile', q: selectedTile.q, r: selectedTile.r });
      resetToInspect();
    } else if (subAction === 'diplomacy') {
      handleDiplomacyConfirm();
    }
  };

  return (
    <footer className="border-t border-slate-200 bg-slate-50">
      {useInlineLayout ? (
        <div className="flex items-stretch border-b border-slate-200">
          <div className="min-w-0 flex-1">
            <TileInfoCard state={state} myPlayerId={myPlayerId} tile={selectedTile} noBorder />
            {isNonOwnedSimpleConfirm && (
              <p className="px-3 pb-2 text-xs text-slate-600">
                {subAction === 'buy'
                  ? `Purchase for ${buyTileCost(selectedTile)} credits?`
                  : `Offer diplomacy? Cost: ${diplomacyGoldCost(selectedTile.garrison)} credits`}
              </p>
            )}
          </div>
          <div className="flex shrink-0 items-center justify-center gap-2 border-l border-slate-200 px-2 py-2">
            {(isNonOwnedInspect || isOwnedInspect) && (
              <ContextualActions
                state={state}
                myPlayerId={myPlayerId}
                tile={selectedTile}
                credits={credits}
                onSubActionChange={onSubActionChange}
                onAttackSourceChange={onAttackSourceChange}
              />
            )}
            {isNonOwnedSimpleConfirm && (
              <div className="flex flex-col gap-2">
                <button
                  type="button"
                  aria-label="Confirm"
                  onClick={handleInlineConfirm}
                  className="flex h-16 w-16 shrink-0 items-center justify-center rounded-md bg-emerald-700 text-2xl text-white hover:bg-emerald-600"
                >
                  ✓
                </button>
                <button
                  type="button"
                  aria-label="Cancel"
                  onClick={resetToInspect}
                  className="flex h-16 w-16 shrink-0 items-center justify-center rounded-md bg-slate-200 text-xl text-slate-700 hover:bg-slate-300"
                >
                  ✗
                </button>
              </div>
            )}
          </div>
        </div>
      ) : (
        <TileInfoCard state={state} myPlayerId={myPlayerId} tile={selectedTile} />
      )}
      <div className="flex flex-col gap-2 px-3 py-2">
        {subAction === 'inspect' && !useInlineLayout && (
          <ContextualActions
            state={state}
            myPlayerId={myPlayerId}
            tile={selectedTile}
            credits={credits}
            onSubActionChange={onSubActionChange}
            onAttackSourceChange={onAttackSourceChange}
          />
        )}
        {subAction === 'recruit' && (
          <RecruitSubPanel
            tile={selectedTile}
            credits={credits}
            steel={steel}
            onConfirm={handleRecruit}
            onCancel={resetToInspect}
          />
        )}
        {subAction === 'upgrade' && (
          <UpgradeSubPanel
            tile={selectedTile}
            credits={credits}
            steel={steel}
            onUpgrade={handleUpgrade}
            onCancel={resetToInspect}
          />
        )}
        {subAction === 'move' && (
          <MoveSubPanel
            tile={selectedTile}
            selectedStackCounts={selectedStackCounts}
            onStackCountChange={onStackCountChange}
            onCancel={resetToInspect}
          />
        )}
        {subAction === 'attack' && attackSource !== undefined && (
          <AttackSubPanel
            target={selectedTile}
            sources={sources}
            sourceTile={attackSource}
            selectedStackCounts={selectedStackCounts}
            onStackCountChange={onStackCountChange}
            onAttackSourceChange={onAttackSourceChange}
            onConfirm={handleAttackConfirm}
            onCancel={resetToInspect}
          />
        )}
        {subAction === 'attack' && attackSource === undefined && (
          <div className="flex flex-col gap-2">
            <p className="text-xs italic text-slate-500">No friendly source in range.</p>
            <CancelButton onClick={resetToInspect} />
          </div>
        )}
        {subAction === 'diplomacy' && !isNonOwnedSimpleConfirm && (
          <DiplomacySubPanel
            target={selectedTile}
            credits={credits}
            pending={pendingOnThisTile}
            onConfirm={handleDiplomacyConfirm}
            onCancel={resetToInspect}
          />
        )}
        {subAction === 'buy' && !isNonOwnedSimpleConfirm && (
          <div className="flex flex-col gap-2">
            <p className="text-xs text-slate-700">
              Confirm purchase of{' '}
              <span className="font-semibold">
                {selectedTile.name !== undefined && selectedTile.name !== ''
                  ? selectedTile.name
                  : `(${selectedTile.q},${selectedTile.r})`}
              </span>{' '}
              for {buyTileCost(selectedTile)} credits.
            </p>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => {
                  onAction({ type: 'buy_tile', q: selectedTile.q, r: selectedTile.r });
                  resetToInspect();
                }}
                className="rounded-md bg-slate-900 px-3 py-2 text-sm font-medium text-white hover:bg-slate-700"
              >
                Confirm
              </button>
              <button
                type="button"
                onClick={resetToInspect}
                className="rounded-md bg-slate-200 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-300"
              >
                Cancel
              </button>
            </div>
          </div>
        )}
        {subAction === 'upgrade_tile' && (
          <div className="flex flex-col gap-2">
            <p className="text-xs text-slate-700">
              Upgrade production of{' '}
              <span className="font-semibold">
                {selectedTile.name !== undefined && selectedTile.name !== ''
                  ? selectedTile.name
                  : `(${selectedTile.q},${selectedTile.r})`}
              </span>{' '}
              for <span className="font-semibold">10 credits</span>.
            </p>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => {
                  onAction({ type: 'upgrade_tile', q: selectedTile.q, r: selectedTile.r });
                  resetToInspect();
                }}
                className="rounded-md bg-slate-900 px-3 py-2 text-sm font-medium text-white hover:bg-slate-700"
              >
                Confirm
              </button>
              <button
                type="button"
                onClick={resetToInspect}
                className="rounded-md bg-slate-200 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-300"
              >
                Cancel
              </button>
            </div>
          </div>
        )}
      </div>
    </footer>
  );
}

export { reachableSources, sortSourcesByPreference };

export default ActionPanel;
