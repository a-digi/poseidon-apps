export type Phase = 'lobby' | 'civ_pick' | 'tile_pick' | 'playing' | 'game_over';

export type ResourceType = 'gold' | 'iron' | 'food' | 'none';

export type UnitType = 'infantry' | 'cavalry' | 'artillery';

export type UnitLevel = 1 | 2 | 3;

export interface GarrisonStack {
  type: UnitType;
  level: UnitLevel;
  count: number;
}

export interface StackPick {
  stackIndex: number;
  count: number;
}

export interface Tile {
  q: number;
  r: number;
  production: ResourceType;
  yields: Partial<Record<ResourceType, number>>;
  ownerId: string;
  foundedBy?: string;
  name?: string;
  garrison: GarrisonStack[];
}

export interface Board {
  tiles: Tile[];
}

export interface Civilization {
  id: string;
  name: string;
  color: string;
  backgroundColor: string;
  flag: string;
  coatOfArms: string;
  startingLoadout?: Record<UnitType, number>;
}

export interface DiplomacyOffer {
  q: number;
  r: number;
  attackerId: string;
  defenderId: string;
}

export interface PlayerState {
  id: string;
  name: string;
  color: string;
  civilizationId: string;
  tileCount: number;
  unitCount: number;
  eliminated: boolean;
}

export interface CurrentTurn {
  playerId: string;
}

export interface YouState {
  resources: Partial<Record<ResourceType, number>>;
}

export interface Welcome {
  type: 'welcome';
  playerId: string;
  room: string;
  resumeToken: string;
  you: { name: string; color: string };
}

export interface StateMsg {
  type: 'state';
  phase: Phase;
  board?: Board;
  players: PlayerState[];
  currentTurn?: CurrentTurn;
  civilizations?: Civilization[];
  pendingDiplomacy?: DiplomacyOffer[];
  winnerId?: string;
  you?: YouState;
}

export interface ErrorMsg {
  type: 'error';
  message: string;
}

export type ServerMessage = Welcome | StateMsg | ErrorMsg;

export interface Hello {
  type: 'hello';
  room: string;
  name: string;
  resumeToken?: string;
}

export interface PickCivilization {
  type: 'pick_civilization';
  civilizationId: string;
}

export interface PickStartingTile {
  type: 'pick_starting_tile';
  q: number;
  r: number;
}

export interface Recruit {
  type: 'recruit';
  q: number;
  r: number;
  unit: UnitType;
  count: number;
}

export interface Upgrade {
  type: 'upgrade';
  q: number;
  r: number;
  stackIndex: number;
}

export interface Move {
  type: 'move';
  fromQ: number;
  fromR: number;
  toQ: number;
  toR: number;
  units: StackPick[];
}

export interface Attack {
  type: 'attack';
  fromQ: number;
  fromR: number;
  toQ: number;
  toR: number;
  units: StackPick[];
}

export interface OfferDiplomacy {
  type: 'offer_diplomacy';
  q: number;
  r: number;
}

export interface AcceptDiplomacy {
  type: 'accept_diplomacy';
  q: number;
  r: number;
}

export interface DeclineDiplomacy {
  type: 'decline_diplomacy';
  q: number;
  r: number;
}

export interface CancelDiplomacy {
  type: 'cancel_diplomacy';
  q: number;
  r: number;
}

export interface EndTurn {
  type: 'end_turn';
}

export interface LeaveGame {
  type: 'leave_game';
}

export type ClientAction =
  | PickCivilization
  | PickStartingTile
  | Recruit
  | Upgrade
  | Move
  | Attack
  | OfferDiplomacy
  | AcceptDiplomacy
  | DeclineDiplomacy
  | CancelDiplomacy
  | EndTurn
  | LeaveGame;
