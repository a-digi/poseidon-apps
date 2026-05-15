import { useMemo } from 'react';
import type { ResourceType, StateMsg, Tile, Unit } from '../types';
import { boardBounds, hexCorner, hexToPixel } from './coords';

interface BoardProps {
  state: StateMsg;
  myPlayerId: string | null;
  startingTileHighlights?: Set<string>;
  reachableHighlights?: Set<string>;
  selectedSource?: { q: number; r: number } | null;
  onTileClick?: (q: number, r: number) => void;
}

const HEX_SIZE = 36;
const MARGIN = 12;

const PRODUCTION_FILL: Record<ResourceType, string> = {
  gold: '#d4a017',
  iron: '#6c757d',
  food: '#80b918',
  none: '#c2b280',
};

const NEUTRAL_STROKE = '#94a3b8';

function tileKey(q: number, r: number): string {
  return `${q},${r}`;
}

function totalPower(garrison: Unit[]): number {
  let p = 0;
  for (const u of garrison) {
    const base = u.type === 'infantry' ? 1 : u.type === 'cavalry' ? 2 : 3;
    p += base * u.level;
  }
  return p;
}

function buildHexPath(cx: number, cy: number, size: number): string {
  const parts: string[] = [];
  for (let i = 0; i < 6; i += 1) {
    const { x, y } = hexCorner(cx, cy, size, i);
    parts.push(`${i === 0 ? 'M' : 'L'}${x.toFixed(2)},${y.toFixed(2)}`);
  }
  parts.push('Z');
  return parts.join(' ');
}

interface RenderedTile {
  tile: Tile;
  cx: number;
  cy: number;
  path: string;
  fill: string;
  strokeColor: string;
  strokeWidth: number;
  isSelectedSource: boolean;
  isReachable: boolean;
  isStartingHighlight: boolean;
  hasPendingDiplomacy: boolean;
}

