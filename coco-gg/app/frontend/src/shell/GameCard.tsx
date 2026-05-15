import type { GameDescriptor } from './games-registry';

interface GameCardProps {
  game: GameDescriptor;
  index: number;
  onOpen: () => void;
}

export function GameCard({ game, index, onOpen }: GameCardProps) {
  return (
    <button
      type="button"
      onClick={onOpen}
      aria-label={`Open ${game.name}`}
      style={{
        animationDuration: '900ms',
        animationDelay: `${index * 250}ms`,
      }}
      className="group relative overflow-hidden rounded-2xl aspect-[4/3] w-full text-left
        bg-gradient-to-br from-white to-slate-50 border border-slate-200
        shadow-sm hover:shadow-md hover:border-indigo-200
        transition-all duration-500 ease-out
        hover:-translate-y-1 hover:scale-[1.01]
        animate-in fade-in slide-in-from-bottom-8
        [animation-fill-mode:both]"
    >
      <div className="relative flex h-full flex-col p-6">
        <div className="flex-1">
          <game.Logo className="mb-3 h-10 w-10 text-slate-700" />
          <h3 className="text-2xl font-bold leading-tight tracking-tight text-slate-900">
            {game.name}
          </h3>
          <p className="mt-2 text-sm leading-relaxed text-slate-500">{game.description}</p>
        </div>
        <div className="mt-auto flex items-center justify-between border-t border-slate-100 pt-4">
          <span className="text-xs font-medium uppercase tracking-widest text-slate-400">
            Game
          </span>
          <span className="text-xs text-slate-400 transition-colors duration-500 group-hover:text-indigo-600">
            Open →
          </span>
        </div>
      </div>
    </button>
  );
}
