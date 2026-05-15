import type { GarrisonStack, StateMsg, Tile, UnitType } from '../types';
import { UNIT_CATALOG, UNIT_ICON, UNIT_NAME } from './units';

export const TILE_SCORE_MULTIPLIER = 5;

export function tileScore(tile: Tile): number {
  let yields = 0;
  for (const v of Object.values(tile.yields ?? {})) yields += v;
  return yields * TILE_SCORE_MULTIPLIER;
}

export function garrisonScore(stack: GarrisonStack): number {
  const spec = UNIT_CATALOG[stack.type];
  if (spec === undefined) return 0;
  return spec.power * stack.level * stack.count;
}

export function playerScore(state: StateMsg, playerId: string): number {
  let total = 0;
  for (const t of state.board?.tiles ?? []) {
    if (t.ownerId !== playerId) continue;
    total += tileScore(t);
    for (const s of t.garrison) total += garrisonScore(s);
  }
  for (const army of state.armies ?? []) {
    if (army.ownerId !== playerId) continue;
    for (const s of army.units) total += garrisonScore(s);
  }
  return total;
}

export interface TileBreakdown {
  q: number;
  r: number;
  name?: string;
  yieldTotal: number;
  score: number;
}

export interface UnitBreakdown {
  type: UnitType;
  icon: string;
  name: string;
  level: number;
  count: number;
  power: number;
  subtotal: number;
}

export interface PlayerBreakdown {
  tiles: TileBreakdown[];
  units: UnitBreakdown[];
  tileScore: number;
  unitScore: number;
  total: number;
}

export function scoreBreakdown(state: StateMsg, playerId: string): PlayerBreakdown {
  const tiles: TileBreakdown[] = [];
  const unitAgg = new Map<string, { type: UnitType; level: number; count: number }>();
  let totalTile = 0;
  let totalUnit = 0;
  for (const t of state.board?.tiles ?? []) {
    if (t.ownerId !== playerId) continue;
    let yieldTotal = 0;
    for (const v of Object.values(t.yields ?? {})) yieldTotal += v;
    const score = yieldTotal * TILE_SCORE_MULTIPLIER;
    tiles.push({ q: t.q, r: t.r, name: t.name, yieldTotal, score });
    totalTile += score;
    for (const s of t.garrison) {
      const key = `${s.type}-${s.level}`;
      const entry = unitAgg.get(key);
      if (entry === undefined) {
        unitAgg.set(key, { type: s.type, level: s.level, count: s.count });
      } else {
        entry.count += s.count;
      }
    }
  }
  for (const army of state.armies ?? []) {
    if (army.ownerId !== playerId) continue;
    for (const s of army.units) {
      const key = `${s.type}-${s.level}`;
      const entry = unitAgg.get(key);
      if (entry === undefined) {
        unitAgg.set(key, { type: s.type, level: s.level, count: s.count });
      } else {
        entry.count += s.count;
      }
    }
  }
  const units: UnitBreakdown[] = [];
  for (const { type, level, count } of unitAgg.values()) {
    const spec = UNIT_CATALOG[type];
    if (spec === undefined) continue;
    const power = spec.power * level;
    const subtotal = power * count;
    totalUnit += subtotal;
    units.push({
      type,
      icon: UNIT_ICON[type],
      name: UNIT_NAME[type],
      level,
      count,
      power,
      subtotal,
    });
  }
  tiles.sort((a, b) => b.score - a.score);
  units.sort((a, b) => b.subtotal - a.subtotal);
  return {
    tiles,
    units,
    tileScore: totalTile,
    unitScore: totalUnit,
    total: totalTile + totalUnit,
  };
}