export function Board({
  state,
  startingTileHighlights,
  reachableHighlights,
  selectedSource,
  onTileClick,
}: BoardProps) {
  const tiles = state.board?.tiles ?? [];

  const ownerColorById = useMemo(() => {
    const m = new Map<string, string>();
    for (const p of state.players) m.set(p.id, p.color);
    return m;
  }, [state.players]);

  const pendingDiplomacyKeys = useMemo(() => {
    const s = new Set<string>();
    for (const o of state.pendingDiplomacy ?? []) s.add(tileKey(o.q, o.r));
    return s;
  }, [state.pendingDiplomacy]);

  const rendered = useMemo<RenderedTile[]>(() => {
    return tiles.map((tile) => {
      const { x, y } = hexToPixel(tile.q, tile.r, HEX_SIZE);
      const ownerColor = tile.ownerId !== '' ? ownerColorById.get(tile.ownerId) : undefined;
      const strokeColor = ownerColor ?? NEUTRAL_STROKE;
      const strokeWidth = tile.ownerId !== '' ? 3 : 2;
      const key = tileKey(tile.q, tile.r);
      return {
        tile,
        cx: x,
        cy: y,
        path: buildHexPath(x, y, HEX_SIZE),
        fill: PRODUCTION_FILL[tile.production],
        strokeColor,
        strokeWidth,
        isSelectedSource:
          selectedSource !== null &&
          selectedSource !== undefined &&
          selectedSource.q === tile.q &&
          selectedSource.r === tile.r,
        isReachable: reachableHighlights?.has(key) ?? false,
        isStartingHighlight: startingTileHighlights?.has(key) ?? false,
        hasPendingDiplomacy: pendingDiplomacyKeys.has(key),
      };
    });
  }, [
    tiles,
    ownerColorById,
    selectedSource,
    reachableHighlights,
    startingTileHighlights,
    pendingDiplomacyKeys,
  ]);

  const bounds = useMemo(() => boardBounds(tiles, HEX_SIZE), [tiles]);

  const ownerCounts = useMemo(() => {
    const counts: Record<string, number> = {};
    for (const t of tiles) {
      const k = t.ownerId === '' ? '(neutral)' : t.ownerId;
      counts[k] = (counts[k] ?? 0) + 1;
    }
    return counts;
  }, [tiles]);

  if (state.board === undefined) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-slate-200 border-t-slate-700" />
      </div>
    );
  }

  const viewBox = `${bounds.minX - MARGIN} ${bounds.minY - MARGIN} ${
    bounds.maxX - bounds.minX + MARGIN * 2
  } ${bounds.maxY - bounds.minY + MARGIN * 2}`;

  console.info('[repko/Board] render', {
    tiles: tiles.length,
    players: state.players.length,
    ownerCounts,
    viewBox,
  });

  const handleClick = (q: number, r: number) => {
    onTileClick?.(q, r);
  };

  return (
    <div className="flex h-full w-full items-center justify-center">
      <svg
        viewBox={viewBox}
        className="h-full w-full max-h-full max-w-full"
        preserveAspectRatio="xMidYMid meet"
      >
        {rendered.map((r) => {
          const yieldLabel = r.tile.production === 'none' ? '' : String(r.tile.yield);
          const garrisonCount = r.tile.garrison.length;
          const garrisonPower = totalPower(r.tile.garrison);
          const k = tileKey(r.tile.q, r.tile.r);
          return (
            <g
              key={k}
              onClick={() => handleClick(r.tile.q, r.tile.r)}
              className="cursor-pointer"
            >
              <path
                d={r.path}
                fill={r.fill}
                stroke={r.strokeColor}
                strokeWidth={r.strokeWidth}
              />
              {r.isStartingHighlight && (
                <path
                  d={r.path}
                  fill="#10b981"
                  fillOpacity={0.25}
                  stroke="#10b981"
                  strokeWidth={2}
                />
              )}
              {r.isReachable && (
                <path
                  d={r.path}
                  fill="#fde68a"
                  fillOpacity={0.35}
                  stroke="#f59e0b"
                  strokeWidth={2}
                  strokeDasharray="4 3"
                >
                  <animate
                    attributeName="fill-opacity"
                    values="0.15;0.45;0.15"
                    dur="1.6s"
                    repeatCount="indefinite"
                  />
                </path>
              )}
              {r.isSelectedSource && (
                <path
                  d={r.path}
                  fill="none"
                  stroke="#f59e0b"
                  strokeWidth={5}
                />
              )}
              {r.hasPendingDiplomacy && (
                <circle
                  cx={r.cx}
                  cy={r.cy}
                  r={HEX_SIZE * 0.9}
                  fill="none"
                  stroke="#d97706"
                  strokeWidth={2}
                  strokeDasharray="3 3"
                />
              )}
              {yieldLabel !== '' && (
                <text
                  x={r.cx}
                  y={r.cy + 4}
                  textAnchor="middle"
                  className="fill-white font-bold"
                  style={{ fontSize: 14, paintOrder: 'stroke', stroke: '#1e293b', strokeWidth: 2 }}
                >
                  {yieldLabel}
                </text>
              )}
              {garrisonCount > 0 && (
                <g transform={`translate(${r.cx - HEX_SIZE * 0.7},${r.cy - HEX_SIZE * 0.85})`}>
                  <rect
                    x={0}
                    y={0}
                    rx={6}
                    ry={6}
                    width={46}
                    height={16}
                    fill="#0f172a"
                    fillOpacity={0.85}
                  />
                  <text
                    x={23}
                    y={12}
                    textAnchor="middle"
                    className="fill-white"
                    style={{ fontSize: 10 }}
                  >
                    🛡 {garrisonCount} ({garrisonPower})
                  </text>
                </g>
              )}
            </g>
          );
        })}
      </svg>
    </div>
  );
}

export default Board;
