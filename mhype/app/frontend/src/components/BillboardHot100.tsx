import { useEffect, useState } from 'react';
import { listItems, type Item, type ChartContext } from '../api';
import { ChartCarousel, type ChartCarouselItem } from './ChartCarousel';

const CRAWLER_ID = 'billboard-hot-100';

interface BillboardHot100Props {
  refreshKey?: number;
  onPlayed?: () => void;
}

type ChartItem = Item & { chart: ChartContext };

export function BillboardHot100({ refreshKey = 0, onPlayed }: BillboardHot100Props) {
  const [items, setItems] = useState<ChartCarouselItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let alive = true;
    setLoading(true);
    setError(null);
    listItems(CRAWLER_ID, 100)
      .then(raw => {
        if (!alive) return;
        setItems(
          raw
            .filter((it): it is ChartItem => it.kind === 'chart-entry' && it.chart != null)
            .sort((a, b) => a.chart.position - b.chart.position)
            .map(it => ({
              id: it.id,
              position: it.chart.position,
              title: it.title,
              artists: it.artists ?? [],
              artworkUrl: it.artworkUrl,
            })),
        );
      })
      .catch(err => {
        if (!alive) return;
        setError(err instanceof Error ? err.message : String(err));
      })
      .finally(() => {
        if (alive) setLoading(false);
      });
    return () => {
      alive = false;
    };
  }, [refreshKey]);

  return <ChartCarousel items={items} loading={loading} error={error} chartName="Billboard Hot 100" country="US" crawlerId={CRAWLER_ID} onPlayed={onPlayed} />;
}
