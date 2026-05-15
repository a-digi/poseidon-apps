import type { ComponentType } from 'react';
import { RepkoLogo } from '@games/repko/frontend/RepkoLogo';

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
    id: 'repko',
    name: 'Repko',
    description:
      'A turn-based hex-board game of resource gathering and settlement, for 3-6 players.',
    Logo: RepkoLogo,
    loadApp: () => import('@games/repko/frontend/App'),
    loadMobile: () => import('@games/repko/frontend/MobilePage'),
  },
];
