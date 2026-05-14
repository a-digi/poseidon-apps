import { RefreshCw } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import type { CrawlerInfo } from '../api';
import { relativeTime } from '../lib/relativeTime';
import { Spinner } from './Spinner';
import { Switch } from './Switch';

interface CrawlerCardProps {
  crawler: CrawlerInfo;
  onTrigger: () => void;
  onToggleActive: (next: boolean) => void;
  triggering: boolean;
  error?: string;
}

export function CrawlerCard({ crawler, onTrigger, onToggleActive, triggering, error }: CrawlerCardProps) {
  const { t } = useTranslation();
  const intervalMinutes = Math.max(1, Math.round(crawler.intervalSec / 60));
  const lastData = crawler.lastSuccessAt > 0 ? relativeTime(crawler.lastSuccessAt) : t('crawler.never');

  return (
    <div className="border border-slate-200 rounded-lg p-4 bg-white flex items-center justify-between gap-4">
      <div className="min-w-0 flex-1">
        <div className="flex items-baseline gap-2 flex-wrap">
          <span className="font-semibold text-slate-900">{crawler.displayName}</span>
          <span className="text-xs text-slate-500">{crawler.source} · {crawler.country}</span>
        </div>
        <div className="text-xs text-slate-500 mt-1">
          {t('crawler.everyN', { n: intervalMinutes })}
          {' · '}
          {t('crawler.fileCount', { count: crawler.fileCount })}
          {' · '}
          {t('crawler.lastData', { ago: lastData })}
        </div>
        {error && (
          <div className="text-xs text-red-700 mt-1">{error}</div>
        )}
      </div>
      <div className="flex items-center gap-2">
        <span className="text-xs text-slate-500">
          {crawler.active ? t('crawler.active') : t('crawler.inactive')}
        </span>
        <Switch
          checked={crawler.active}
          onChange={onToggleActive}
          ariaLabel={`${crawler.displayName}: ${crawler.active ? t('crawler.inactive') : t('crawler.active')}`}
        />
        <div className="relative group">
          <button
            type="button"
            onClick={onTrigger}
            disabled={triggering}
            aria-label={t('crawler.runNow')}
            className="flex items-center justify-center w-8 h-8 rounded-md text-slate-400 hover:text-slate-700 hover:bg-slate-100 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
          >
            {triggering ? <Spinner className="w-4 h-4" /> : <RefreshCw className="w-4 h-4" />}
          </button>
          <span className="pointer-events-none absolute right-0 top-full mt-1.5 whitespace-nowrap rounded bg-slate-800 px-2 py-1 text-xs text-white opacity-0 transition-opacity group-hover:opacity-100 z-10">
            {t('crawler.runNow')}
          </span>
        </div>
      </div>
    </div>
  );
}
