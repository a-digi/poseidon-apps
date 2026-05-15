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
import { hexDistance, reachableForAction } from './coords';

export type SubAction =
  | 'inspect'
  | 'recruit'
  | 'upgrade'
  | 'move'
  | 'attack'
  | 'diplomacy'
  | 'buy';

const BASE_POWER: Record<UnitType, number> = { infantry: 1, cavalry: 2, artillery: 3 };

const UNIT_COSTS: Record<UnitType, { gold: number; iron: number; food: number }> = {
  infantry: { gold: 5, iron: 1, food: 0 },
  cavalry: { gold: 10, iron: 2, food: 0 },
  artillery: { gold: 20, iron: 3, food: 0 },
};

const UNIT_ICON: Record<UnitType, string> = {
  infantry: '⚔',
  cavalry: '🐎',
  artillery: '💥',
};

const UNIT_NAME: Record<UnitType, string> = {
  infantry: 'Infantry',
  cavalry: 'Cavalry',
  artillery: 'Artillery',
};

const PROD_ICON: Record<string, string> = {
  gold: '💰',
  iron: '⚒',
  food: '🍞',
};

const PROD_NAME: Record<string, string> = {
  gold: 'Gold',
  iron: 'Iron',
  food: 'Food',
};

const RESOURCE_ICON: Record<ResourceType, string> = {
  gold: '💰',
  iron: '⚒',
  food: '🍞',
  none: '',
};

