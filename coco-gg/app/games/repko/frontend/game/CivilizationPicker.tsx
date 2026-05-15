import { useState } from 'react';
import type { ClientAction, Civilization, PlayerState, StateMsg, UnitType } from '../types';
import { UNIT_ICON } from './units';

interface CivilizationPickerProps {
  state: StateMsg;
  myPlayerId: string | null;
  onAction: (action: ClientAction) => void;
}

interface CivTileProps {
  civ: Civilization;
  status: 'available' | 'tentative' | 'mine' | 'taken';
  onPick: () => void;
}

function CivTile({ civ, status, onPick }: CivTileProps) {
  const disabled = status === 'mine' || status === 'taken';
  const baseCls =
    'flex flex-col items-center justify-center gap-1 rounded-md border p-3 text-center transition-colors';
  const statusCls =
    status === 'mine'
      ? 'border-emerald-400 bg-emerald-50 text-emerald-900'
      : status === 'tentative'
        ? 'border-emerald-400 bg-emerald-50 text-emerald-900'
        : status === 'taken'
          ? 'border-slate-200 bg-slate-100 text-slate-400 opacity-50'
          : 'border-slate-200 bg-white text-slate-900 hover:border-slate-400 hover:bg-slate-50';
  const cursorCls = disabled ? 'cursor-not-allowed' : 'cursor-pointer';
  const loadout = civ.startingLoadout;
  const tintBg =
    (status === 'available' || status === 'tentative') && civ.backgroundColor !== ''
      ? { backgroundColor: `${civ.backgroundColor}40` }
      : undefined;
  return (
    <button
      type="button"
      onClick={onPick}
      disabled={disabled}
      className={`${baseCls} ${statusCls} ${cursorCls}`}
      style={tintBg}
    >
      <span className="text-3xl leading-none">{civ.flag}</span>
      <span className="text-base leading-none">{civ.coatOfArms}</span>
      <span className="text-xs font-medium leading-tight">{civ.name}</span>
      <span className="flex items-center gap-1">
        <span className="inline-block h-2.5 w-2.5 rounded-full" style={{ background: civ.color }} />
        {status === 'mine' && (
          <span className="text-[10px] font-semibold uppercase text-emerald-700">Your pick</span>
        )}
        {status === 'tentative' && (
          <span className="text-[10px] font-semibold uppercase text-emerald-700">
            Selected — confirm below
          </span>
        )}
      </span>
      {loadout !== undefined && (
        <span className="text-[10px] text-slate-500">
          {Object.entries(loadout).flatMap(([type, count]) =>
            count > 0 ? [
              <span key={type} className="mr-2">
                {UNIT_ICON[type as UnitType]} {count}
              </span>,
            ] : []
          )}
        </span>
      )}
    </button>
  );
}

interface RosterEntryProps {
  player: PlayerState;
  civ: Civilization | undefined;
}

function RosterEntry({ player, civ }: RosterEntryProps) {
  return (
    <li className="flex items-center gap-2 text-xs">
      <span className="inline-block h-2.5 w-2.5 rounded-full" style={{ background: player.color }} />
      <span className="font-medium text-slate-700">{player.name}</span>
      {civ !== undefined ? (
        <span className="text-base leading-none">{civ.flag}</span>
      ) : (
        <span className="text-[10px] italic text-slate-400">(choosing…)</span>
      )}
    </li>
  );
}

export function CivilizationPicker({ state, myPlayerId, onAction }: CivilizationPickerProps) {
  const [tentativeCivId, setTentativeCivId] = useState<string | null>(null);

  const civs = state.civilizations ?? [];
  if (civs.length === 0) {
    return (
      <div className="flex h-full flex-1 items-center justify-center text-xs text-slate-500">
        Loading civilizations…
      </div>
    );
  }

  const ownerByCiv = new Map<string, PlayerState>();
  for (const p of state.players) {
    if (p.civilizationId !== '') ownerByCiv.set(p.civilizationId, p);
  }

  const myCommittedCivId =
    myPlayerId !== null
      ? state.players.find((p) => p.id === myPlayerId)?.civilizationId ?? ''
      : '';

  const handlePick = (civId: string) => {
    if (myCommittedCivId !== '') return;
    setTentativeCivId(civId);
  };

  const chosenCiv = myCommittedCivId !== '' ? civs.find((c) => c.id === myCommittedCivId) : undefined;
  const tentativeCiv = tentativeCivId !== null ? civs.find((c) => c.id === tentativeCivId) : undefined;

  const showConfirmBar = myCommittedCivId === '' && tentativeCivId !== null && tentativeCiv !== undefined;

  return (
    <div className="fixed inset-0 flex flex-col bg-slate-50">
      <header className="border-b border-slate-200 bg-white px-4 py-3">
        <h1 className="text-base font-semibold text-slate-900">
          {myCommittedCivId !== '' ? 'Waiting for other players…' : 'Choose your civilization'}
        </h1>
        <p className="text-xs text-slate-500">
          {myCommittedCivId !== '' ? (
            chosenCiv !== undefined ? (
              <>
                Your civilization has been locked in: <span className="text-base">{chosenCiv.flag}</span>{' '}
                <span className="font-medium text-slate-700">{chosenCiv.name}</span>.
              </>
            ) : (
              'Your civilization has been locked in.'
            )
          ) : tentativeCivId === null ? (
            'Pick one to claim it.'
          ) : (
            'Press Confirm to lock in your choice.'
          )}
        </p>
      </header>

      <main className="flex-1 overflow-auto px-3 pt-3 pb-32">
        <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 md:grid-cols-4">
          {civs.map((civ) => {
            const owner = ownerByCiv.get(civ.id);
            const status: 'available' | 'tentative' | 'mine' | 'taken' =
              owner !== undefined
                ? owner.id === myPlayerId
                  ? 'mine'
                  : 'taken'
                : myCommittedCivId === '' && tentativeCivId === civ.id
                  ? 'tentative'
                  : 'available';
            return (
              <CivTile
                key={civ.id}
                civ={civ}
                status={status}
                onPick={status === 'available' ? () => handlePick(civ.id) : () => {}}
              />
            );
          })}
        </div>
      </main>

      {showConfirmBar && (
        <div className="border-t border-slate-200 bg-white px-4 py-3 flex items-center gap-2">
          <button
            type="button"
            onClick={() => setTentativeCivId(null)}
            className="rounded-md bg-slate-200 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-300"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={() => onAction({ type: 'pick_civilization', civilizationId: tentativeCiv.id })}
            className="flex-1 rounded-md bg-slate-900 px-3 py-2 text-sm font-medium text-white hover:bg-slate-700"
          >
            Confirm <span className="text-base leading-none">{tentativeCiv.flag}</span> {tentativeCiv.name}
          </button>
        </div>
      )}

      <footer className="border-t border-slate-200 bg-white px-4 py-3">
        <p className="mb-2 text-[10px] font-semibold uppercase tracking-wide text-slate-500">
          Players
        </p>
        <ul className="flex flex-wrap gap-x-4 gap-y-1">
          {state.players.map((p) => {
            const civ = p.civilizationId !== '' ? civs.find((c) => c.id === p.civilizationId) : undefined;
            return <RosterEntry key={p.id} player={p} civ={civ} />;
          })}
        </ul>
      </footer>
    </div>
  );
}

export default CivilizationPicker;
