import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import './index.css';

const sp = new URLSearchParams(window.location.search);
const token = sp.get('t');
if (token !== null) {
  try {
    localStorage.setItem('coco_gg_mobile_token', token);
  } catch {
    /* storage unavailable — ignore */
  }
  sp.delete('t');
  const newSearch = sp.toString();
  const newUrl =
    window.location.pathname + (newSearch ? '?' + newSearch : '') + window.location.hash;
  history.replaceState(null, '', newUrl);
}

const rootEl = document.getElementById('root');
if (rootEl === null) {
  throw new Error('#root element missing from index.html');
}

if (sp.get('mode') === 'mobile') {
  import('./shell/games-registry').then(({ GAMES }) => {
    const gameId = sp.get('game') ?? GAMES[0]?.id ?? '';
    const descriptor = GAMES.find((g) => g.id === gameId);
    if (descriptor === undefined) {
      createRoot(rootEl).render(
        <StrictMode>
          <div className="fixed inset-0 flex items-center justify-center p-1text-center text-slate-700">
            <p>Unknown game. Open a fresh QR code from the desktop.</p>
          </div>
        </StrictMode>,
      );
      return;
    }
    descriptor.loadMobile().then(({ default: MobilePage }) => {
      createRoot(rootEl).render(
        <StrictMode>
          <MobilePage />
        </StrictMode>,
      );
    });
  });
} else {
  import('./shell/ServerShell').then(({ default: ServerShell }) => {
    createRoot(rootEl).render(
      <StrictMode>
        <ServerShell />
      </StrictMode>,
    );
  });
}
