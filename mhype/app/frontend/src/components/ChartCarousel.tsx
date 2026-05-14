import { useState } from 'react';
import { ChevronLeft, ChevronRight, Info, Loader2, Play } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Spinner } from './Spinner';
import { openVideoSearch } from '../lib/openVideoSearch';
import { usePlayYouTube } from '../lib/usePlayYouTube';
import { VideoModal } from './VideoModal';

const VISIBLE = 6;
const STEP = 4;

const TOP_10_COLORS: Readonly<Record<number, string>> = {
  1:  'bg-amber-500',
  2:  'bg-slate-500',
  3:  'bg-orange-600',
  4:  'bg-red-500',
  5:  'bg-pink-500',
  6:  'bg-purple-600',
  7:  'bg-blue-500',
  8:  'bg-sky-500',
  9:  'bg-emerald-500',
  10: 'bg-teal-500',
};

function positionBadgeColor(position: number): string {
  return TOP_10_COLORS[position] ?? 'bg-black';
}

export interface ChartCarouselItem {
  id: string;
  position: number;
  title: string;
  artists: string[];
  artworkUrl?: string;
}

interface ChartCarouselProps {
  items: ChartCarouselItem[];
  loading?: boolean;
  error?: string | null;
  chartName?: string;
  country?: string;
  crawlerId?: string;
  onPlayed?: () => void;
}

interface CarouselCardProps {
  item: ChartCarouselItem;
  onInfo: () => void;
  onPlay: () => void;
  playLoading: boolean;
}

export function ChartCarousel({ items, loading = false, error = null, chartName, country, crawlerId, onPlayed }: ChartCarouselProps) {
  const { t } = useTranslation();
  const [start, setStart] = useState(0);
  const [animKey, setAnimKey] = useState(0);
  const { play, close, content, loadingKey, error: playError } = usePlayYouTube({ onPlayed });

  if (loading && items.length === 0) {
    return (
      <div className="flex items-center gap-2 text-sm text-slate-500">
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
    return <div className="text-sm text-slate-500">{t('items.empty')}</div>;
  }

  const canPrev = start > 0;
  const canNext = start + VISIBLE < items.length;
  const visible = items.slice(start, start + VISIBLE);

  function prev() {
    setStart(s => Math.max(0, s - STEP));
    setAnimKey(k => k + 1);
  }

  function next() {
    setStart(s => s + STEP);
    setAnimKey(k => k + 1);
  }

  return (
    <div className="flex flex-col gap-1">
      <div className="relative group">
        <button
          type="button"
          onClick={prev}
          disabled={!canPrev}
          aria-label="Previous"
          className="absolute left-0 inset-y-0 z-10 w-12 flex items-center justify-center bg-white/90 text-black opacity-0 group-hover:opacity-100 disabled:opacity-0 disabled:pointer-events-none transition-opacity duration-200"
        >
          <ChevronLeft className="w-7 h-7 drop-shadow" />
        </button>

        <div className="grid grid-cols-6 gap-3">
          {visible.map((item, index) => {
            const itemKey = `${item.title}-${item.artists.join(',')}`;
            return (
              <div
                key={`${item.id}-${animKey}`}
                className="animate-in fade-in slide-in-from-bottom-6 duration-700 fill-mode-backwards"
                style={{ animationDelay: `${index * 270}ms` }}
              >
                <CarouselCard
                  item={item}
                  onInfo={() => openVideoSearch(item.artists[0] ?? '', item.title)}
                  onPlay={() => play({ key: itemKey, title: item.title, artists: item.artists, artworkUrl: item.artworkUrl, position: item.position, chartName, crawlerId, country })}
                  playLoading={loadingKey === itemKey}
                />
              </div>
            );
          })}
        </div>

        <button
          type="button"
          onClick={next}
          disabled={!canNext}
          aria-label="Next"
          className="absolute right-0 inset-y-0 z-10 w-12 flex items-center justify-center bg-white/90 text-black opacity-0 group-hover:opacity-100 disabled:opacity-0 disabled:pointer-events-none transition-opacity duration-200"
        >
          <ChevronRight className="w-7 h-7 drop-shadow" />
        </button>
      </div>

      {playError && <p className="text-xs text-red-500 px-1 mt-1">{playError}</p>}
      {content && <VideoModal {...content} onClose={close} />}
    </div>
  );
}

function CarouselCard({ item, onInfo, onPlay, playLoading }: CarouselCardProps) {
  return (
    <div className="flex flex-col gap-1 min-w-0 w-full group/card">
      <div className="relative aspect-square overflow-hidden rounded-lg bg-gradient-to-br from-slate-200 to-slate-300">
        {item.artworkUrl && (
          <img
            src={item.artworkUrl}
            alt=""
            className="absolute inset-0 w-full h-full object-cover"
          />
        )}
        <span className={`absolute top-2 left-2 ${positionBadgeColor(item.position)} text-white text-xl font-black px-3 py-2 rounded-lg leading-none`}>
          {item.position}
        </span>
        <div className="absolute bottom-2 right-2 flex gap-1 opacity-0 group-hover/card:opacity-100 transition-opacity">
          <button
            type="button"
            onClick={onInfo}
            aria-label={`Search ${item.title} on Google`}
            className="p-1.5 rounded-full bg-white/90 hover:bg-white text-slate-900 shadow focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors"
          >
            <Info className="w-3.5 h-3.5" />
          </button>
          <button
            type="button"
            onClick={onPlay}
            disabled={playLoading}
            aria-label={`Play ${item.title} on YouTube`}
            className="p-1.5 rounded-full bg-white/90 hover:bg-white text-slate-900 shadow focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 disabled:opacity-40 transition-colors"
          >
            {playLoading ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Play className="w-3.5 h-3.5" />}
          </button>
        </div>
      </div>
      <p className="text-xs font-semibold leading-tight truncate text-slate-900">
        {item.title}
      </p>
      {item.artists.length > 0 && (
        <p className="text-xs text-slate-500 truncate -mt-0.5">
          {item.artists[0]}
        </p>
      )}
    </div>
  );
}
