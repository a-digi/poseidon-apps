import React, { useEffect, useState } from 'react';
import { Key, Link2 } from 'lucide-react';
import { ToastProvider } from './contexts/ToastContext';
import OAuthClientManager from './components/OAuthClientManager';
import CloudProvider from './components/CloudProvider';
import Storage from './components/Storage';
import { pluginApi } from './api/pluginApi';
import type { AccountInfo } from './api/types';

type View =
  | { kind: 'clients' }
  | { kind: 'connections' }
  | { kind: 'storage'; tokenId: string; account: AccountInfo };

type NavId = 'clients' | 'connections';

interface NavItem {
  id: NavId;
  label: string;
  description: string;
  icon: React.ReactNode;
}

const NAV_ITEMS: NavItem[] = [
  {
    id: 'clients',
    label: 'OAuth Apps',
    description: 'Manage credentials',
    icon: <Key className="w-5 h-5" />,
  },
  {
    id: 'connections',
    label: 'Connections',
    description: 'Linked accounts',
    icon: <Link2 className="w-5 h-5" />,
  },
];

const App: React.FC = () => {
  const [view, setView] = useState<View>({ kind: 'clients' });

  useEffect(() => {
    pluginApi.initTables().catch((err) => {
      console.warn('[digibox] init_tables failed:', err);
    });
  }, []);

  const activeNav: NavId = view.kind === 'storage' ? 'connections' : view.kind;

  return (
    <ToastProvider>
      <div className="flex flex-col min-h-screen bg-slate-50">
        <header className="bg-white border-b border-slate-200 px-6 py-3 flex items-center gap-3 shadow-sm shrink-0">
          <div className="w-9 h-9 rounded-lg bg-gradient-to-br from-blue-500 to-blue-700 flex items-center justify-center text-white font-bold shadow-sm">
            D
          </div>
          <div className="min-w-0">
            <div className="font-bold text-slate-900 leading-tight">Digibox</div>
            <div className="text-xs text-slate-500 leading-tight">Cloud connector</div>
          </div>
        </header>
        <div className="flex flex-1">
          <aside className="w-64 bg-white border-r border-slate-200 flex flex-col">
            <nav className="flex-1 px-3 py-4 space-y-1">
              {NAV_ITEMS.map((item) => {
                const isActive = activeNav === item.id;
                return (
                  <button
                    key={item.id}
                    type="button"
                    onClick={() => setView({ kind: item.id })}
                    className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-left transition-colors ${
                      isActive
                        ? 'bg-blue-50 text-blue-700'
                        : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'
                    }`}
                  >
                    <span className={isActive ? 'text-blue-600' : 'text-slate-400'}>
                      {item.icon}
                    </span>
                    <div className="flex-1 min-w-0">
                      <div className={`text-sm ${isActive ? 'font-semibold' : 'font-medium'}`}>
                        {item.label}
                      </div>
                      <div className="text-xs text-slate-400">{item.description}</div>
                    </div>
                  </button>
                );
              })}
            </nav>
            <footer className="px-6 py-3 text-xs text-slate-400 border-t border-slate-100">
              v1.0.0
            </footer>
          </aside>
          <main className="flex-1">
            {view.kind === 'clients' && <OAuthClientManager />}
            {view.kind === 'connections' && (
              <CloudProvider onBrowse={(tokenId, account) => setView({ kind: 'storage', tokenId, account })} />
            )}
            {view.kind === 'storage' && (
              <Storage
                tokenId={view.tokenId}
                account={view.account}
                onBack={() => setView({ kind: 'connections' })}
              />
            )}
          </main>
        </div>
      </div>
    </ToastProvider>
  );
};

export default App;
