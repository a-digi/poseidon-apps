import { Briefcase, CalendarClock, FileText, Repeat, Users as UsersRound } from 'lucide-react';

type TabKey = 'business' | 'users' | 'invoice' | 'recurring' | 'upcoming';

interface TabBarProps {
  active: TabKey;
  onChange: (k: TabKey) => void;
}

interface TabDef {
  key: TabKey;
  label: string;
  Icon: typeof Briefcase;
}

const TABS: TabDef[] = [
  { key: 'business', label: 'Businesses', Icon: Briefcase },
  { key: 'users', label: 'Users', Icon: UsersRound },
  { key: 'invoice', label: 'Invoices', Icon: FileText },
  { key: 'recurring', label: 'Recurring', Icon: Repeat },
  { key: 'upcoming', label: 'Upcoming', Icon: CalendarClock },
];

export function TabBar({ active, onChange }: TabBarProps) {
  return (
    <div className="flex items-center gap-1">
      {TABS.map(({ key, label, Icon }) => {
        const isActive = key === active;
        const cls = isActive
          ? 'bg-slate-900 text-white'
          : 'text-slate-600 hover:bg-slate-100';
        return (
          <button
            key={key}
            type="button"
            onClick={() => onChange(key)}
            className={`flex items-center gap-2 px-3 py-2 rounded-md text-sm font-medium transition-colors ${cls}`}
          >
            <Icon className="w-4 h-4" />
            {label}
          </button>
        );
      })}
    </div>
  );
}
