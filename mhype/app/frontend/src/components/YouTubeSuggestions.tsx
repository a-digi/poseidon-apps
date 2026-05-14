import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { getYouTubeSuggestions, type YouTubeResult } from '../api';
import { Spinner } from './Spinner';

interface YouTubeSuggestionsProps {
  artist: string;
  title: string;
}

export function YouTubeSuggestions({ artist, title }: YouTubeSuggestionsProps) {
  const { t } = useTranslation();
  const [results, setResults] = useState<YouTubeResult[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState(0);

  useEffect(() => {
    let alive = true;
    setLoading(true);
    setError(null);
    setResults(null);
    setActiveTab(0);
    getYouTubeSuggestions(artist, title)
      .then((res) => { if (alive) setResults(res.results); })
      .catch((err) => { if (alive) setError(err instanceof Error ? err.message : String(err)); })
      .finally(() => { if (alive) setLoading(false); });
    return () => { alive = false; };
  }, [artist, title]);

  if (loading) return <div className="flex items-center gap-2 text-sm text-slate-500 px-3 py-4"><Spinner className="w-4 h-4" />{t('items.loading')}</div>;
  if (error)   return <div className="text-sm text-red-700 bg-red-50 border border-red-200 rounded px-3 py-2 m-3">{error}</div>;
  if (!results || results.length === 0) {
    return <div className="text-sm text-slate-500 italic px-3 py-4">{t('youtube.notReady')}</div>;
  }

  const active = results[activeTab];

  return (
    <div className="flex flex-col">
      <div className="flex border-b border-slate-200 bg-slate-50">
        {results.map((r, i) => (
          <button
            key={r.videoId}
            type="button"
            onClick={() => setActiveTab(i)}
            className={`flex-1 px-3 py-2 text-xs font-medium transition-colors ${
              i === activeTab ? 'bg-white text-slate-900 border-b-2 border-violet-500' : 'text-slate-500 hover:text-slate-700'
            }`}
          >
            {t('youtube.tabLabel', { n: i + 1 })}
          </button>
        ))}
      </div>
      <div className="aspect-video w-full bg-black">
        <iframe
          key={active.videoId}
          src={`https://www.youtube.com/embed/${active.videoId}`}
          title={active.title}
          allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
          allowFullScreen
          className="w-full h-full border-0"
        />
      </div>
      <div className="px-3 py-2 text-xs text-slate-500 truncate" title={active.title}>
        {active.title} · <span className="text-slate-400">{active.channelTitle}</span>
      </div>
    </div>
  );
}
