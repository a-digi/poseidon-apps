import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { listItems, type Item, type ItemKind } from '../api';
import { relativeTime } from '../lib/relativeTime';
import { Spinner } from './Spinner';

interface ItemListProps {
  crawlerId: string;
  refreshKey: number;
}

const INITIAL_VISIBLE = 20;
const FETCH_LIMIT = 50;

export function ItemList({ crawlerId, refreshKey }: ItemListProps) {
  const { t } = useTranslation();
  const [items, setItems] = useState<Item[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showAll, setShowAll] = useState(false);

  useEffect(() => {
    let alive = true;
    setLoading(true);
    setError(null);
    listItems(crawlerId, FETCH_LIMIT)
      .then(result => {
        if (!alive) return;
        setItems(result);
        setShowAll(false);
      })
      .catch(err => {
        if (!alive) return;
        setError(err instanceof Error ? err.message : String(err));
      })
      .finally(() => {
        if (alive) setLoading(false);
      });
    return () => { alive = false; };
  }, [crawlerId, refreshKey]);

  if (loading && items.length === 0) {
    return (
      <div className="flex items-center gap-2 text-sm text-slate-500 px-1">
        <Spinner className="w-4 h-4" />
        {t('items.loading')}
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-sm text-red-700 bg-red-50 border border-red-200 rounded px-3 py-2">
        {error}
      </div>
    );
  }

  if (items.length === 0) {
    return <div className="text-sm text-slate-500 px-1">{t('items.empty')}</div>;
  }

  const visible = showAll ? items : items.slice(0, INITIAL_VISIBLE);

  return (
    <div className="flex flex-col gap-1">
      <ul className="flex flex-col gap-2">
        {visible.map((item, index) => (
          <li
            key={item.id}
            className="border border-slate-200 rounded-lg bg-white overflow-hidden animate-in fade-in slide-in-from-bottom-2 duration-300 fill-mode-both"
            style={{ animationDelay: `${Math.min(index, 8) * 60}ms` }}
          >
            <ItemRow item={item} />
          </li>
        ))}
      </ul>
      {!showAll && items.length > INITIAL_VISIBLE && (
        <button
          type="button"
          onClick={() => setShowAll(true)}
          className="self-start text-sm text-blue-600 hover:text-blue-700 hover:underline px-1 py-1"
        >
          {t('items.showMore')}
        </button>
      )}
    </div>
  );
}

// ─── Chart entry row ──────────────────────────────────────────────────────────

type Movement = 'new' | 'up' | 'down' | 'same';

function MovementBadge({ movement, diff }: { movement: Movement; diff: number }) {
  if (movement === 'new') {
    return (
      <span className="text-[9px] font-bold uppercase tracking-widest text-emerald-500 leading-none">
        NEW
      </span>
    );
  }
  if (movement === 'same' || diff === 0) {
    return <span className="text-xs text-slate-300 leading-none">—</span>;
  }
  if (movement === 'up') {
    return <span className="text-[10px] font-bold text-emerald-500 leading-none">↑{diff}</span>;
  }
  return <span className="text-[10px] font-bold text-rose-400 leading-none">↓{diff}</span>;
}

function ChartEntryRow({ item }: { item: Item }) {
  const chart = item.chart!;
  const pos = chart.position;
  const prev = chart.prevPosition ?? 0;
  const movement: Movement =
    prev === 0 ? 'new' : prev > pos ? 'up' : prev < pos ? 'down' : 'same';
  const diff = prev === 0 ? 0 : Math.abs(prev - pos);
  const posLabel = String(pos).padStart(2, '0');

  return (
    <div className="flex items-center gap-3 px-3 py-2.5 hover:bg-slate-50 transition-colors group">
      {/* Position — clipped-gradient number */}
      <div className="shrink-0 w-12 flex flex-col items-center gap-0.5">
        <span
          className="text-3xl font-black tabular-nums leading-none bg-gradient-to-br from-violet-600 via-purple-500 to-indigo-500 bg-clip-text text-transparent select-none"
        >
          {posLabel}
        </span>
        <MovementBadge movement={movement} diff={diff} />
      </div>

      {/* Artwork */}
      {item.artworkUrl ? (
        <img
          src={item.artworkUrl}
          alt=""
          className="w-10 h-10 rounded-md object-cover shrink-0 shadow-sm transition-transform duration-300 group-hover:scale-110"
        />
      ) : (
        <div className="w-10 h-10 rounded-md bg-gradient-to-br from-violet-100 to-indigo-100 shrink-0 transition-transform duration-300 group-hover:scale-110" />
      )}

      {/* Content */}
      <div className="flex-1 min-w-0">
        <p className="font-semibold text-slate-900 truncate text-sm leading-tight">
          {item.title}
        </p>
        {item.artists && item.artists.length > 0 && (
          <p className="text-xs text-slate-500 truncate mt-0.5">
            {item.artists.join(' · ')}
          </p>
        )}
      </div>

      {/* Meta column */}
      <div className="shrink-0 flex flex-col items-end gap-0.5 min-w-[3rem] text-right">
        {chart.peakPosition != null && chart.peakPosition > 0 && (
          <span className="text-[10px] font-bold text-amber-500 leading-none" title="Peak position">
            ★{chart.peakPosition}
          </span>
        )}
        {chart.weeksOnChart != null && chart.weeksOnChart > 0 && (
          <span className="text-[10px] text-slate-400 leading-none">
            {chart.weeksOnChart}w
          </span>
        )}
      </div>
    </div>
  );
}

// ─── Generic row (news, track, release) ──────────────────────────────────────

const KIND_BADGE_CLASS: Record<ItemKind, string> = {
  news: 'bg-slate-100 text-slate-700',
  track: 'bg-blue-100 text-blue-700',
  'chart-entry': 'bg-violet-100 text-violet-700',
  release: 'bg-emerald-100 text-emerald-700',
};

const PILL_CLASS = 'text-[10px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded';

function formatDuration(durationSec: number): string {
  return `${Math.floor(durationSec / 60)}:${(durationSec % 60).toString().padStart(2, '0')}`;
}

function summaryFor(item: Item): string {
  switch (item.kind) {
    case 'chart-entry':
      return '';
    case 'release':
      return item.releasedAt ?? '';
    case 'news': {
      const publishedMs = item.news?.publishedAt ? Date.parse(item.news.publishedAt) : NaN;
      return Number.isFinite(publishedMs) ? relativeTime(publishedMs) : relativeTime(item.scrapedAt);
    }
    case 'track':
      return item.durationSec ? formatDuration(item.durationSec) : '';
  }
}

function GenericRow({ item }: { item: Item }) {
  const summary = summaryFor(item);
  return (
    <div className="flex items-baseline gap-2 flex-wrap px-3 py-2 hover:bg-slate-50 transition-colors">
      <span className={`${PILL_CLASS} ${KIND_BADGE_CLASS[item.kind]} shrink-0`}>{item.kind}</span>
      <span className="font-medium text-slate-900 truncate">{item.title}</span>
      {item.artists && item.artists.length > 0 && (
        <span className="text-sm text-slate-600 truncate">{item.artists.join(' · ')}</span>
      )}
      {item.album && <span className="text-sm text-slate-400 truncate">{item.album}</span>}
      {summary && <span className="text-xs text-slate-400 ml-auto shrink-0">{summary}</span>}
    </div>
  );
}

function ItemRow({ item }: { item: Item }) {
  if (item.kind === 'chart-entry' && item.chart != null) {
    return <ChartEntryRow item={item} />;
  }
  return <GenericRow item={item} />;
}
