import type { ComponentType } from 'react';
import { MovementLogo } from '../games/movement/MovementLogo';
import { RepkoLogo } from '../games/repko/RepkoLogo';

export interface GameDescriptor {
  id: string;
  name: string;
  description: string;
  Logo: ComponentType<{ className?: string }>;
  loadApp: () => Promise<{ default: ComponentType }>;
  loadMobile: () => Promise<{ default: ComponentType }>;
}

export const GAMES: GameDescriptor[] = [
  {
    id: 'movement',
    name: 'Movement Arena',
    description:
      'A shared 2D arena where players move colored circles. Placeholder scaffold for richer mechanics.',
    Logo: MovementLogo,
    loadApp: () => import('../games/movement/App'),
    loadMobile: () => import('../games/movement/MobilePage'),
  },
  {
    id: 'repko',
    name: 'Repko',
    description:
      'A turn-based hex-board game of resource gathering and settlement, for 3-6 players.',
    Logo: RepkoLogo,
    loadApp: () => import('../games/repko/App'),
    loadMobile: () => import('../games/repko/MobilePage'),
  },
];
