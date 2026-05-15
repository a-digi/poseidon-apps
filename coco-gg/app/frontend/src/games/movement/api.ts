const HOST = 'http://localhost:2014';
const PLUGIN_ID = 'coco-gg';
const GAME_ID = 'movement';

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

const base = `${HOST}/plugins/${PLUGIN_ID}/api/games/${GAME_ID}`;

export async function createRoom(): Promise<RoomCreated> {
  const r = await fetch(`${base}/rooms`, { method: 'POST' });
  if (!r.ok) throw new Error(`createRoom: ${r.statusText}`);
  return r.json() as Promise<RoomCreated>;
}

export async function listRooms(): Promise<RoomsList> {
  const r = await fetch(`${base}/rooms`);
  if (!r.ok) throw new Error(`listRooms: ${r.statusText}`);
  return r.json() as Promise<RoomsList>;
}

export async function getRoom(code: string): Promise<RoomStatus | null> {
  const r = await fetch(`${base}/rooms/${encodeURIComponent(code)}`);
  if (r.status === 404) return null;
  if (!r.ok) throw new Error(`getRoom: ${r.statusText}`);
  return r.json() as Promise<RoomStatus>;
}

export async function destroyRoom(code: string): Promise<void> {
  const r = await fetch(`${base}/rooms/${encodeURIComponent(code)}`, { method: 'DELETE' });
  if (!r.ok && r.status !== 404) throw new Error(`destroyRoom: ${r.statusText}`);
}

export async function kickPlayer(code: string, playerId: string): Promise<void> {
  const r = await fetch(
    `${base}/rooms/${encodeURIComponent(code)}/players/${encodeURIComponent(playerId)}`,
    { method: 'DELETE' },
  );
  if (!r.ok && r.status !== 404) throw new Error(`kickPlayer: ${r.statusText}`);
}
