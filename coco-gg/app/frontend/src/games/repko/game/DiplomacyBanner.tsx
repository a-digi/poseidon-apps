import type { ClientAction, Civilization, DiplomacyOffer, PlayerState, StateMsg } from '../types';

interface DiplomacyBannerProps {
  state: StateMsg;
  myPlayerId: string | null;
  onAction: (action: ClientAction) => void;
}

interface DefenderBannerProps {
  offer: DiplomacyOffer;
  attacker: PlayerState | undefined;
  attackerCiv: Civilization | undefined;
  onAccept: () => void;
  onDecline: () => void;
}

function DefenderBanner({ offer, attacker, attackerCiv, onAccept, onDecline }: DefenderBannerProps) {
  const attackerName = attacker?.name ?? 'Unknown';
  const flag = attackerCiv?.flag ?? '❔';
  return (
    <div className="rounded-md border border-amber-200 bg-amber-50 px-3 py-2 text-amber-900 shadow-sm">
      <p className="text-xs">
        <span className="font-semibold">{attackerName}</span> ({flag}) offers diplomacy on tile (
        {offer.q},{offer.r}).
      </p>
      <div className="mt-2 flex flex-wrap gap-2">
        <button
          type="button"
          onClick={onAccept}
          className="rounded bg-amber-600 px-2 py-1 text-xs font-medium text-white hover:bg-amber-700"
        >
          Accept (retreat home)
        </button>
        <button
          type="button"
          onClick={onDecline}
          className="rounded border border-amber-400 bg-white px-2 py-1 text-xs font-medium text-amber-800 hover:bg-amber-100"
        >
          Decline (prepare for war)
        </button>
      </div>
    </div>
  );
}

interface AttackerBannerProps {
  offer: DiplomacyOffer;
  defender: PlayerState | undefined;
  onCancel: () => void;
}

function AttackerBanner({ offer, defender, onCancel }: AttackerBannerProps) {
  const defenderName = defender?.name ?? 'Unknown';
  return (
    <div className="rounded-md border border-slate-200 bg-slate-100 px-3 py-2 text-slate-700 shadow-sm">
      <p className="text-xs">
        You offered diplomacy on <span className="font-semibold">{defenderName}</span>'s tile (
        {offer.q},{offer.r}).
      </p>
      <div className="mt-2">
        <button
          type="button"
          onClick={onCancel}
          className="rounded border border-slate-300 bg-white px-2 py-1 text-xs font-medium text-slate-700 hover:bg-slate-200"
        >
          Withdraw offer
        </button>
      </div>
    </div>
  );
}

export function DiplomacyBanner({ state, myPlayerId, onAction }: DiplomacyBannerProps) {
  const offers = state.pendingDiplomacy ?? [];
  if (offers.length === 0 || myPlayerId === null) return null;

  const playersById = new Map<string, PlayerState>();
  for (const p of state.players) playersById.set(p.id, p);
  const civs = state.civilizations ?? [];
  const civById = new Map<string, Civilization>();
  for (const c of civs) civById.set(c.id, c);

  const relevant = offers.filter(
    (o) => o.defenderId === myPlayerId || o.attackerId === myPlayerId,
  );
  if (relevant.length === 0) return null;

  return (
    <div className="flex flex-col gap-2 px-3 py-2">
      {relevant.map((offer) => {
        const key = `${offer.q},${offer.r}`;
        if (offer.defenderId === myPlayerId) {
          const attacker = playersById.get(offer.attackerId);
          const attackerCiv =
            attacker !== undefined ? civById.get(attacker.civilizationId) : undefined;
          return (
            <DefenderBanner
              key={key}
              offer={offer}
              attacker={attacker}
              attackerCiv={attackerCiv}
              onAccept={() => onAction({ type: 'accept_diplomacy', q: offer.q, r: offer.r })}
              onDecline={() => onAction({ type: 'decline_diplomacy', q: offer.q, r: offer.r })}
            />
          );
        }
        const defender = playersById.get(offer.defenderId);
        return (
          <AttackerBanner
            key={key}
            offer={offer}
            defender={defender}
            onCancel={() => onAction({ type: 'cancel_diplomacy', q: offer.q, r: offer.r })}
          />
        );
      })}
    </div>
  );
}

export default DiplomacyBanner;
