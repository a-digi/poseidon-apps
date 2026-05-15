const HOST = 'http://localhost:2014';
const PLUGIN_ID = 'coco-gg';

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

export async function startPlugin(): Promise<LongRunningStatus> {
  const r = await fetch(`${HOST}/api/plugins/long-running/start`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
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
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pluginId: PLUGIN_ID }),
  });
  if (!r.ok) throw new Error(`stopPlugin: ${r.statusText}`);
}

export async function getStatus(): Promise<LongRunningStatus> {
  const r = await fetch(
    `${HOST}/api/plugins/long-running/status?pluginId=${encodeURIComponent(PLUGIN_ID)}`,
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
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pluginId: PLUGIN_ID, ttlSeconds }),
  });
  if (!r.ok) throw new Error(`mobile-session: ${r.statusText}`);
  return r.json() as Promise<MobileSession>;
}

export async function getLanAddresses(): Promise<string[]> {
  const r = await fetch(`${HOST}/api/system/lan-addresses`);
  if (!r.ok) return [];
  const j = (await r.json()) as { addresses?: string[] };
  return j.addresses ?? [];
}

export async function listGames(): Promise<GameInfo[]> {
  const r = await fetch(`${HOST}/plugins/${PLUGIN_ID}/api/games`);
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
