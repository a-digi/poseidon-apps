import type { ResourceType, UnitType } from '../types';

interface ArmyBreakdown {
  infantry: number;
  armor: number;
  jet: number;
}

interface ResourcePanelProps {
  resources: Partial<Record<ResourceType, number>>;
  armyBreakdown: ArmyBreakdown;
}

const RESOURCE_CHIPS: ReadonlyArray<{ type: Exclude<ResourceType, 'none'>; icon: string }> = [
  { type: 'credits', icon: '💵' },
  { type: 'steel', icon: '⚒' },
  { type: 'fuel', icon: '⛽' },
];

const ARMY_CHIPS: ReadonlyArray<{ type: UnitType; icon: string }> = [
  { type: 'infantry', icon: '🪖' },
  { type: 'armor', icon: '🚛' },
  { type: 'jet', icon: '✈️' },
];

function Chip({ icon, count }: { icon: string; count: number }) {
  return (
    <span className={`inline-flex items-center gap-0.5 rounded px-1.5 py-1 text-xs font-bold ${count === 0 ? 'opacity-40' : ''} bg-slate-100 text-slate-800`}>
      <span>{icon}</span>
      <span>{count}</span>
    </span>
  );
}

export function ResourcePanel({ resources, armyBreakdown }: ResourcePanelProps) {
  return (
    <div className="flex items-center gap-1.5 flex-wrap">
      {RESOURCE_CHIPS.map((r) => (
        <Chip key={r.type} icon={r.icon} count={resources[r.type] ?? 0} />
      ))}
      <span className="text-slate-300 text-xs mx-0.5">│</span>
      {ARMY_CHIPS.map((a) => (
        <Chip key={a.type} icon={a.icon} count={armyBreakdown[a.type]} />
      ))}
    </div>
  );
}

export default ResourcePanel;
