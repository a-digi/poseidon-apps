import { useEffect, useRef, useState } from 'react';
import { listItems, type ChartContext, type CrawlerInfo, type Item } from '../api';

export interface SearchResult {
  id: string;
  title: string;
  artists: string[];
  artworkUrl?: string;
  position: number;
  chartName: string;
  country: string;
  crawlerId: string;
}

type ChartEntry = Item & { chart: ChartContext };

export function useSearch(query: string, scopeIds: string[], crawlers: CrawlerInfo[]) {
  const [items, setItems] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const crawlersRef = useRef(crawlers);
  crawlersRef.current = crawlers;

  const scopeKey = [...scopeIds].sort().join(',');
  const hasQuery = query.trim().length >= 2;

  useEffect(() => {
    if (!hasQuery || !scopeKey) return;
    const ids = scopeKey.split(',');
    let alive = true;
    setLoading(true);
    setItems([]);
    Promise.all(ids.map(id => listItems(id, 100)))
      .then(results => {
        if (!alive) return;
        const crawlerMap = new Map(crawlersRef.current.map(c => [c.id, c]));
        const merged: SearchResult[] = results.flatMap((raw, i) => {
          const cid = ids[i];
          const crawler = crawlerMap.get(cid);
          return raw
            .filter((it): it is ChartEntry => it.kind === 'chart-entry' && it.chart != null)
            .map(it => ({
              id: `${cid}::${it.title}::${(it.artists ?? []).join(',')}`,
              title: it.title,
              artists: it.artists ?? [],
              artworkUrl: it.artworkUrl,
              position: it.chart.position,
              chartName: crawler?.displayName ?? cid,
              country: crawler?.country ?? '',
              crawlerId: cid,
            }));
        });
        setItems(merged);
      })
      .catch(() => undefined)
      .finally(() => { if (alive) setLoading(false); });
    return () => { alive = false; };
  // scopeIds encoded in scopeKey; crawlers accessed via ref to avoid re-fetching on reference churn
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [scopeKey, hasQuery]);

  const q = query.trim().toLowerCase();
  const results = hasQuery
    ? items.filter(item =>
        item.title.toLowerCase().includes(q) ||
        item.artists.some(a => a.toLowerCase().includes(q))
      )
    : [];

  return { results, loading: loading && hasQuery };
}
