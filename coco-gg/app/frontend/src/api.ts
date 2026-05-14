const HOST = 'http://localhost:2014';

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

export async function startPlugin(pluginId: string): Promise<LongRunningStatus> {
  const r = await fetch(`${HOST}/api/plugins/long-running/start`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pluginId }),
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

export async function stopPlugin(pluginId: string): Promise<void> {
  const r = await fetch(`${HOST}/api/plugins/long-running/stop`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pluginId }),
  });
  if (!r.ok) throw new Error(`stopPlugin: ${r.statusText}`);
}

export async function getStatus(pluginId: string): Promise<LongRunningStatus> {
  const r = await fetch(
    `${HOST}/api/plugins/long-running/status?pluginId=${encodeURIComponent(pluginId)}`,
  );
  if (!r.ok) throw new Error(`getStatus: ${r.statusText}`);
  const env = (await r.json()) as { status: string; message: LongRunningStatus | string };
  if (env.status !== 'success') {
    throw new Error(typeof env.message === 'string' ? env.message : 'getStatus failed');
  }
  return env.message as LongRunningStatus;
}

export async function createMobileSession(
  pluginId: string,
  ttlSeconds = 3600,
): Promise<MobileSession> {
  const r = await fetch(`${HOST}/api/mobile-sessions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pluginId, ttlSeconds }),
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

export function buildWsUrl(origin: string, room: string, token: string): string {
  const cleanOrigin = origin.replace(/^https?:\/\//, '').replace(/^wss?:\/\//, '');
  const proto = origin.startsWith('https') || origin.startsWith('wss') ? 'wss' : 'ws';
  const params = new URLSearchParams({ t: token });
  if (room) params.set('room', room);
  return `${proto}://${cleanOrigin}/plugins/coco-gg/ws?${params.toString()}`;
}

export { HOST };
