export type MessageType = 'hello' | 'input' | 'welcome' | 'snapshot' | 'left' | 'error';

export interface Hello {
  type: 'hello';
  room: string;
  name: string;
}

export interface Input {
  type: 'input';
  up: boolean;
  down: boolean;
  left: boolean;
  right: boolean;
  seq: number;
}

export interface Bounds {
  w: number;
  h: number;
}

export interface Welcome {
  type: 'welcome';
  playerId: string;
  room: string;
  tickHz: number;
  arena: Bounds;
}

export interface PlayerSnapshot {
  id: string;
  name: string;
  x: number;
  y: number;
  color: string;
}

export interface Snapshot {
  type: 'snapshot';
  tick: number;
  players: PlayerSnapshot[];
}

export interface Left {
  type: 'left';
  playerId: string;
}

export interface ErrorMsg {
  type: 'error';
  message: string;
}

export type ServerMessage = Welcome | Snapshot | Left | ErrorMsg;
