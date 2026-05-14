import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import './index.css';

const sp = new URLSearchParams(window.location.search);
const token = sp.get('t');
if (token) {
  try {
    localStorage.setItem('coco_gg_mobile_token', token);
  } catch {
    /* storage unavailable — ignore */
  }
  sp.delete('t');
  const query = sp.toString();
  const cleaned = window.location.pathname + (query ? '?' + query : '') + window.location.hash;
  history.replaceState(null, '', cleaned);
}

const rootEl = document.getElementById('root');
if (!rootEl) {
  throw new Error('#root element missing from index.html');
}

if (sp.get('mode') === 'mobile') {
  import('./mobile/MobilePage').then(({ MobilePage }) => {
    createRoot(rootEl).render(
      <StrictMode>
        <MobilePage />
      </StrictMode>,
    );
  });
} else {
  import('./App').then(({ default: App }) => {
    createRoot(rootEl).render(
      <StrictMode>
        <App />
      </StrictMode>,
    );
  });
}
