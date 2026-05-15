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

export function reachableForAction(
  state: StateMsg,
  _playerId: string,
  from: HexCoord,
): HexCoord[] {
  const tiles = state.board?.tiles ?? [];
  const out: HexCoord[] = [];
  for (const t of tiles) {
    if (t.q === from.q && t.r === from.r) continue;
    if (hexDistance(from, { q: t.q, r: t.r }) === 1) {
      out.push({ q: t.q, r: t.r });
    }
  }
  return out;
}

function hexKey(h: HexCoord): string {
  return `${h.q},${h.r}`;
}

export function pathThroughOwnedTiles(
  state: StateMsg,
  ownerId: string,
  from: HexCoord,
  to: HexCoord,
): HexCoord[] | null {
  if (from.q === to.q && from.r === to.r) return null;
  const tiles = state.board?.tiles ?? [];
  const tileByKey = new Map<string, Tile>();
  for (const t of tiles) tileByKey.set(hexKey({ q: t.q, r: t.r }), t);
  const fromTile = tileByKey.get(hexKey(from));
  const toTile = tileByKey.get(hexKey(to));
  if (fromTile === undefined || toTile === undefined) return null;
  if (fromTile.ownerId !== ownerId || toTile.ownerId !== ownerId) return null;

  const visited = new Set<string>([hexKey(from)]);
  const prev = new Map<string, HexCoord>();
  const queue: HexCoord[] = [from];
  while (queue.length > 0) {
    const cur = queue.shift() as HexCoord;
    if (cur.q === to.q && cur.r === to.r) {
      const rev: HexCoord[] = [];
      let walker: HexCoord = cur;
      while (walker.q !== from.q || walker.r !== from.r) {
        rev.push(walker);
        const p = prev.get(hexKey(walker));
        if (p === undefined) return null;
        walker = p;
      }
      rev.reverse();
      return rev;
    }
    for (const n of hexNeighbors(cur)) {
      const k = hexKey(n);
      if (visited.has(k)) continue;
      const t = tileByKey.get(k);
      if (t === undefined || t.ownerId !== ownerId) continue;
      visited.add(k);
      prev.set(k, cur);
      queue.push(n);
    }
  }
  return null;
}
