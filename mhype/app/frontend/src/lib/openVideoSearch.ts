import { callPlugin } from '../api';

export function openVideoSearch(artist: string, title: string): void {
  const params = new URLSearchParams({ q: `${artist} ${title}`.trim(), sclient: 'gws-wiz-modeless-video' });
  const url = `https://www.google.com/search?${params}`;
  void callPlugin<{ ok: true }>('open_browser', { url });
}
