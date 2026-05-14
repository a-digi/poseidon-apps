import { useState, useEffect, useCallback } from 'react';
import type { HighlightList } from '../api';

interface ListsCarouselProps {
  lists: HighlightList[];
}

const rankGradients = [
  'from-amber-500 to-yellow-400',
  'from-slate-400 to-zinc-300',
  'from-orange-700 to-amber-600',
  'from-violet-700 to-purple-600',
  'from-fuchsia-700 to-pink-600',
  'from-indigo-600 to-blue-500',
  'from-teal-600 to-emerald-500',
  'from-rose-600 to-pink-500',
  'from-cyan-600 to-sky-500',
  'from-lime-600 to-green-500',
];

export function ListsCarousel({ lists }: ListsCarouselProps) {
  const [activeIndex, setActiveIndex] = useState(0);
  const n = lists.length;

  const advance = useCallback(() => {
    setActiveIndex(i => (i + 1) % n);
  }, [n]);

  useEffect(() => {
    const timer = setInterval(advance, 2500);
    return () => clearInterval(timer);
  }, [advance]);

  return (
    <div className="relative w-full h-[440px] overflow-hidden">
      {lists.map((list, idx) => {
        let offset = idx - activeIndex;
        if (offset > n / 2) offset -= n;
        if (offset < -(n / 2)) offset += n;

        const absOffset = Math.abs(offset);
        if (absOffset > 1) return null;

        const isCenter = offset === 0;
        const gradient = rankGradients[idx] ?? 'from-purple-700 to-violet-600';
        const scale = isCenter ? 1 : 0.65;
        const translateXpx = offset * 300;

        return (
          <button
            key={list.crawlerId}
            type="button"
            onClick={() => setActiveIndex(idx)}
            aria-label={`View ${list.displayName}`}
            className="absolute top-1/2 left-1/2 flex flex-col items-center gap-3 focus-visible:outline-none"
            style={{
              transform: `translate(calc(-50% + ${translateXpx}px), -50%) scale(${scale})`,
              filter: isCenter ? 'none' : 'blur(5px)',
              opacity: isCenter ? 1 : 0.5,
              transition: 'transform 0.5s cubic-bezier(0.4,0,0.2,1), filter 0.5s ease, opacity 0.5s ease',
              zIndex: isCenter ? 10 : 1,
            }}
          >
            <div className={`w-96 h-96 rounded-2xl bg-gradient-to-br ${gradient} flex flex-col items-center justify-center shadow-xl overflow-hidden shrink-0`}>
              <span className="text-white font-black text-6xl leading-none">
                {idx + 1}
              </span>
              {isCenter && (
                <span className="text-white/80 text-sm font-semibold mt-4 px-6 text-center leading-snug line-clamp-3">
                  {list.displayName}
                </span>
              )}
            </div>
            {isCenter && list.country && (
              <span className="text-slate-500 text-sm font-medium tracking-wide">{list.country}</span>
            )}
          </button>
        );
      })}
    </div>
  );
}
