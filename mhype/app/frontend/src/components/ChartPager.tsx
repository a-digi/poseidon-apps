import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { ChevronLeft, ChevronRight, Info, Loader2, Play } from 'lucide-react';
import { listItems, type ChartContext, type Item } from '../api';
import { FlipNumber } from './FlipNumber';
import { Spinner } from './Spinner';
import { openVideoSearch } from '../lib/openVideoSearch';
import { usePlayYouTube } from '../lib/usePlayYouTube';
import { VideoModal } from './VideoModal';

interface ChartPagerProps {
  crawlerId: string;
  refreshKey: number;
  chartName?: string;
  country?: string;
  onPlayed?: () => void;
}

type ChartItem = Item & { chart: ChartContext };

type PageDir = 'initial' | 'fwd' | 'bwd';

const PAGE_SIZE = 5;
const FETCH_LIMIT = 100;

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

function paletteForId(crawlerId: string): readonly string[] {
  let h = 0;
  for (let i = 0; i < crawlerId.length; i++) h = (h * 31 + crawlerId.charCodeAt(i)) >>> 0;
  return PALETTES[h % PALETTES.length];
}

const PAGE_ENTER: Record<PageDir, string> = {
  initial: 'animate-in fade-in slide-in-from-bottom-2 duration-300',
  fwd: 'animate-in fade-in slide-in-from-right-4 duration-300',
  bwd: 'animate-in fade-in slide-in-from-left-4 duration-300',
};

export function ChartPager({ crawlerId, refreshKey, chartName, country, onPlayed }: ChartPagerProps) {
  const { t } = useTranslation();
  const [items, setItems] = useState<ChartItem[]>([]);
  const [fetchLoading, setFetchLoading] = useState(false);
  const [fetchError, setFetchError] = useState<string | null>(null);
  const [page, setPage] = useState(0);
  const [pageDir, setPageDir] = useState<PageDir>('initial');
  const [pageKey, setPageKey] = useState(0);
  const { play, close, content, loadingKey, error } = usePlayYouTube({ onPlayed });

  useEffect(() => {
    let alive = true;
    setFetchLoading(true);
    setFetchError(null);
    setPage(0);
    setPageDir('initial');
    setPageKey(0);
    listItems(crawlerId, FETCH_LIMIT)
      .then(result => {
        if (!alive) return;
        setItems(result.filter((it): it is ChartItem => it.kind === 'chart-entry' && it.chart != null));
      })
      .catch(err => {
        if (!alive) return;
        setFetchError(err instanceof Error ? err.message : String(err));
      })
      .finally(() => { if (alive) setFetchLoading(false); });
    return () => { alive = false; };
  }, [crawlerId, refreshKey]);

  if (fetchLoading && items.length === 0) {
    return (
      <div className="flex items-center gap-2 text-sm text-slate-500 px-1">
        <Spinner className="w-4 h-4" />
        {t('items.loading')}
      </div>
    );
  }

  if (fetchError) {
    return (
      <div className="text-sm text-red-700 bg-red-50 border border-red-200 rounded px-3 py-2">
        {fetchError}
      </div>
    );
  }

  if (items.length === 0) {
    return <div className="text-sm text-slate-500 px-1">{t('items.empty')}</div>;
  }

  const totalPages = Math.ceil(items.length / PAGE_SIZE);
  const pageSlice = items.slice(page * PAGE_SIZE, (page + 1) * PAGE_SIZE);

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
    <div className="flex flex-col gap-3 @container">
      <ul key={pageKey} className={`flex flex-col gap-2 ${PAGE_ENTER[pageDir]}`}>
        {pageSlice.map(item => {
          const itemKey = `${item.title}-${(item.artists ?? []).join(',')}`;
          return (
            <li key={item.id} className="border border-slate-200 rounded-xl bg-white overflow-hidden shadow-sm">
              <ChartRow
                item={item}
                palette={paletteForId(crawlerId)}
                onInfo={() => openVideoSearch((item.artists ?? [])[0] ?? '', item.title)}
                onPlay={() => play({ key: itemKey, title: item.title, artists: item.artists ?? [], artworkUrl: item.artworkUrl, chartName, crawlerId, position: item.chart.position, country })}
                playLoading={loadingKey === itemKey}
              />
            </li>
          );
        })}
      </ul>

      {totalPages > 1 && (
        <div className="flex items-center justify-between px-1 pt-1">
          <button
            type="button"
            onClick={goPrev}
            disabled={page === 0}
            aria-label="Previous page"
            className="flex items-center gap-1 px-3 py-1.5 rounded-lg text-sm font-medium text-slate-600 hover:bg-slate-100 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
          >
            <ChevronLeft className="w-4 h-4" />
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
            className="flex items-center gap-1 px-3 py-1.5 rounded-lg text-sm font-medium text-slate-600 hover:bg-slate-100 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
          >
            Next
            <ChevronRight className="w-4 h-4" />
          </button>
        </div>
      )}

      {error && <p className="text-xs text-red-500 px-1 mt-1">{error}</p>}
      {content && <VideoModal {...content} onClose={close} />}
    </div>
  );
}

