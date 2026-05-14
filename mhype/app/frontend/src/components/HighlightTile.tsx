interface HighlightTileProps {
  artworkUrl?: string;
  primaryLabel: string;
  secondaryLabel?: string;
  count: number;
  position?: number;
  size?: 'md' | 'lg' | '2xl' | 'xl';
  fill?: boolean;
  onClick?: () => void;
  ariaLabel?: string;
}

export function HighlightTile({ artworkUrl, primaryLabel, secondaryLabel, count, position, size = 'md', fill = false, onClick, ariaLabel }: HighlightTileProps) {
  const dim = fill
    ? 'w-full aspect-square'
    : size === 'xl' ? 'w-96 h-96' : size === '2xl' ? 'w-56 h-56' : size === 'lg' ? 'w-36 h-36' : 'w-24 h-24';
  const wrapperBase = fill ? 'w-full flex flex-col gap-1.5' : `${size === 'xl' ? 'w-96' : size === '2xl' ? 'w-56' : size === 'lg' ? 'w-36' : 'w-24'} shrink-0 flex flex-col gap-1.5`;

  const wrapperClass = [
    wrapperBase,
    onClick ? 'text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-purple-400 rounded-xl' : '',
  ].join(' ');

  const artwork = (
    <div className={`relative ${dim} rounded-xl overflow-hidden bg-gradient-to-br from-slate-200 to-slate-300 ${fill ? '' : 'shrink-0'}`}>
      {artworkUrl && <img src={artworkUrl} alt="" className="w-full h-full object-cover" />}
      {fill ? (
        <>
          <div className="absolute inset-x-0 bottom-0 h-2/3 bg-gradient-to-t from-black/80 to-transparent" />
          {position !== undefined && (
            <span className="absolute top-2 left-2 bg-black/60 text-white text-xs font-bold rounded px-1.5 py-0.5 leading-none">
              #{position}
            </span>
          )}
          <p className="absolute bottom-2 left-2 right-2 text-white text-lg font-semibold truncate leading-none">
            {primaryLabel}
          </p>
        </>
      ) : (
        <span className="absolute top-1 right-1 bg-black/60 text-white text-xs font-bold rounded-full px-1.5 py-0.5 leading-none">
          ×{count}
        </span>
      )}
    </div>
  );

  const labels = fill ? null : (
    <>
      <p className="text-xs font-semibold text-white truncate">{primaryLabel}</p>
      {secondaryLabel && <p className="text-xs text-white/60 truncate">{secondaryLabel}</p>}
    </>
  );

  if (onClick) {
    return (
      <button type="button" className={wrapperClass} onClick={onClick} aria-label={ariaLabel}>
        {artwork}
        {labels}
      </button>
    );
  }

  return (
    <div className={wrapperClass}>
      {artwork}
      {labels}
    </div>
  );
}
