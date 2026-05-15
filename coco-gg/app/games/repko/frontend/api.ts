import { HOST, authHeaders } from '@shell/api';
import type { Phase } from './types';

const PLUGIN_ID = 'coco-gg';
const GAME_ID = 'repko';
const base = `${HOST}/plugins/${encodeURIComponent(PLUGIN_ID)}/api/games/${GAME_ID}`;

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
  const r = await fetch(`${base}/rooms`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify({ expectedPlayers, maxRounds }),
  });
  if (!r.ok) throw new Error(`createRoom: ${r.statusText}`);
  return r.json() as Promise<RoomCreated>;
}

export async function listRooms(): Promise<RoomsList> {
  const r = await fetch(`${base}/rooms`, { headers: { ...authHeaders() } });
  if (!r.ok) throw new Error(`listRooms: ${r.statusText}`);
  return r.json() as Promise<RoomsList>;
}

export async function getRoom(code: string): Promise<RoomStatus | null> {
  const r = await fetch(`${base}/rooms/${encodeURIComponent(code)}`, { headers: { ...authHeaders() } });
  if (r.status === 404) return null;
  if (!r.ok) throw new Error(`getRoom: ${r.statusText}`);
  return r.json() as Promise<RoomStatus>;
}

export async function destroyRoom(code: string): Promise<void> {
  const r = await fetch(`${base}/rooms/${encodeURIComponent(code)}`, {
    method: 'DELETE',
    headers: { ...authHeaders() },
  });
  if (!r.ok && r.status !== 404) throw new Error(`destroyRoom: ${r.statusText}`);
}

export async function kickPlayer(code: string, playerId: string): Promise<void> {
  const r = await fetch(
    `${base}/rooms/${encodeURIComponent(code)}/players/${encodeURIComponent(playerId)}`,
    { method: 'DELETE', headers: { ...authHeaders() } },
  );
  if (!r.ok && r.status !== 404) throw new Error(`kickPlayer: ${r.statusText}`);
}

export async function leaveRoom(code: string, resumeToken: string): Promise<void> {
  const r = await fetch(`${base}/rooms/${encodeURIComponent(code)}/leave`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify({ resumeToken }),
  });
  if (!r.ok && r.status !== 404) {
    throw new Error(`leaveRoom: ${r.statusText}`);
  }
}

export async function startGame(code: string): Promise<void> {
  const r = await fetch(`${base}/rooms/${encodeURIComponent(code)}/start`, {
    method: 'POST',
    headers: { ...authHeaders() },
  });
  if (!r.ok) {
    let detail = r.statusText;
    try {
      const body = (await r.json()) as { error?: unknown };
      if (typeof body?.error === 'string') detail = body.error;
    } catch {
      /* ignore */
    }
    throw new Error(`startGame: ${detail}`);
  }
}
