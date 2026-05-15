import type { ResourceType } from '../types';

interface ResourcePanelProps {
  resources: Partial<Record<ResourceType, number>>;
  armyCount: number;
}

interface VisibleResource {
  type: Exclude<ResourceType, 'none'>;
  label: string;
  color: string;
}

const VISIBLE: ReadonlyArray<VisibleResource> = [
  { type: 'gold', label: 'Gold', color: '#d4a017' },
  { type: 'iron', label: 'Iron', color: '#6c757d' },
  { type: 'food', label: 'Food', color: '#80b918' },
];

const ARMY_COLOR = '#7c3aed';

export function ResourcePanel({ resources, armyCount }: ResourcePanelProps) {
  const armyDimmed = armyCount === 0;
  const armyOpacityCls = armyDimmed ? 'opacity-50' : 'opacity-100';
  return (
    <div className="flex items-center justify-center gap-2">
      {VISIBLE.map((r) => {
        const count = resources[r.type] ?? 0;
        const dimmed = count === 0;
        const opacityCls = dimmed ? 'opacity-50' : 'opacity-100';
        return (
          <div
            key={r.type}
            className={`flex flex-col items-center rounded-md text-white w-9 h-12 ${opacityCls}`}
            style={{ background: r.color }}
          >
            <span className="text-[10px] font-medium leading-tight pt-1">{r.label}</span>
            <span className="text-lg font-bold leading-none mt-1">{count}</span>
          </div>
        );
      })}
      <div
        className={`flex flex-col items-center rounded-md text-white w-9 h-12 ${armyOpacityCls}`}
        style={{ background: ARMY_COLOR }}
      >
        <span className="text-[10px] font-medium leading-tight pt-1">⚔</span>
        <span className="text-lg font-bold leading-none mt-1">{armyCount}</span>
      </div>
    </div>
  );
}

export default ResourcePanel;
