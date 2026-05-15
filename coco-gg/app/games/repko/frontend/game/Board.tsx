import { useCallback, useMemo, useRef, useState } from 'react';
import type { PointerEvent as ReactPointerEvent, WheelEvent as ReactWheelEvent } from 'react';
import type { GarrisonStack, ResourceType, StateMsg, Tile, UnitType } from '../types';
import { boardBounds, hexCorner, hexToPixel } from './coords';

interface BoardProps {
  state: StateMsg;
  myPlayerId: string | null;
  startingTileHighlights?: Set<string>;
  reachableHighlights?: Set<string>;
  selectedSource?: { q: number; r: number } | null;
  onTileClick?: (q: number, r: number) => void;
  attackMode?: boolean;
}

const HEX_SIZE = 28;
const MARGIN = 20;
const MIN_ZOOM = 0.5;
const MAX_ZOOM = 4.0;
const TIER1 = 1.2;
const TIER2 = 2.0;
const TIER3 = 3.0;
const DRAG_THRESHOLD_PX = 8;

const PRODUCTION_GRADIENT: Record<ResourceType, string> = {
  gold: 'url(#grad-gold)',
  iron: 'url(#grad-iron)',
  food: 'url(#grad-food)',
  none: 'url(#grad-none)',
};

const PRODUCTION_ICON: Record<ResourceType, string> = {
  gold: '💰',
  iron: '⚒',
  food: '🍞',
  none: '',
};

const BASE_POWER: Record<UnitType, number> = {
  infantry: 1,
  cavalry: 2,
  artillery: 3,
};

const UNIT_ICON: Record<UnitType, string> = {
  infantry: '🏹',
  cavalry: '🐎',
  artillery: '💥',
};

const NEUTRAL_STROKE = '#94a3b8';

function tileKey(q: number, r: number): string {
  return `${q},${r}`;
}

function clamp(value: number, min: number, max: number): number {
  if (value < min) return min;
  if (value > max) return max;
  return value;
}

function totalUnits(garrison: GarrisonStack[]): number {
  let n = 0;
  for (const s of garrison) n += s.count;
  return n;
}

function totalPowerOf(garrison: GarrisonStack[]): number {
  let p = 0;
  for (const s of garrison) p += s.count * BASE_POWER[s.type] * s.level;
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
  strokeDasharray?: string;
  civBackgroundColor: string;
  foundedByOwner: boolean;
  isCapital: boolean;
  ownerFlag: string;
  isOwnedByMe: boolean;
  isOwnedByOther: boolean;
  isSelectedSource: boolean;
  isReachable: boolean;
  isStartingHighlight: boolean;
  hasPendingDiplomacy: boolean;
  units: number;
  power: number;
}

