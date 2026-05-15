import type { UnitClass, UnitType } from '../types';

export interface UnitSpec {
  id: UnitType;
  name: string;
  icon: string;
  class: UnitClass;
  power: number;
  cost: { credits: number; steel: number; fuel: number };
}

export const UNIT_ORDER: UnitType[] = [
  'riflemen', 'marines', 'snipers', 'engineers', 'paratroopers', 'commandos',
  'apc', 'light_tank', 'medium_tank', 'anti_tank', 'heavy_tank',
  'mortar', 'howitzer', 'rocket_artillery',
  'drone_swarm', 'helicopter', 'fighter_jet', 'bomber', 'stealth_bomber',
  'naval_strike', 'missile_launcher', 'cyber_unit',
];

export const UNIT_CATALOG: Record<UnitType, UnitSpec> = {
  riflemen:        { id: 'riflemen',        name: 'Riflemen',         icon: '🪖',   class: 'infantry',  power: 3,  cost: { credits: 5,  steel: 1, fuel: 0  } },
  marines:         { id: 'marines',         name: 'Marines',          icon: '🛡️',   class: 'infantry',  power: 4,  cost: { credits: 8,  steel: 2, fuel: 0  } },
  snipers:         { id: 'snipers',         name: 'Snipers',          icon: '🎯',   class: 'infantry',  power: 5,  cost: { credits: 10, steel: 1, fuel: 0  } },
  engineers:       { id: 'engineers',       name: 'Engineers',        icon: '🛠️',   class: 'infantry',  power: 4,  cost: { credits: 8,  steel: 2, fuel: 1  } },
  paratroopers:    { id: 'paratroopers',    name: 'Paratroopers',     icon: '🪂',   class: 'infantry',  power: 6,  cost: { credits: 12, steel: 2, fuel: 1  } },
  commandos:       { id: 'commandos',       name: 'Commandos',        icon: '🥷',   class: 'infantry',  power: 8,  cost: { credits: 18, steel: 3, fuel: 1  } },
  apc:             { id: 'apc',             name: 'APC',              icon: '🚙',   class: 'armor',     power: 5,  cost: { credits: 8,  steel: 4, fuel: 2  } },
  light_tank:      { id: 'light_tank',      name: 'Light Tank',       icon: '🚜',   class: 'armor',     power: 6,  cost: { credits: 10, steel: 3, fuel: 1  } },
  medium_tank:     { id: 'medium_tank',     name: 'Medium Tank',      icon: '🚛',   class: 'armor',     power: 8,  cost: { credits: 14, steel: 5, fuel: 2  } },
  anti_tank:       { id: 'anti_tank',       name: 'Anti-Tank',        icon: '🛡️',   class: 'armor',     power: 9,  cost: { credits: 12, steel: 4, fuel: 2  } },
  heavy_tank:      { id: 'heavy_tank',      name: 'Heavy Tank',       icon: '🦾',   class: 'armor',     power: 11, cost: { credits: 18, steel: 7, fuel: 3  } },
  mortar:          { id: 'mortar',          name: 'Mortar',           icon: '💥',   class: 'artillery', power: 7,  cost: { credits: 12, steel: 3, fuel: 1  } },
  howitzer:        { id: 'howitzer',        name: 'Howitzer',         icon: '🔫',   class: 'artillery', power: 10, cost: { credits: 16, steel: 5, fuel: 2  } },
  rocket_artillery:{ id: 'rocket_artillery',name: 'Rocket Battery',   icon: '🚀',   class: 'artillery', power: 13, cost: { credits: 22, steel: 6, fuel: 3  } },
  drone_swarm:     { id: 'drone_swarm',     name: 'Drone Swarm',      icon: '🛸',   class: 'air',       power: 8,  cost: { credits: 16, steel: 2, fuel: 4  } },
  helicopter:      { id: 'helicopter',      name: 'Helicopter',       icon: '🚁',   class: 'air',       power: 9,  cost: { credits: 18, steel: 3, fuel: 5  } },
  fighter_jet:     { id: 'fighter_jet',     name: 'Fighter Jet',      icon: '✈️',   class: 'air',       power: 12, cost: { credits: 25, steel: 4, fuel: 7  } },
  bomber:          { id: 'bomber',          name: 'Bomber',           icon: '💣',   class: 'air',       power: 15, cost: { credits: 32, steel: 5, fuel: 10 } },
  stealth_bomber:  { id: 'stealth_bomber',  name: 'Stealth Bomber',   icon: '🦇',   class: 'air',       power: 18, cost: { credits: 45, steel: 7, fuel: 12 } },
  naval_strike:    { id: 'naval_strike',    name: 'Naval Strike',     icon: '🚢',   class: 'special',   power: 12, cost: { credits: 22, steel: 6, fuel: 4  } },
  missile_launcher:{ id: 'missile_launcher',name: 'Missile Launcher', icon: '🛰️',   class: 'special',   power: 16, cost: { credits: 30, steel: 8, fuel: 6  } },
  cyber_unit:      { id: 'cyber_unit',      name: 'Cyberwarfare',     icon: '💻',   class: 'special',   power: 10, cost: { credits: 25, steel: 1, fuel: 1  } },
};

export const BASE_POWER: Record<UnitType, number> =
  Object.fromEntries(UNIT_ORDER.map((u) => [u, UNIT_CATALOG[u].power])) as Record<UnitType, number>;

export const UNIT_ICON: Record<UnitType, string> =
  Object.fromEntries(UNIT_ORDER.map((u) => [u, UNIT_CATALOG[u].icon])) as Record<UnitType, string>;

export const UNIT_NAME: Record<UnitType, string> =
  Object.fromEntries(UNIT_ORDER.map((u) => [u, UNIT_CATALOG[u].name])) as Record<UnitType, string>;

export const UNIT_CLASS: Record<UnitType, UnitClass> =
  Object.fromEntries(UNIT_ORDER.map((u) => [u, UNIT_CATALOG[u].class])) as Record<UnitType, UnitClass>;

export const UNIT_COSTS: Record<UnitType, { credits: number; steel: number; fuel: number }> =
  Object.fromEntries(UNIT_ORDER.map((u) => [u, UNIT_CATALOG[u].cost])) as Record<UnitType, { credits: number; steel: number; fuel: number }>;

export const CLASS_ICON: Record<UnitClass, string> = {
  infantry: '🪖',
  armor: '🚛',
  artillery: '💥',
  air: '✈️',
  special: '🛰️',
};

export const CLASS_NAME: Record<UnitClass, string> = {
  infantry: 'Infantry',
  armor: 'Armor',
  artillery: 'Artillery',
  air: 'Air',
  special: 'Special',
};

export const CLASS_ORDER: UnitClass[] = ['infantry', 'armor', 'artillery', 'air', 'special'];
