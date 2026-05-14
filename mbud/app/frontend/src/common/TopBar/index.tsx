import { useState } from 'react';
import { Briefcase, CalendarClock, FileText, Menu, Repeat, Smartphone, Tag, Users, X } from 'lucide-react';
import { useIsMobile } from '../../lib/useIsMobile';

export type TabKey = 'business' | 'users' | 'tags' | 'invoice' | 'recurring' | 'upcoming';

interface TopBarProps {
  active: TabKey;
  onChange: (k: TabKey) => void;
  onSharePhone: () => void;
}

interface TabDef {
  key: TabKey;
  label: string;
  Icon: typeof Briefcase;
}

const TABS: TabDef[] = [
  { key: 'business',  label: 'Businesses', Icon: Briefcase },
  { key: 'users',     label: 'Users',       Icon: Users },
  { key: 'tags',      label: 'Tags',        Icon: Tag },
  { key: 'invoice',   label: 'Invoices',    Icon: FileText },
  { key: 'recurring', label: 'Recurring',   Icon: Repeat },
  { key: 'upcoming',  label: 'Upcoming',    Icon: CalendarClock },
];

export function TopBar({ active, onChange, onSharePhone }: TopBarProps) {
  const isMobile = useIsMobile();
  const [menuOpen, setMenuOpen] = useState(false);

  const activeTab = TABS.find(t => t.key === active);

  if (isMobile) {
    return (
      <>
        <header className="sticky top-0 z-30 bg-gray-50 border-b border-slate-200 px-4 flex items-center h-14">
          <span className="text-base font-bold text-slate-900">Budgeting</span>
          {activeTab && (
            <span className="ml-3 text-sm text-slate-500 font-medium">{activeTab.label}</span>
          )}
          <button
            type="button"
            onClick={() => setMenuOpen(o => !o)}
            aria-label="Open menu"
            aria-expanded={menuOpen}
            className="ml-auto p-2 rounded-md text-slate-600 hover:bg-slate-200 transition-colors"
          >
            {menuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
          </button>
        </header>

        {menuOpen && (
          <div
            className="fixed inset-0 z-20 bg-black/40"
            onClick={() => setMenuOpen(false)}
          >
            <nav
              className="absolute right-0 top-14 w-56 bg-white border-l border-b border-slate-200 rounded-bl-xl shadow-xl overflow-hidden"
              aria-label="Main navigation"
              onClick={e => e.stopPropagation()}
            >
              {TABS.map(({ key, label, Icon }) => {
                const isActive = key === active;
                return (
                  <button
                    key={key}
                    type="button"
                    onClick={() => { onChange(key); setMenuOpen(false); }}
                    aria-current={isActive ? 'page' : undefined}
                    className={`w-full flex items-center gap-3 px-4 py-3 text-sm font-medium transition-colors ${
                      isActive
                        ? 'bg-slate-900 text-white'
                        : 'text-slate-700 hover:bg-slate-100'
                    }`}
                  >
                    <Icon className="w-4 h-4 shrink-0" />
                    {label}
                  </button>
                );
              })}
            </nav>
          </div>
        )}
      </>
    );
  }

  return (
    <header className="sticky top-0 z-10 bg-gray-50 border-b border-slate-200 px-6 flex items-center gap-6 h-14">
      <span className="text-base font-bold text-slate-900 shrink-0">Budgeting</span>
      <nav className="flex items-center gap-1" aria-label="Main navigation">
        {TABS.map(({ key, label, Icon }) => {
          const isActive = key === active;
          return (
            <button
              key={key}
              type="button"
              onClick={() => onChange(key)}
              aria-current={isActive ? 'page' : undefined}
              className={`flex items-center gap-2 px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                isActive
                  ? 'bg-slate-900 text-white'
                  : 'text-slate-600 hover:bg-slate-200'
              }`}
            >
              <Icon className="w-4 h-4" />
              {label}
            </button>
          );
        })}
      </nav>
      <div className="ml-auto">
        <button type="button" onClick={onSharePhone} aria-label="Open on phone" className="p-2 rounded-md text-slate-600 hover:bg-slate-200 transition-colors">
          <Smartphone className="w-5 h-5" />
        </button>
      </div>
    </header>
  );
}
