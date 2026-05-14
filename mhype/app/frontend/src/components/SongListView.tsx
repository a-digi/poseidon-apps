import { useState } from 'react';
import { ChevronLeft, ChevronRight, Info, Loader2, Play } from 'lucide-react';

export interface SongListEntry {
  id: string;
  position: number;
  title: string;
  artists: string[];
  artworkUrl?: string;
  chartName?: string;
  country?: string;
  crawlerId?: string;
}

interface SongListViewProps {
  entries: SongListEntry[];
  paletteId: string;
  loadingId: string | null;
  onPlay: (entry: SongListEntry) => void;
  onSearch: (artist: string, title: string) => void;
}

const PAGE_SIZE = 10;

const PALETTES: readonly (readonly string[])[] = [
  ['bg-amber-500', 'bg-orange-500', 'bg-yellow-500', 'bg-amber-600', 'bg-orange-600',
   'bg-yellow-600', 'bg-amber-700', 'bg-orange-700', 'bg-yellow-700', 'bg-amber-800'],
  ['bg-blue-600', 'bg-indigo-600', 'bg-sky-600', 'bg-cyan-600', 'bg-blue-700',
   'bg-indigo-700', 'bg-sky-700', 'bg-cyan-700', 'bg-blue-800', 'bg-indigo-800'],
  ['bg-emerald-600', 'bg-teal-600', 'bg-green-600', 'bg-emerald-700', 'bg-teal-700',
   'bg-green-700', 'bg-emerald-800', 'bg-teal-800', 'bg-green-800', 'bg-emerald-500'],
  ['bg-fuchsia-600', 'bg-purple-600', 'bg-pink-600', 'bg-violet-600', 'bg-fuchsia-700',
   'bg-purple-700', 'bg-pink-700', 'bg-violet-700', 'bg-fuchsia-800', 'bg-purple-800'],
];

function paletteForId(id: string): readonly string[] {
  let h = 0;
  for (let i = 0; i < id.length; i++) h = (h * 31 + id.charCodeAt(i)) >>> 0;
  return PALETTES[h % PALETTES.length];
}

type PageDir = 'initial' | 'fwd' | 'bwd';

const PAGE_ENTER: Record<PageDir, string> = {
  initial: 'animate-in fade-in slide-in-from-bottom-2 duration-300',
  fwd: 'animate-in fade-in slide-in-from-right-4 duration-300',
  bwd: 'animate-in fade-in slide-in-from-left-4 duration-300',
};

export function SongListView({ entries, paletteId, loadingId, onPlay, onSearch }: SongListViewProps) {
  const [page, setPage] = useState(0);
  const [pageDir, setPageDir] = useState<PageDir>('initial');
  const [pageKey, setPageKey] = useState(0);

  const palette = paletteForId(paletteId);
  const totalPages = Math.ceil(entries.length / PAGE_SIZE);
  const pageSlice = entries.slice(page * PAGE_SIZE, (page + 1) * PAGE_SIZE);

  function goNext() {
    setPageDir('fwd');
    setPage(p => p + 1);
    setPageKey(k => k + 1);
  }

  function goPrev() {
    setPageDir('bwd');
    setPage(p => p - 1);
    setPageKey(k => k + 1);
  }

  return (
    <div className="flex flex-col gap-2">
      <ul
        key={pageKey}
        className={`flex flex-col divide-y divide-slate-100 border border-slate-200 rounded-xl bg-white shadow-sm overflow-hidden ${PAGE_ENTER[pageDir]}`}
      >
        {pageSlice.map(entry => {
          const bg = entry.position >= 1 && entry.position <= 10
            ? palette[entry.position - 1]
            : 'bg-slate-900';
          const isLoading = loadingId === entry.id;
          return (
            <li key={entry.id}>
              <div className="w-full flex items-center gap-3 px-3 py-2.5 text-left">
                <span className={`shrink-0 w-8 h-8 rounded-md flex items-center justify-center text-xs font-black text-white tabular-nums ${bg}`}>
                  {entry.position}
                </span>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-semibold text-slate-900 truncate leading-tight">{entry.title}</p>
                  {entry.artists.length > 0 && (
                    <p className="text-xs text-slate-400 truncate mt-0.5">{entry.artists[0]}</p>
                  )}
                </div>
                <div className="flex items-center gap-1 shrink-0 ml-2">
                  <button
                    type="button"
                    onClick={() => onSearch(entry.artists[0] ?? '', entry.title)}
                    aria-label={`Search ${entry.title} on Google`}
                    className="p-1.5 rounded-md hover:bg-slate-100 text-slate-500 hover:text-slate-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors"
                  >
                    <Info className="w-4 h-4" />
                  </button>
                  <button
                    type="button"
                    onClick={() => onPlay(entry)}
                    disabled={isLoading}
                    aria-label={`Play ${entry.title} on YouTube`}
                    className="p-1.5 rounded-md hover:bg-slate-100 text-slate-500 hover:text-slate-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 disabled:opacity-40 transition-colors"
                  >
                    {isLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4" />}
                  </button>
                </div>
              </div>
            </li>
          );
        })}
      </ul>

      {totalPages > 1 && (
        <div className="flex items-center justify-between px-1">
          <button
            type="button"
            onClick={goPrev}
            disabled={page === 0}
            aria-label="Previous page"
            className="flex items-center gap-1 px-2 py-1 rounded-lg text-xs font-medium text-slate-600 hover:bg-slate-100 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
          >
            <ChevronLeft className="w-3.5 h-3.5" />
            Prev
          </button>
          <span className="text-xs font-mono text-slate-400 tabular-nums">
            {String(page + 1).padStart(2, '0')} / {String(totalPages).padStart(2, '0')}
          </span>
          <button
            type="button"
            onClick={goNext}
            disabled={page >= totalPages - 1}
            aria-label="Next page"
            className="flex items-center gap-1 px-2 py-1 rounded-lg text-xs font-medium text-slate-600 hover:bg-slate-100 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
          >
            Next
            <ChevronRight className="w-3.5 h-3.5" />
          </button>
        </div>
      )}
    </div>
  );
}
