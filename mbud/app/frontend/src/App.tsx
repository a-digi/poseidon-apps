import { useState } from 'react';
import { TopBar, type TabKey } from './common/TopBar';
import { Businesses } from './components/Businesses';
import { Users } from './components/Users';
import { Tags } from './components/Tags';
import { Invoices } from './components/Invoices';
import { Recurring } from './components/Recurring';
import { Upcoming } from './components/Upcoming';
import { SharePhoneDialog } from './components/SharePhoneDialog';
import './index.css';

export default function App() {
  const [tab, setTab] = useState<TabKey>('business');
  const [showShare, setShowShare] = useState(false);
  return (
    <div className="flex flex-col min-h-screen text-slate-900">
      <TopBar active={tab} onChange={setTab} onSharePhone={() => setShowShare(true)} />
      <main className="p-6 flex flex-col gap-6">
        {tab === 'business' && <Businesses />}
        {tab === 'users' && <Users />}
        {tab === 'tags' && <Tags />}
        {tab === 'invoice' && <Invoices />}
        {tab === 'recurring' && <Recurring />}
        {tab === 'upcoming' && <Upcoming />}
      </main>
      {showShare && <SharePhoneDialog onClose={() => setShowShare(false)} />}
    </div>
  );
}
