import { HOST, authHeaders } from '@shell/api';
import type { Phase } from './types';

const PLUGIN_ID = 'coco-gg';
const GAME_ID = 'repko';
const base = `${HOST}/plugins/${encodeURIComponent(PLUGIN_ID)}/api/games/${GAME_ID}`;

async function fetchOrThrow(label: string, url: string, init?: RequestInit): Promise<Response> {
  const method = init?.method ?? 'GET';
  try {
    return await fetch(url, init);
  } catch (cause) {
    const reason = cause instanceof Error ? cause.message : String(cause);
    throw new Error(`${label} ${method} ${url}: network error: ${reason}`);
  }
}

async function throwHTTP(label: string, method: string, url: string, r: Response): Promise<never> {
  let body = '';
  try {
    body = (await r.text()).slice(0, 200);
  } catch {
    body = '';
  }
  throw new Error(`${label} ${method} ${url} -> ${r.status} ${r.statusText}${body ? `: ${body}` : ''}`);
}

async function requestJSON<T>(label: string, url: string, init?: RequestInit): Promise<T> {
  const r = await fetchOrThrow(label, url, init);
  if (!r.ok) await throwHTTP(label, init?.method ?? 'GET', url, r);
  return r.json() as Promise<T>;
}

async function requestNoContent(label: string, url: string, init?: RequestInit): Promise<void> {
  const r = await fetchOrThrow(label, url, init);
  if (!r.ok) await throwHTTP(label, init?.method ?? 'GET', url, r);
}

async function requestNoContentTolerate404(label: string, url: string, init?: RequestInit): Promise<void> {
  const r = await fetchOrThrow(label, url, init);
  if (r.ok || r.status === 404) return;
  await throwHTTP(label, init?.method ?? 'GET', url, r);
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
  phase: Phase;
  players: PlayerRef[];
  createdAt: number;
}

export interface RoomsStats {
  activeRooms: number;
  totalPlayers: number;
  activeGames: number;
}

export interface RoomsList {
  rooms: RoomStatus[];
  stats: RoomsStats;
}

export async function createRoom(expectedPlayers: number, maxRounds: number): Promise<RoomCreated> {
  return requestJSON<RoomCreated>('createRoom', `${base}/rooms`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify({ expectedPlayers, maxRounds }),
  });
}

export async function listRooms(): Promise<RoomsList> {
  return requestJSON<RoomsList>('listRooms', `${base}/rooms`, { headers: { ...authHeaders() } });
}

export async function getRoom(code: string): Promise<RoomStatus | null> {
  const url = `${base}/rooms/${encodeURIComponent(code)}`;
  const init: RequestInit = { headers: { ...authHeaders() } };
  const r = await fetchOrThrow('getRoom', url, init);
  if (r.status === 404) return null;
  if (!r.ok) await throwHTTP('getRoom', 'GET', url, r);
  return r.json() as Promise<RoomStatus>;
}

export async function destroyRoom(code: string): Promise<void> {
  return requestNoContentTolerate404('destroyRoom', `${base}/rooms/${encodeURIComponent(code)}`, {
    method: 'DELETE',
    headers: { ...authHeaders() },
  });
}

export async function kickPlayer(code: string, playerId: string): Promise<void> {
  return requestNoContentTolerate404(
    'kickPlayer',
    `${base}/rooms/${encodeURIComponent(code)}/players/${encodeURIComponent(playerId)}`,
    { method: 'DELETE', headers: { ...authHeaders() } },
  );
}

export async function leaveRoom(code: string, resumeToken: string): Promise<void> {
  return requestNoContentTolerate404('leaveRoom', `${base}/rooms/${encodeURIComponent(code)}/leave`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify({ resumeToken }),
  });
}

export async function startGame(code: string): Promise<void> {
  return requestNoContent('startGame', `${base}/rooms/${encodeURIComponent(code)}/start`, {
    method: 'POST',
    headers: { ...authHeaders() },
  });
}
