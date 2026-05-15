import type { GameEvent } from '../types';

export function eventToText(event: GameEvent): string {
  const name = event.actorName || 'Someone';
  const tile = event.tileName ? ` ${event.tileName}` : '';
  const defender = event.defenderName ?? 'an enemy';
  switch (event.kind) {
    case 'recruit':
      return `${name} recruited ${event.unitCount ?? 0} ${event.unit ?? 'units'}${tile ? ` on${tile}` : ''}`;
    case 'attack_won':
      return `${name} captured${tile || ' a tile'}`;
    case 'attack_lost':
      return `${name} attacked${tile || ' a tile'} and was repulsed`;
    case 'attack_tie':
      return `${name} fought ${defender} to a draw${tile}`;
    case 'buy_tile':
      return `${name} bought${tile || ' a tile'}`;
    case 'move':
      return `${name} moved units${tile ? ` to${tile}` : ''}`;
    case 'upgrade':
      return `${name} upgraded ${event.unit ?? 'units'}${tile ? ` on${tile}` : ''}`;
    case 'upgrade_tile':
      return `${name} improved production${tile ? ` at${tile}` : ''}`;
    case 'offer_diplomacy':
      return `${name} offered diplomacy to ${defender}`;
    case 'accept_diplomacy':
      return `${name} accepted a diplomacy offer`;
    case 'decline_diplomacy':
      return `${name} declined a diplomacy offer`;
    case 'pick_civilization':
      return `${name} chose a civilization`;
    case 'pick_starting_tile':
      return `${name} placed their capital${tile ? ` at${tile}` : ''}`;
    case 'end_turn':
      return `${name} ended their turn`;
  }
}
