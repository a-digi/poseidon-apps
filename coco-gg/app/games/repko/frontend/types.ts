export type Phase = 'lobby' | 'civ_pick' | 'tile_pick' | 'playing' | 'game_over';

export type ResourceType = 'credits' | 'steel' | 'fuel' | 'none';

export type UnitType =
  | 'riflemen' | 'marines' | 'snipers' | 'engineers' | 'paratroopers' | 'commandos'
  | 'apc' | 'light_tank' | 'medium_tank' | 'anti_tank' | 'heavy_tank'
  | 'mortar' | 'howitzer' | 'rocket_artillery'
  | 'drone_swarm' | 'helicopter' | 'fighter_jet' | 'bomber' | 'stealth_bomber'
  | 'naval_strike' | 'missile_launcher' | 'cyber_unit';

export type UnitClass = 'infantry' | 'armor' | 'artillery' | 'air' | 'special';

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
  startingLoadout?: Partial<Record<UnitType, number>>;
  unitRoster?: UnitType[];
  incomePercent?: number;
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
  deadlineMs?: number;
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
  maxRounds?: number;
  roundNumber?: number;
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

export interface BuyTile {
  type: 'buy_tile';
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

export interface UpgradeTile {
  type: 'upgrade_tile';
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
  | BuyTile
  | UpgradeTile
  | EndTurn
  | LeaveGame;