function ChartRow({ item, palette, onInfo, onPlay, playLoading }: {
  item: ChartItem;
  palette: readonly string[];
  onInfo: () => void;
  onPlay: () => void;
  playLoading: boolean;
}) {
  const { position } = item.chart;
  const bg = position >= 1 && position <= 10 ? palette[position - 1] : 'bg-slate-900';

  return (
    <div className="flex items-center gap-2 @[300px]:gap-3 @[420px]:gap-4 px-3 @[420px]:px-4 py-3 @[420px]:py-4 w-full text-left group">
      <FlipNumber
        value={position}
        className={`shrink-0 w-14 h-14 @[300px]:w-16 @[300px]:h-16 @[420px]:w-24 @[420px]:h-24 rounded-lg @[420px]:rounded-xl ${bg} flex items-center justify-center gap-0.5 @[420px]:gap-1 text-2xl @[300px]:text-3xl @[420px]:text-4xl text-white shadow-inner`}
      />

      {item.artworkUrl ? (
        <img
          src={item.artworkUrl}
          alt=""
          className="w-12 h-12 @[300px]:w-16 @[300px]:h-16 @[420px]:w-20 @[420px]:h-20 rounded-lg object-cover shrink-0 shadow-sm transition-transform duration-300 group-hover:scale-110"
        />
      ) : (
        <div className="w-12 h-12 @[300px]:w-16 @[300px]:h-16 @[420px]:w-20 @[420px]:h-20 rounded-lg bg-gradient-to-br from-violet-100 to-indigo-100 shrink-0 transition-transform duration-300 group-hover:scale-110" />
      )}

      <div className="flex-1 min-w-0">
        <p className="font-semibold text-slate-900 truncate text-sm @[300px]:text-base @[420px]:text-lg leading-tight">
          {item.title}
        </p>
        {item.artists && item.artists.length > 0 && (
          <p className="text-xs @[300px]:text-sm text-slate-500 truncate mt-1">
            {item.artists.join(' · ')}
          </p>
        )}
      </div>

      <div className="flex items-center gap-1 shrink-0 ml-auto pl-2">
        <button type="button" onClick={onInfo} aria-label={`Search ${item.title} on Google`} className="p-1.5 rounded-md hover:bg-slate-100 text-slate-500 hover:text-slate-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors">
          <Info className="w-4 h-4" />
        </button>
        <button type="button" onClick={onPlay} disabled={playLoading} aria-label={`Play ${item.title} on YouTube`} className="p-1.5 rounded-md hover:bg-slate-100 text-slate-500 hover:text-slate-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 disabled:opacity-40 transition-colors">
          {playLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4" />}
        </button>
      </div>
    </div>
  );
}
