function detectApiBase(): string {
  // Three layers, in priority order. See plan/coco-gg-local-remote-modes.md
  // for the verification against PluginPersistentHost.tsx.
  if (typeof window === 'undefined') return 'http://localhost:2014';
  try {
    const sp = new URLSearchParams(window.location.search);
    const fromQuery = sp.get('backendUrl');
    if (fromQuery) return fromQuery;
  } catch { /* fall through */ }
  if (window.location.origin && window.location.origin.startsWith('http')) {
    return window.location.origin;
  }
  return 'http://localhost:2014';
}

export const HOST = detectApiBase();
const PLUGIN_ID = 'coco-gg';
const ADMIN_TOKEN_KEY = 'coco_gg_admin_token';

export interface MobileSession {
  token: string;
  expiresAt: number;
}

export interface LongRunningStatus {
  running: boolean;
  pid?: number;
  startedAt?: number;
  exitCode?: number;
}

export interface GameInfo {
  id: string;
  name: string;
  description: string;
}

export interface ServerConfig {
  mode: 'local' | 'remote';
  baseUrl: string | null;
}

export function getAdminToken(): string | null {
  try { return localStorage.getItem(ADMIN_TOKEN_KEY); } catch { return null; }
}

export function setAdminToken(token: string): void {
  try { localStorage.setItem(ADMIN_TOKEN_KEY, token); } catch { /* ignore */ }
}

export function clearAdminToken(): void {
  try { localStorage.removeItem(ADMIN_TOKEN_KEY); } catch { /* ignore */ }
}

// In local mode (no admin token in localStorage), returns an empty object —
// the Wails host's LoopbackOnly middleware does the gating instead. In
// remote mode the operator has signed in and saved a token; we attach it as
// Authorization: Bearer.
export function authHeaders(): Record<string, string> {
  const t = getAdminToken();
  return t === null ? {} : { Authorization: `Bearer ${t}` };
}

export function getServerConfig(): ServerConfig {
  // The deployment mode is encoded in the URL the bundle was loaded with.
  // Wails always injects ?pluginId=…&backendUrl=… on plugin iframe URLs
  // (frontend/src/components/Plugin/PluginPersistentHost.tsx:25). If either
  // is present, we're in a Wails iframe — local mode. Otherwise the bundle
  // was loaded directly from a public host — remote mode, where this origin
  // is the public base URL.
  //
  // Doing this synchronously and client-side avoids a chicken-and-egg with
  // the plugin's lifecycle: the dashboard is what starts the plugin, so
  // it can't depend on the plugin being up at first render.
  if (typeof window === 'undefined') return { mode: 'local', baseUrl: null };
  try {
    const sp = new URLSearchParams(window.location.search);
    if (sp.has('pluginId') || sp.has('backendUrl')) {
      return { mode: 'local', baseUrl: null };
    }
  } catch {
    /* fall through to origin check */
  }
  if (window.location.origin && window.location.origin.startsWith('http')) {
    return { mode: 'remote', baseUrl: window.location.origin };
  }
  return { mode: 'local', baseUrl: null };
}

export async function startPlugin(): Promise<LongRunningStatus> {
  const r = await fetch(`${HOST}/api/plugins/long-running/start`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify({ pluginId: PLUGIN_ID }),
  });
  if (!r.ok && r.status !== 409) {
    throw new Error(`startPlugin: ${r.statusText}`);
  }
  const env = (await r.json()) as { status: string; message: LongRunningStatus | string };
  if (env.status !== 'success') {
    throw new Error(typeof env.message === 'string' ? env.message : 'startPlugin failed');
  }
  return env.message as LongRunningStatus;
}

export async function stopPlugin(): Promise<void> {
  const r = await fetch(`${HOST}/api/plugins/long-running/stop`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify({ pluginId: PLUGIN_ID }),
  });
  if (!r.ok) throw new Error(`stopPlugin: ${r.statusText}`);
}

export async function getStatus(): Promise<LongRunningStatus> {
  const r = await fetch(
    `${HOST}/api/plugins/long-running/status?pluginId=${encodeURIComponent(PLUGIN_ID)}`,
    { headers: { ...authHeaders() } },
  );
  if (!r.ok) throw new Error(`getStatus: ${r.statusText}`);
  const env = (await r.json()) as { status: string; message: LongRunningStatus | string };
  if (env.status !== 'success') {
    throw new Error(typeof env.message === 'string' ? env.message : 'getStatus failed');
  }
  return env.message as LongRunningStatus;
}

export async function createMobileSession(ttlSeconds = 3600): Promise<MobileSession> {
  const r = await fetch(`${HOST}/api/mobile-sessions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify({ pluginId: PLUGIN_ID, ttlSeconds }),
  });
  if (!r.ok) throw new Error(`mobile-session: ${r.statusText}`);
  return r.json() as Promise<MobileSession>;
}

export async function getLanAddresses(): Promise<string[]> {
  const r = await fetch(`${HOST}/api/system/lan-addresses`, { headers: { ...authHeaders() } });
  if (!r.ok) return [];
  const j = (await r.json()) as { addresses?: string[] };
  return j.addresses ?? [];
}

export async function listGames(): Promise<GameInfo[]> {
  const r = await fetch(`${HOST}/plugins/${PLUGIN_ID}/api/games`, { headers: { ...authHeaders() } });
  if (!r.ok) throw new Error(`listGames: ${r.statusText}`);
  return r.json() as Promise<GameInfo[]>;
}

export function buildWsUrl(origin: string, gameId: string, room: string, token: string): string {
  const proto = origin.startsWith('https') ? 'wss' : 'ws';
  const cleanHost = origin.replace(/^https?:\/\//, '');
  const params = new URLSearchParams({ t: token });
  if (room) params.set('room', room);
  return `${proto}://${cleanHost}/plugins/${PLUGIN_ID}/ws/games/${encodeURIComponent(gameId)}?${params.toString()}`;
}
