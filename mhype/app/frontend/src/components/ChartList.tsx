import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { listItems, type ChartContext, type Item } from '../api';
import { Spinner } from './Spinner';
import { openVideoSearch } from '../lib/openVideoSearch';
import { usePlayYouTube } from '../lib/usePlayYouTube';
import { VideoModal } from './VideoModal';
import { SongListView, type SongListEntry } from './SongListView';

interface ChartListProps {
  crawlerId: string;
  refreshKey: number;
  chartName?: string;
  country?: string;
  onPlayed?: () => void;
}

type ChartItem = Item & { chart: ChartContext };

const FETCH_LIMIT = 100;

export function ChartList({ crawlerId, refreshKey, chartName, country, onPlayed }: ChartListProps) {
  const { t } = useTranslation();
  const [items, setItems] = useState<ChartItem[]>([]);
  const [fetchLoading, setFetchLoading] = useState(false);
  const [fetchError, setFetchError] = useState<string | null>(null);
  const { play, close, content, loadingKey, error } = usePlayYouTube({ onPlayed });

  useEffect(() => {
    let alive = true;
    setFetchLoading(true);
    setFetchError(null);
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

  const entries: SongListEntry[] = items.map(item => {
    const id = `${item.title}-${(item.artists ?? []).join(',')}`;
    return {
      id,
      position: item.chart.position,
      title: item.title,
      artists: item.artists ?? [],
      artworkUrl: item.artworkUrl,
      chartName,
      country,
      crawlerId,
    };
  });

  return (
    <div className="flex flex-col gap-2">
      <SongListView
        key={`${crawlerId}-${refreshKey}`}
        entries={entries}
        paletteId={crawlerId}
        loadingId={loadingKey}
        onPlay={entry => play({
          key: entry.id,
          title: entry.title,
          artists: entry.artists,
          artworkUrl: entry.artworkUrl,
          chartName: entry.chartName,
          crawlerId: entry.crawlerId,
          position: entry.position,
          country: entry.country,
        })}
        onSearch={openVideoSearch}
      />
      {error && <p className="text-xs text-red-500 px-1 mt-1">{error}</p>}
      {content && <VideoModal {...content} onClose={close} />}
    </div>
  );
}
