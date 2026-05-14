import { useState } from 'react';
import { Loader2, Play, Search, X } from 'lucide-react';
import { type CrawlerInfo } from '../api';
import { usePlayYouTube } from '../lib/usePlayYouTube';
import { useSearch, type SearchResult } from '../lib/useSearch';
import { Spinner } from './Spinner';
import { VideoModal } from './VideoModal';

interface SearchPanelProps {
  activeCrawlers: CrawlerInfo[];
  onPlayed?: () => void;
}

interface CardProps {
  result: SearchResult;
  isLoading: boolean;
  onPlay: () => void;
}

function SearchResultCard({ result, isLoading, onPlay }: CardProps) {
  return (
    <button
      type="button"
      onClick={onPlay}
      disabled={isLoading}
      className="flex items-center gap-3 p-3 bg-white rounded-xl shadow-sm border border-slate-100 hover:border-blue-200 hover:shadow-md transition-all text-left w-full disabled:opacity-60"
    >
      <div className="shrink-0 w-12 h-12 rounded-lg bg-slate-100 overflow-hidden flex items-center justify-center">
        {result.artworkUrl
          ? <img src={result.artworkUrl} alt="" className="w-full h-full object-cover" />
          : <Play className="w-5 h-5 text-slate-300" />}
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-sm font-semibold text-slate-900 truncate leading-tight">{result.title}</p>
        {result.artists.length > 0 && (
          <p className="text-xs text-slate-500 truncate mt-0.5">{result.artists[0]}</p>
        )}
        <p className="text-xs text-slate-400 mt-0.5 truncate">
          {result.chartName}{result.country ? ` · ${result.country}` : ''} · #{result.position}
        </p>
      </div>
      <div className="shrink-0 text-slate-400">
        {isLoading
          ? <Loader2 className="w-5 h-5 animate-spin" />
          : <Play className="w-5 h-5" />}
      </div>
    </button>
  );
}

export function SearchPanel({ activeCrawlers, onPlayed }: SearchPanelProps) {
  const [query, setQuery] = useState('');
  const [scopedId, setScopedId] = useState<string | null>(null);

  const scopeIds = scopedId ? [scopedId] : activeCrawlers.map(c => c.id);
  const { results, loading } = useSearch(query, scopeIds, activeCrawlers);
  const { play, close, content, loadingKey } = usePlayYouTube({ onPlayed });

  if (activeCrawlers.length === 0) return null;

  const hasQuery = query.trim().length >= 2;

  return (
    <div className="flex flex-col gap-3">
      <div className="flex gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
          <input
            type="search"
            value={query}
            onChange={e => setQuery(e.target.value)}
            placeholder="Search titles or artists…"
            className="w-full pl-9 pr-9 py-2 text-sm bg-white border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
          {query && (
            <button
              type="button"
              onClick={() => setQuery('')}
              aria-label="Clear search"
              className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600"
            >
              <X className="w-4 h-4" />
            </button>
          )}
        </div>
        {activeCrawlers.length > 1 && (
          <select
            value={scopedId ?? ''}
            onChange={e => setScopedId(e.target.value || null)}
            aria-label="Scope search to chart"
            className="text-sm bg-white border border-slate-200 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 text-slate-700 cursor-pointer"
          >
            <option value="">All charts</option>
            {activeCrawlers.map(c => (
              <option key={c.id} value={c.id}>{c.displayName}</option>
            ))}
          </select>
        )}
      </div>

      {hasQuery && (
        <div className="flex flex-col gap-3">
          {loading && (
            <div className="flex items-center gap-2 text-sm text-slate-500 px-1">
              <Spinner className="w-4 h-4" />
              Searching…
            </div>
          )}
          {!loading && results.length === 0 && (
            <p className="text-sm text-slate-500 px-1">No results for &ldquo;{query.trim()}&rdquo;</p>
          )}
          {results.length > 0 && (
            <div className="grid grid-cols-2 gap-2">
              {results.map(r => (
                <SearchResultCard
                  key={r.id}
                  result={r}
                  isLoading={loadingKey === r.id}
                  onPlay={() => play({
                    key: r.id,
                    title: r.title,
                    artists: r.artists,
                    artworkUrl: r.artworkUrl,
                    chartName: r.chartName,
                    crawlerId: r.crawlerId,
                    position: r.position,
                    country: r.country,
                  })}
                />
              ))}
            </div>
          )}
        </div>
      )}

      {content && <VideoModal {...content} onClose={close} />}
    </div>
  );
}
