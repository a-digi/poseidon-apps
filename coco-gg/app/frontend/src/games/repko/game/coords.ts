import type { StateMsg, Tile } from '../types';

export interface HexCoord {
  q: number;
  r: number;
}

const NEIGHBOR_OFFSETS: ReadonlyArray<readonly [number, number]> = [
  [1, 0],
  [1, -1],
  [0, -1],
  [-1, 0],
  [-1, 1],
  [0, 1],
];

export function hexToPixel(q: number, r: number, size: number): { x: number; y: number } {
  const x = size * Math.sqrt(3) * (q + r / 2);
  const y = size * 1.5 * r;
  return { x, y };
}

export function hexCorner(cx: number, cy: number, size: number, i: number): { x: number; y: number } {
  const angleDeg = 60 * i - 90;
  const angleRad = (angleDeg * Math.PI) / 180;
  return { x: cx + size * Math.cos(angleRad), y: cy + size * Math.sin(angleRad) };
}

export function boardBounds(
  tiles: Tile[],
  size: number,
): { minX: number; minY: number; maxX: number; maxY: number } {
  if (tiles.length === 0) {
    return { minX: -size, minY: -size, maxX: size, maxY: size };
  }
  let minX = Number.POSITIVE_INFINITY;
  let minY = Number.POSITIVE_INFINITY;
  let maxX = Number.NEGATIVE_INFINITY;
  let maxY = Number.NEGATIVE_INFINITY;
  for (const t of tiles) {
    const { x, y } = hexToPixel(t.q, t.r, size);
    if (x - size < minX) minX = x - size;
    if (y - size < minY) minY = y - size;
    if (x + size > maxX) maxX = x + size;
    if (y + size > maxY) maxY = y + size;
  }
  return { minX, minY, maxX, maxY };
}

export function hexNeighbors({ q, r }: HexCoord): HexCoord[] {
  return NEIGHBOR_OFFSETS.map(([dq, dr]) => ({ q: q + dq, r: r + dr }));
}

export function hexDistance(a: HexCoord, b: HexCoord): number {
  const aS = -a.q - a.r;
  const bS = -b.q - b.r;
  return (Math.abs(a.q - b.q) + Math.abs(a.r - b.r) + Math.abs(aS - bS)) / 2;
}

export function intermediatesAtDistance2(from: HexCoord, to: HexCoord): HexCoord[] {
  if (hexDistance(from, to) !== 2) return [];
  const fromNeighbors = hexNeighbors(from);
  const out: HexCoord[] = [];
  for (const n of fromNeighbors) {
    if (hexDistance(n, to) === 1) out.push(n);
  }
  return out;
}

export function reachableForAction(
  state: StateMsg,
  playerId: string,
  from: HexCoord,
): HexCoord[] {
  const tiles = state.board?.tiles ?? [];
  const tileByKey = new Map<string, Tile>();
  for (const t of tiles) tileByKey.set(`${t.q},${t.r}`, t);

  const out: HexCoord[] = [];
  for (const t of tiles) {
    if (t.q === from.q && t.r === from.r) continue;
    const d = hexDistance(from, { q: t.q, r: t.r });
    if (d === 1) {
      out.push({ q: t.q, r: t.r });
      continue;
    }
    if (d !== 2) continue;
    const mids = intermediatesAtDistance2(from, { q: t.q, r: t.r });
    for (const m of mids) {
      const mid = tileByKey.get(`${m.q},${m.r}`);
      if (mid !== undefined && mid.ownerId === playerId) {
        out.push({ q: t.q, r: t.r });
        break;
      }
    }
  }
  return out;
}
