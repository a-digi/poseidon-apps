import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import './index.css';

const _sp = new URLSearchParams(window.location.search);
if (_sp.get('mode') === 'mobile') {
  import('./mobile/MobileUploaderPage').then(({ MobileUploaderPage }) => {
    createRoot(document.getElementById('root')!).render(
      <StrictMode><MobileUploaderPage /></StrictMode>
    );
  });
} else {
  import('./App').then(({ default: App }) => {
    createRoot(document.getElementById('root')!).render(
      <StrictMode><App /></StrictMode>
    );
  });
}