export function Board({
  state,
  myPlayerId,
  startingTileHighlights,
  reachableHighlights,
  selectedSource,
  onTileClick,
  attackMode,
}: BoardProps) {
  const tiles = state.board?.tiles ?? [];

  const [zoom, setZoom] = useState(1);
  const [pan, setPan] = useState({ x: 0, y: 0 });
  const pointersRef = useRef<Map<number, { x: number; y: number }>>(new Map());
  const initialPinchRef = useRef<{ distance: number; zoom: number } | null>(null);
  const dragMovedRef = useRef(false);
  const svgRef = useRef<SVGSVGElement | null>(null);

  const civById = useMemo(() => {
    const m = new Map<string, { color: string; backgroundColor: string; flag: string; coatOfArms: string }>();
    const civs = state.civilizations ?? [];
    for (const c of civs) {
      m.set(c.id, {
        color: c.color,
        backgroundColor: c.backgroundColor,
        flag: c.flag,
        coatOfArms: c.coatOfArms,
      });
    }
    return m;
  }, [state.civilizations]);

  const playerColorAndFlag = useMemo(() => {
    const m = new Map<string, { color: string; backgroundColor: string; flag: string; coatOfArms: string }>();
    for (const p of state.players) {
      const civ = civById.get(p.civilizationId);
      m.set(p.id, {
        color: civ?.color ?? p.color,
        backgroundColor: civ?.backgroundColor ?? '',
        flag: civ?.flag ?? '',
        coatOfArms: civ?.coatOfArms ?? '',
      });
    }
    return m;
  }, [state.players, civById]);

  const pendingDiplomacyKeys = useMemo(() => {
    const s = new Set<string>();
    for (const o of state.pendingDiplomacy ?? []) s.add(tileKey(o.q, o.r));
    return s;
  }, [state.pendingDiplomacy]);

  const rendered = useMemo<RenderedTile[]>(() => {
    return tiles.map((tile) => {
      const { x, y } = hexToPixel(tile.q, tile.r, HEX_SIZE);
      const ownerInfo = tile.ownerId !== '' ? playerColorAndFlag.get(tile.ownerId) : undefined;
      const strokeColor = ownerInfo?.color ?? NEUTRAL_STROKE;
      const foundedByOwner =
        tile.foundedBy !== undefined && tile.foundedBy !== '' && tile.foundedBy === tile.ownerId;
      const isCapital = tile.foundedBy !== undefined && tile.foundedBy !== '';
      let strokeWidth: number;
      let strokeDasharray: string | undefined;
      if (tile.ownerId === '') {
        strokeWidth = 1.5;
      } else if (foundedByOwner) {
        strokeWidth = 2.5;
      } else {
        strokeWidth = 3.5;
        strokeDasharray = '6 3';
      }
      const key = tileKey(tile.q, tile.r);
      return {
        tile,
        cx: x,
        cy: y,
        path: buildHexPath(x, y, HEX_SIZE),
        fill: PRODUCTION_GRADIENT[tile.production],
        strokeColor,
        strokeWidth,
        strokeDasharray,
        civBackgroundColor: ownerInfo?.backgroundColor ?? '',
        foundedByOwner,
        isCapital,
        ownerFlag: ownerInfo?.flag ?? '',
        isOwnedByMe: myPlayerId !== null && tile.ownerId === myPlayerId,
        isOwnedByOther: tile.ownerId !== '' && tile.ownerId !== myPlayerId,
        isSelectedSource:
          selectedSource !== null &&
          selectedSource !== undefined &&
          selectedSource.q === tile.q &&
          selectedSource.r === tile.r,
        isReachable: reachableHighlights?.has(key) ?? false,
        isStartingHighlight: startingTileHighlights?.has(key) ?? false,
        hasPendingDiplomacy: pendingDiplomacyKeys.has(key),
        units: totalUnits(tile.garrison),
        power: totalPowerOf(tile.garrison),
      };
    });
  }, [
    tiles,
    playerColorAndFlag,
    selectedSource,
    reachableHighlights,
    startingTileHighlights,
    pendingDiplomacyKeys,
    myPlayerId,
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

  const bvw = bounds.maxX - bounds.minX + MARGIN * 2;
  const bvh = bounds.maxY - bounds.minY + MARGIN * 2;
  const baseCenterX = bounds.minX - MARGIN + bvw / 2;
  const baseCenterY = bounds.minY - MARGIN + bvh / 2;
  const vw = bvw / zoom;
  const vh = bvh / zoom;
  const clampedPanX = clamp(pan.x, -bvw / 2, bvw / 2);
  const clampedPanY = clamp(pan.y, -bvh / 2, bvh / 2);
  const visMinX = baseCenterX + clampedPanX - vw / 2;
  const visMinY = baseCenterY + clampedPanY - vh / 2;
  const viewBox = `${visMinX} ${visMinY} ${vw} ${vh}`;

  const handleZoomIn = useCallback(() => {
    setZoom((z) => clamp(z * 1.25, MIN_ZOOM, MAX_ZOOM));
  }, []);
  const handleZoomOut = useCallback(() => {
    setZoom((z) => clamp(z / 1.25, MIN_ZOOM, MAX_ZOOM));
  }, []);
  const handleZoomReset = useCallback(() => {
    setZoom(1);
    setPan({ x: 0, y: 0 });
  }, []);

  const handleWheel = useCallback((e: ReactWheelEvent<SVGSVGElement>) => {
    e.preventDefault();
    setZoom((z) => clamp(z * Math.exp(-e.deltaY * 0.001), MIN_ZOOM, MAX_ZOOM));
  }, []);

  const handlePointerDown = useCallback((e: ReactPointerEvent<SVGSVGElement>) => {
    const target = e.target as SVGElement;
    if (typeof target.setPointerCapture === 'function') {
      try {
        target.setPointerCapture(e.pointerId);
      } catch {
        /* capture may fail on some elements; safe to ignore */
      }
    }
    pointersRef.current.set(e.pointerId, { x: e.clientX, y: e.clientY });
    dragMovedRef.current = false;
    if (pointersRef.current.size === 2) {
      const pts = Array.from(pointersRef.current.values());
      const dx = pts[0].x - pts[1].x;
      const dy = pts[0].y - pts[1].y;
      initialPinchRef.current = { distance: Math.hypot(dx, dy), zoom };
    }
  }, [zoom]);

  const handlePointerMove = useCallback((e: ReactPointerEvent<SVGSVGElement>) => {
    const prev = pointersRef.current.get(e.pointerId);
    if (prev === undefined) return;
    const next = { x: e.clientX, y: e.clientY };
    pointersRef.current.set(e.pointerId, next);
    const dxCss = next.x - prev.x;
    const dyCss = next.y - prev.y;
    if (Math.hypot(dxCss, dyCss) > 0) {
      dragMovedRef.current = dragMovedRef.current || Math.hypot(dxCss, dyCss) > DRAG_THRESHOLD_PX;
    }

    if (pointersRef.current.size >= 2 && initialPinchRef.current !== null) {
      const pts = Array.from(pointersRef.current.values());
      const px = pts[0].x - pts[1].x;
      const py = pts[0].y - pts[1].y;
      const newDistance = Math.hypot(px, py);
      const ratio = newDistance / initialPinchRef.current.distance;
      setZoom(clamp(initialPinchRef.current.zoom * ratio, MIN_ZOOM, MAX_ZOOM));
      dragMovedRef.current = true;
      return;
    }

    if (pointersRef.current.size === 1) {
      const rect = svgRef.current?.getBoundingClientRect();
      if (rect === undefined || rect.width === 0 || rect.height === 0) return;
      const panDx = (dxCss * vw) / rect.width;
      const panDy = (dyCss * vh) / rect.height;
      setPan((p) => ({
        x: clamp(p.x - panDx, -bvw / 2, bvw / 2),
        y: clamp(p.y - panDy, -bvh / 2, bvh / 2),
      }));
    }
  }, [vw, vh, bvw, bvh]);

  const handlePointerUp = useCallback((e: ReactPointerEvent<SVGSVGElement>) => {
    pointersRef.current.delete(e.pointerId);
    if (pointersRef.current.size < 2) {
      initialPinchRef.current = null;
    }
  }, []);

  const handleDoubleClick = useCallback(() => {
    setZoom(1);
    setPan({ x: 0, y: 0 });
  }, []);

  const handleTileClick = useCallback((q: number, r: number) => {
    if (dragMovedRef.current) {
      dragMovedRef.current = false;
      return;
    }
    onTileClick?.(q, r);
  }, [onTileClick]);

  if (state.board === undefined) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-slate-200 border-t-slate-700" />
      </div>
    );
  }

  console.info('[repko/Board] render', {
    tiles: tiles.length,
    players: state.players.length,
    ownerCounts,
    viewBox,
    zoom,
    pan: { x: clampedPanX, y: clampedPanY },
  });

  const showTier1 = zoom >= TIER1;
  const showTier2 = zoom >= TIER2;
  const showTier3 = zoom >= TIER3;
  const yieldFont = Math.max(8, Math.round(10 * Math.min(zoom, 2)));
  const iconFont = Math.max(7, Math.round(9 * Math.min(zoom, 2)));
  const badgeFont = 9;
  const stackFont = 9;

  return (
    <div className="relative flex h-full w-full items-center justify-center">
      <svg
        ref={svgRef}
        viewBox={viewBox}
        className="h-full w-full max-h-full max-w-full touch-none select-none"
        preserveAspectRatio="xMidYMid meet"
        onWheel={handleWheel}
        onPointerDown={handlePointerDown}
        onPointerMove={handlePointerMove}
        onPointerUp={handlePointerUp}
        onPointerCancel={handlePointerUp}
        onDoubleClick={handleDoubleClick}
      >
        <defs>
          <radialGradient id="grad-gold" cx="38%" cy="28%" r="75%">
            <stop offset="0%" stopColor="#fde68a" />
            <stop offset="52%" stopColor="#d4a017" />
            <stop offset="100%" stopColor="#78340f" />
          </radialGradient>
          <radialGradient id="grad-iron" cx="38%" cy="28%" r="75%">
            <stop offset="0%" stopColor="#e2e8f0" />
            <stop offset="52%" stopColor="#64748b" />
            <stop offset="100%" stopColor="#1e293b" />
          </radialGradient>
          <radialGradient id="grad-food" cx="38%" cy="28%" r="75%">
            <stop offset="0%" stopColor="#d9f99d" />
            <stop offset="52%" stopColor="#65a30d" />
            <stop offset="100%" stopColor="#1a2e05" />
          </radialGradient>
          <radialGradient id="grad-none" cx="38%" cy="28%" r="75%">
            <stop offset="0%" stopColor="#f5eedb" />
            <stop offset="52%" stopColor="#c2b280" />
            <stop offset="100%" stopColor="#5c4a1e" />
          </radialGradient>
          <linearGradient id="tile-sheen" x1="0" y1="0" x2="0.3" y2="1">
            <stop offset="0%" stopColor="#ffffff" stopOpacity="0.28" />
            <stop offset="45%" stopColor="#ffffff" stopOpacity="0.06" />
            <stop offset="100%" stopColor="#000000" stopOpacity="0.18" />
          </linearGradient>
        </defs>
        {rendered.map((r) => {
          const yieldLabel = r.tile.production === 'none' ? '' : String(r.tile.yield);
          const k = tileKey(r.tile.q, r.tile.r);
          const showAttackRing =
            attackMode === true && r.isReachable && !r.isOwnedByMe;
          const productionIcon = PRODUCTION_ICON[r.tile.production];

          const stackRow: { icon: string; level: number; count: number }[] = [];
          if (showTier3) {
            for (const stack of r.tile.garrison) {
              if (stack.count <= 0) continue;
              stackRow.push({
                icon: UNIT_ICON[stack.type],
                level: stack.level,
                count: stack.count,
              });
            }
          }

          return (
            <g
              key={k}
              onClick={() => handleTileClick(r.tile.q, r.tile.r)}
              className="cursor-pointer"
            >
              <path
                d={r.path}
                fill={r.fill}
                stroke={r.strokeColor}
                strokeWidth={r.strokeWidth}
                strokeDasharray={r.strokeDasharray}
              />
              {r.civBackgroundColor !== '' && (
                <path d={r.path} fill={r.civBackgroundColor} fillOpacity={0.28} stroke="none" />
              )}
              <path d={r.path} fill="url(#tile-sheen)" stroke="none" />
              {r.isStartingHighlight && (
                <path
                  d={r.path}
                  fill="#10b981"
                  fillOpacity={0.25}
                  stroke="#10b981"
                  strokeWidth={2}
                />
              )}
              {r.isReachable && !showAttackRing && (
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
              {showAttackRing && (
                <g className="animate-pulse">
                  <path
                    d={r.path}
                    fill="none"
                    stroke="#dc2626"
                    strokeWidth={3}
                    strokeDasharray="4 2"
                  />
                </g>
              )}
              {showAttackRing && showTier1 && r.units > 0 && (
                <text
                  x={r.cx + HEX_SIZE * 0.7}
                  y={r.cy + HEX_SIZE * 0.95}
                  textAnchor="end"
                  fill="#dc2626"
                  style={{ fontSize: badgeFont, fontWeight: 600, paintOrder: 'stroke', stroke: '#ffffff', strokeWidth: 2 }}
                >
                  def {r.units}({r.power})
                </text>
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
              {r.isCapital && r.tile.ownerId !== '' && (
                <g
                  transform={`translate(${r.cx},${r.cy - HEX_SIZE * 0.35}) scale(0.9)`}
                >
                  <path
                    d="M -8 4 L -8 -2 L -6 -2 L -6 -4 L -4 -4 L -4 -2 L -2 -2 L -2 -4 L 2 -4 L 2 -2 L 4 -2 L 4 -4 L 6 -4 L 6 -2 L 8 -2 L 8 4 Z"
                    fill={r.strokeColor}
                    stroke="#ffffff"
                    strokeWidth={0.8}
                  />
                </g>
              )}
              {yieldLabel !== '' && (
                <text
                  x={r.cx}
                  y={r.cy + yieldFont / 3}
                  textAnchor="middle"
                  fill="#ffffff"
                  style={{ fontSize: yieldFont, fontWeight: 700, paintOrder: 'stroke', stroke: '#1e293b', strokeWidth: 2 }}
                >
                  {yieldLabel}
                </text>
              )}
              {showTier1 && r.units > 0 && (
                <g transform={`translate(${r.cx - HEX_SIZE * 0.55},${r.cy - HEX_SIZE * 0.95})`}>
                  <rect
                    x={0}
                    y={0}
                    rx={5}
                    ry={5}
                    width={showTier2 ? 50 : 32}
                    height={13}
                    fill="#0f172a"
                    fillOpacity={0.85}
                  />
                  <text
                    x={(showTier2 ? 50 : 32) / 2}
                    y={9.5}
                    textAnchor="middle"
                    fill="#ffffff"
                    style={{ fontSize: badgeFont }}
                  >
                    {showTier2 ? `🛡 ${r.units} (${r.power})` : `🛡 ${r.units}`}
                  </text>
                </g>
              )}
              {productionIcon !== '' && (
                <text
                  x={r.cx - HEX_SIZE * 0.55}
                  y={r.cy - HEX_SIZE * 0.2}
                  textAnchor="middle"
                  style={{ fontSize: iconFont }}
                >
                  {productionIcon}
                </text>
              )}
              {showTier2 && r.ownerFlag !== '' && (
                <text
                  x={r.cx + HEX_SIZE * 0.55}
                  y={r.cy + HEX_SIZE * 0.7}
                  textAnchor="middle"
                  style={{ fontSize: 10 }}
                >
                  {r.ownerFlag}
                </text>
              )}
              {showTier3 && stackRow.length > 0 && (
                <text
                  x={r.cx}
                  y={r.cy + HEX_SIZE * 0.55}
                  textAnchor="middle"
                  fill="#0f172a"
                  style={{ fontSize: stackFont, paintOrder: 'stroke', stroke: '#ffffff', strokeWidth: 2 }}
                >
                  {stackRow.map((s) => `${s.icon}${s.level === 1 ? '' : s.level === 2 ? '²' : '³'}×${s.count}`).join(' ')}
                </text>
              )}
            </g>
          );
        })}
      </svg>
      <div className="absolute bottom-3 right-3 flex flex-col gap-1 bg-white/90 rounded-md shadow p-1">
        <button
          type="button"
          onClick={handleZoomIn}
          className="h-10 w-10 flex items-center justify-center text-lg font-semibold text-slate-700 hover:bg-slate-100 rounded"
          aria-label="Zoom in"
        >
          +
        </button>
        <button
          type="button"
          onClick={handleZoomOut}
          className="h-10 w-10 flex items-center justify-center text-lg font-semibold text-slate-700 hover:bg-slate-100 rounded"
          aria-label="Zoom out"
        >
          −
        </button>
        <button
          type="button"
          onClick={handleZoomReset}
          className="h-10 w-10 flex items-center justify-center text-lg font-semibold text-slate-700 hover:bg-slate-100 rounded"
          aria-label="Reset zoom"
        >
          ⊙
        </button>
      </div>
    </div>
  );
}

export default Board;
