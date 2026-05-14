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

export interface RoomCreated {
  code: string;
}

export interface PlayerRef {
  id: string;
  name: string;
}

export interface RoomStatus {
  code: string;
  players: PlayerRef[];
  createdAt: number;
}

export interface RoomsStats {
  activeRooms: number;
  totalPlayers: number;
}

export interface RoomsList {
  rooms: RoomStatus[];
  stats: RoomsStats;
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

export async function createRoom(pluginId: string): Promise<RoomCreated> {
  const r = await fetch(`${HOST}/plugins/${encodeURIComponent(pluginId)}/api/rooms`, {
    method: 'POST',
  });
  if (!r.ok) throw new Error(`createRoom: ${r.statusText}`);
  return r.json() as Promise<RoomCreated>;
}

export async function listRooms(pluginId: string): Promise<RoomsList> {
  const r = await fetch(`${HOST}/plugins/${encodeURIComponent(pluginId)}/api/rooms`);
  if (!r.ok) throw new Error(`listRooms: ${r.statusText}`);
  return r.json() as Promise<RoomsList>;
}

export async function getRoom(pluginId: string, code: string): Promise<RoomStatus | null> {
  const r = await fetch(
    `${HOST}/plugins/${encodeURIComponent(pluginId)}/api/rooms/${encodeURIComponent(code)}`,
  );
  if (r.status === 404) return null;
  if (!r.ok) throw new Error(`getRoom: ${r.statusText}`);
  return r.json() as Promise<RoomStatus>;
}

export async function destroyRoom(pluginId: string, code: string): Promise<void> {
  const r = await fetch(
    `${HOST}/plugins/${encodeURIComponent(pluginId)}/api/rooms/${encodeURIComponent(code)}`,
    { method: 'DELETE' },
  );
  if (!r.ok && r.status !== 404) throw new Error(`destroyRoom: ${r.statusText}`);
}

export async function kickPlayer(pluginId: string, code: string, playerId: string): Promise<void> {
  const url = `${HOST}/plugins/${encodeURIComponent(pluginId)}/api/rooms/${encodeURIComponent(code)}/players/${encodeURIComponent(playerId)}`;
  const r = await fetch(url, { method: 'DELETE' });
  if (!r.ok && r.status !== 404) throw new Error(`kickPlayer: ${r.statusText}`);
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
