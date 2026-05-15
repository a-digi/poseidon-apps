import type { Army, StateMsg } from '../types';
import { UNIT_ICON, UNIT_NAME } from './units';

interface ArmyInspectorPopoverProps {
  army: Army | null;
  state: StateMsg;
  onClose: () => void;
}

export function ArmyInspectorPopover({ army, state, onClose }: ArmyInspectorPopoverProps) {
  if (army === null) return null;
  const owner = state.players.find((p) => p.id === army.ownerId);
  const totalUnits = army.units.reduce((sum, u) => sum + u.count, 0);
  const eta = Math.ceil(army.pathRemaining.length / 3);
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      onClick={onClose}
    >
      <div
        className="max-h-[80vh] w-full max-w-sm overflow-y-auto rounded-lg bg-white p-4 shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="mb-3 flex items-center justify-between border-b border-slate-200 pb-2">
          <h2 className="text-base font-bold text-slate-900">
            🚶 {owner?.name ?? 'Army'} — {totalUnits} units
          </h2>
          <button
            type="button"
            aria-label="Close"
            onClick={onClose}
            className="rounded px-2 py-0.5 text-slate-500 hover:bg-slate-100 hover:text-slate-900"
          >
            ✕
          </button>
        </header>
        <p className="mb-2 text-xs text-slate-600">
          Marching to ({army.destQ}, {army.destR}) — ETA {eta} round{eta === 1 ? '' : 's'}
        </p>
        <ul className="space-y-1 text-xs">
          {army.units.map((u) => (
            <li key={`${u.type}-${u.level}`} className="flex items-center justify-between">
              <span>
                {UNIT_ICON[u.type]} {UNIT_NAME[u.type]}{' '}
                <span className="text-slate-500">L{u.level} ×{u.count}</span>
              </span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}

export default ArmyInspectorPopover;
