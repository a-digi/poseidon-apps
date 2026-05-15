import type { ComponentType } from 'react';
import { MovementLogo } from '../games/movement/MovementLogo';

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
];