function productiveYields(
  yields: Partial<Record<ResourceType, number>> | undefined,
): { type: ResourceType; amount: number }[] {
  if (yields === undefined) return [];
  const out: { type: ResourceType; amount: number }[] = [];
  for (const t of ['gold', 'iron', 'food'] as const) {
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

function describeBuy(tile: Tile, sources: Tile[], gold: number): ButtonState {
  const total = productiveYields(tile.yields).reduce((s, y) => s + y.amount, 0);
  const cost = total * 20;
  if (sources.length === 0) {
    return {
      label: 'Buy this tile — out of range (need an adjacent friendly tile)',
      disabled: true,
    };
  }
  if (gold < cost) {
    return {
      label: `Buy this tile — costs ${cost}g (need ${cost - gold} more)`,
      disabled: true,
    };
  }
  return { label: `Buy this tile — costs ${cost}g`, disabled: false };
}

function describeDiplomacy(
  tile: Tile,
  sources: Tile[],
  gold: number,
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
  if (gold < cost) {
    return {
      label: `Offer Diplomacy — costs ${cost}g (need ${cost - gold} more)`,
      disabled: true,
    };
  }
  return { label: `Offer Diplomacy — costs ${cost}g`, disabled: false };
}

function recruitGoldIronCost(unit: UnitType, count: number): { gold: number; iron: number } {
  const c = UNIT_COSTS[unit];
  return { gold: c.gold * count, iron: c.iron * count };
}

function upgradeGoldIronCost(stack: GarrisonStack): { gold: number; iron: number } | null {
  if (stack.level >= 3) return null;
  const c = UNIT_COSTS[stack.type];
  const mult = stack.level === 1 ? 1 : 2;
  return { gold: c.gold * mult, iron: c.iron * mult };
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

function reachableSources(state: StateMsg, myPlayerId: string, target: Tile): Tile[] {
  const tiles = state.board?.tiles ?? [];
  const owned = tiles.filter((t) => t.ownerId === myPlayerId);
  const out: Tile[] = [];
  for (const src of owned) {
    if (src.q === target.q && src.r === target.r) continue;
    const reach = reachableForAction(state, myPlayerId, { q: src.q, r: src.r });
    if (reach.some((h) => h.q === target.q && h.r === target.r)) out.push(src);
  }
  return out;
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
}

function rangeBadge(distance: number | null): { label: string; cls: string } {
  if (distance === null) return { label: 'No territory yet', cls: 'bg-slate-200 text-slate-700' };
  if (distance === 0) return { label: 'On your territory', cls: 'bg-emerald-100 text-emerald-800' };
  if (distance === 1) return { label: 'Adjacent', cls: 'bg-amber-100 text-amber-800' };
  return { label: 'Out of range', cls: 'bg-slate-200 text-slate-600' };
}

function TileInfoCard({ state, myPlayerId, tile }: InfoCardProps) {
  const isMine = myPlayerId !== null && tile.ownerId === myPlayerId;
  const isNeutral = tile.ownerId === '';
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
    <div className="border-b border-slate-200 px-3 py-2 text-xs text-slate-700">
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
    </div>
  );
}

interface RecruitSubPanelProps {
  tile: Tile;
  gold: number;
  iron: number;
  onConfirm: (unit: UnitType, count: number) => void;
  onCancel: () => void;
}

function RecruitSubPanel({ tile, gold, iron, onConfirm, onCancel }: RecruitSubPanelProps) {
  const [counts, setCounts] = useState<Record<UnitType, number>>({
    infantry: 0,
    cavalry: 0,
    artillery: 0,
  });
  const setCount = (u: UnitType, next: number) =>
    setCounts((prev) => ({ ...prev, [u]: next }));

  const units: UnitType[] = ['infantry', 'cavalry', 'artillery'];

  return (
    <div className="flex flex-col gap-2">
      <p className="text-[10px] text-slate-500">Recruit on tile ({tile.q},{tile.r})</p>
      {units.map((u) => {
        const cost = recruitGoldIronCost(u, counts[u]);
        const unitCost = recruitGoldIronCost(u, 1);
        const canConfirm = counts[u] > 0 && cost.gold <= gold && cost.iron <= iron;
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
              {unitCost.gold}g + {unitCost.iron}⚒
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
  gold: number;
  iron: number;
  onUpgrade: (stackIndex: number) => void;
  onCancel: () => void;
}

function UpgradeSubPanel({ tile, gold, iron, onUpgrade, onCancel }: UpgradeSubPanelProps) {
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
          const cost = upgradeGoldIronCost(stack);
          const affordable =
            cost !== null && cost.gold <= gold && cost.iron <= iron;
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
                {cost === null ? 'MAX' : `Upgrade — ${cost.gold}g + ${cost.iron}⚒`}
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
  gold: number;
  pending: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

function DiplomacySubPanel({ target, gold, pending, onConfirm, onCancel }: DiplomacySubPanelProps) {
  const defenderPower = totalPower(target.garrison);
  const cost = diplomacyGoldCost(target.garrison);
  const canConfirm = !pending && cost <= gold;
  return (
    <div className="flex flex-col gap-2">
      <p className="text-[10px] text-slate-700">
        Defender power: {defenderPower} | Cost: {defenderPower} × 2 = {cost} gold
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

interface ContextualActionsProps {
  state: StateMsg;
  myPlayerId: string;
  tile: Tile;
  gold: number;
  onSubActionChange: (next: SubAction) => void;
  onAttackSourceChange: (src: { q: number; r: number } | null) => void;
}

function ContextualActions({
  state,
  myPlayerId,
  tile,
  gold,
  onSubActionChange,
  onAttackSourceChange,
}: ContextualActionsProps) {
  const isMine = tile.ownerId === myPlayerId;
  const isNeutral = tile.ownerId === '';
  const isEnemy = !isMine && !isNeutral;

  if (isMine) {
    return (
      <div className="flex flex-col gap-2">
        <ActionButton label="Recruit" onClick={() => onSubActionChange('recruit')} />
        {tile.garrison.length > 0 && (
          <>
            <ActionButton label="Upgrade" onClick={() => onSubActionChange('upgrade')} />
            <ActionButton
              label="Move units from here"
              onClick={() => onSubActionChange('move')}
            />
          </>
        )}
      </div>
    );
  }

  const sources = sortSourcesByPreference(reachableSources(state, myPlayerId, tile));

  if (isEnemy) {
    const { state: attackBtn, source } = describeAttack(tile, sources);
    const pending =
      state.pendingDiplomacy?.some((o) => o.q === tile.q && o.r === tile.r) ?? false;
    const dipBtn = describeDiplomacy(tile, sources, gold, pending);
    return (
      <div className="flex flex-col gap-2">
        <ActionButton
          label={attackBtn.label}
          disabled={attackBtn.disabled}
          onClick={() => {
            if (source !== null) {
              onAttackSourceChange({ q: source.q, r: source.r });
            }
            onSubActionChange('attack');
          }}
        />
        <ActionButton
          label={dipBtn.label}
          disabled={dipBtn.disabled}
          onClick={() => onSubActionChange('diplomacy')}
        />
      </div>
    );
  }

  if (isNeutral) {
    const { state: attackBtn, source } = describeAttack(tile, sources);
    const buyBtn = describeBuy(tile, sources, gold);
    return (
      <div className="flex flex-col gap-2">
        <ActionButton
          label={attackBtn.label}
          disabled={attackBtn.disabled}
          onClick={() => {
            if (source !== null) {
              onAttackSourceChange({ q: source.q, r: source.r });
            }
            onSubActionChange('attack');
          }}
        />
        <ActionButton
          label={buyBtn.label}
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

  const gold = state.you?.resources?.gold ?? 0;
  const iron = state.you?.resources?.iron ?? 0;

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

  return (
    <footer className="border-t border-slate-200 bg-slate-50">
      <TileInfoCard state={state} myPlayerId={myPlayerId} tile={selectedTile} />
      <div className="flex flex-col gap-2 px-3 py-2">
        {subAction === 'inspect' && (
          <ContextualActions
            state={state}
            myPlayerId={myPlayerId}
            tile={selectedTile}
            gold={gold}
            onSubActionChange={onSubActionChange}
            onAttackSourceChange={onAttackSourceChange}
          />
        )}
        {subAction === 'recruit' && (
          <RecruitSubPanel
            tile={selectedTile}
            gold={gold}
            iron={iron}
            onConfirm={handleRecruit}
            onCancel={resetToInspect}
          />
        )}
        {subAction === 'upgrade' && (
          <UpgradeSubPanel
            tile={selectedTile}
            gold={gold}
            iron={iron}
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
        {subAction === 'diplomacy' && (
          <DiplomacySubPanel
            target={selectedTile}
            gold={gold}
            pending={pendingOnThisTile}
            onConfirm={handleDiplomacyConfirm}
            onCancel={resetToInspect}
          />
        )}
        {subAction === 'buy' && (
          <div className="flex flex-col gap-2">
            <p className="text-xs text-slate-700">
              Confirm purchase of{' '}
              <span className="font-semibold">
                {selectedTile.name !== undefined && selectedTile.name !== ''
                  ? selectedTile.name
                  : `(${selectedTile.q},${selectedTile.r})`}
              </span>{' '}
              for {productiveYields(selectedTile.yields).reduce((s, y) => s + y.amount, 0) * 20} gold.
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
      </div>
    </footer>
  );
}

export { reachableSources, sortSourcesByPreference };

export default ActionPanel;
