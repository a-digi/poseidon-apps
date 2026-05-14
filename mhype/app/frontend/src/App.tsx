import { useCallback, useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  orchestratorStatus,
  ensureOrchestrator,
  listCrawlers,
  restartOrchestrator,
  stopOrchestrator,
  triggerCrawl,
  setCrawlerActive,
  type OrchestratorStatus as OrchestratorStatusType,
  type CrawlerInfo,
} from './api';
import { OrchestratorStatus } from './components/OrchestratorStatus';
import { CrawlerCard } from './components/CrawlerCard';
import { InfoAlert } from './components/InfoAlert';
import { ChartPager } from './components/ChartPager';
import { ChartList } from './components/ChartList';
import { BillboardHot100 } from './components/BillboardHot100';
import { OverflowMenu } from './components/OverflowMenu';
import { Highlights } from './components/Highlights';
import { SearchPanel } from './components/SearchPanel';
import { relativeTime } from './lib/relativeTime';
import { useHighlights } from './lib/useHighlights';
import { EyeOff, Settings, X } from 'lucide-react';
import './index.css';

type BusySlot = 'start' | 'stop' | 'restart' | null;

function App() {
  const { t } = useTranslation();
  const highlights = useHighlights();
  const [status, setStatus] = useState<OrchestratorStatusType | null>(null);
  const [crawlers, setCrawlers] = useState<CrawlerInfo[]>([]);
  const [busy, setBusy] = useState<BusySlot>(null);
  const [triggering, setTriggering] = useState<Record<string, boolean>>({});
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [refreshKeys, setRefreshKeys] = useState<Record<string, number>>({});
  const [triggerErrors, setTriggerErrors] = useState<Record<string, string>>({});
  const errorTimers = useRef<Record<string, number>>({});
  const autoTriggeredRef = useRef<Set<string>>(new Set());

  const handleTrigger = useCallback(async (crawlerId: string) => {
    const existing = errorTimers.current[crawlerId];
    if (existing) {
      clearTimeout(existing);
      delete errorTimers.current[crawlerId];
    }
    setTriggerErrors(prev => {
      if (!(crawlerId in prev)) return prev;
      const { [crawlerId]: _, ...rest } = prev;
      return rest;
    });
    setTriggering(prev => ({ ...prev, [crawlerId]: true }));
    try {
      await triggerCrawl(crawlerId);
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      console.error(`trigger_crawl ${crawlerId} failed`, err);
      setTriggerErrors(prev => ({ ...prev, [crawlerId]: msg }));
      errorTimers.current[crawlerId] = window.setTimeout(() => {
        setTriggerErrors(prev => {
          const { [crawlerId]: _, ...rest } = prev;
          return rest;
        });
        delete errorTimers.current[crawlerId];
      }, 5000);
    } finally {
      setTriggering(prev => ({ ...prev, [crawlerId]: false }));
      setRefreshKeys(prev => ({ ...prev, [crawlerId]: (prev[crawlerId] ?? 0) + 1 }));
    }
  }, []);

  const handleSetActive = useCallback(async (crawlerId: string, active: boolean) => {
    setCrawlers(prev => prev.map(c => c.id === crawlerId ? { ...c, active } : c));
    try { await setCrawlerActive(crawlerId, active); }
    catch (err) {
      console.error(`set_crawler_active ${crawlerId} failed`, err);
      setCrawlers(prev => prev.map(c => c.id === crawlerId ? { ...c, active: !active } : c));
    }
  }, []);

  useEffect(() => {
    let alive = true;
    let intervalId: number;

    const tick = async () => {
      try {
        const [s, c] = await Promise.all([orchestratorStatus(), listCrawlers()]);
        if (!alive) return;
        setStatus(s);
        setCrawlers(c);
      } catch {
        // keep last good state
      }
    };

    const init = async () => {
      await ensureOrchestrator().catch(() => undefined);
      try {
        const [s, c] = await Promise.all([orchestratorStatus(), listCrawlers()]);
        if (!alive) return;
        setStatus(s);
        setCrawlers(c);
        for (const crawler of c) {
          if (crawler.lastSuccessAt === 0 && !autoTriggeredRef.current.has(crawler.id)) {
            autoTriggeredRef.current.add(crawler.id);
            void handleTrigger(crawler.id);
          }
        }
      } catch {
        // keep last good state
      }
      intervalId = window.setInterval(tick, 5000);
    };

    void init();

    return () => {
      alive = false;
      clearInterval(intervalId);
    };
  }, [handleTrigger]);

  const refreshStatus = useCallback(async () => {
    try {
      const s = await orchestratorStatus();
      setStatus(s);
    } catch {
      // ignore
    }
  }, []);

  const runOrchestratorAction = useCallback(
    async (slot: Exclude<BusySlot, null>, fn: () => Promise<unknown>) => {
      setBusy(slot);
      try {
        await fn();
        await refreshStatus();
      } catch (err) {
        console.error(`orchestrator ${slot} failed`, err);
      } finally {
        setBusy(null);
      }
    },
    [refreshStatus],
  );

  const handleStart = useCallback(() => runOrchestratorAction('start', ensureOrchestrator), [runOrchestratorAction]);
  const handleStop = useCallback(() => runOrchestratorAction('stop', stopOrchestrator), [runOrchestratorAction]);
  const handleRestart = useCallback(
    () => runOrchestratorAction('restart', restartOrchestrator),
    [runOrchestratorAction],
  );

  useEffect(() => () => {
    Object.values(errorTimers.current).forEach(id => clearTimeout(id));
    errorTimers.current = {};
  }, []);

  return (
    <div className="p-6 flex flex-col gap-6 text-slate-900">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">{t('app.title')}</h1>
        <button
          onClick={() => setSettingsOpen(prev => !prev)}
          className="flex items-center gap-1.5 text-sm font-medium text-slate-500 hover:text-slate-900 transition-colors"
          aria-expanded={settingsOpen}
          aria-controls="settings-panel"
        >
          {settingsOpen ? <X className="w-4 h-4" /> : <Settings className="w-4 h-4" />}
          {t('settings.title')}
        </button>
      </div>
      <SearchPanel activeCrawlers={crawlers.filter(c => c.active)} onPlayed={highlights.refetch} />
      {settingsOpen && (
        <div id="settings-panel" className="bg-gray-50/80 backdrop-blur-sm px-4 pb-4 pt-3 flex flex-col gap-3 text-gray-900">
          <InfoAlert message={t('settings.infoText')} />
          <OrchestratorStatus
            status={status}
            onStart={handleStart}
            onStop={handleStop}
            onRestart={handleRestart}
            busy={busy}
          />
          {crawlers.length > 0 && (
            <details className="border border-slate-200 rounded-lg">
              <summary className="cursor-pointer select-none px-3 py-2 text-sm font-medium text-slate-600 flex items-center gap-2">
                {t('settings.crawlers', 'Platforms')}
                <span className="text-xs bg-slate-200 text-slate-600 px-1.5 py-0.5 rounded-full tabular-nums">
                  {crawlers.length}
                </span>
              </summary>
              <div className="px-3 pb-3 pt-1 grid grid-cols-3 md:grid-cols-2 lg:grid-cols-3 gap-2">
                {crawlers.map(c => (
                  <CrawlerCard
                    key={c.id}
                    crawler={c}
                    onTrigger={() => handleTrigger(c.id)}
                    onToggleActive={(next) => handleSetActive(c.id, next)}
                    triggering={!!triggering[c.id]}
                    error={triggerErrors[c.id]}
                  />
                ))}
              </div>
            </details>
          )}
        </div>
      )}
      {crawlers.some(c => c.id === 'billboard-hot-100' && c.active) && (
        <section className="flex flex-col gap-2">
          <div className="flex items-center gap-2 flex-wrap px-1">
            <span className="font-semibold text-slate-900">Billboard Hot 100</span>
            <span className="text-xs text-slate-500">US</span>
            <span className="text-xs text-slate-500 ml-auto">
              {t('crawler.lastData', {
                ago: (() => {
                  const c = crawlers.find(c => c.id === 'billboard-hot-100');
                  return c && c.lastSuccessAt > 0 ? relativeTime(c.lastSuccessAt) : t('crawler.never');
                })(),
              })}
            </span>
            <OverflowMenu
              ariaLabel={t('overflow.menu')}
              items={[{
                label: t('menu.deactivate'),
                icon: <EyeOff className="w-4 h-4" />,
                onClick: () => handleSetActive('billboard-hot-100', false),
              }]}
            />
          </div>
          <BillboardHot100 refreshKey={refreshKeys['billboard-hot-100'] ?? 0} onPlayed={highlights.refetch} />
        </section>
      )}
      <div className="grid grid-cols-3 gap-6">
        {crawlers
          .filter(c => c.id !== 'billboard-hot-100' && c.active)
          .map(c => {
            const lastData = c.lastSuccessAt > 0 ? relativeTime(c.lastSuccessAt) : t('crawler.never');
            const isShazam = c.id.startsWith('shazam');
            return (
              <section key={c.id} className="flex flex-col gap-2">
                <div className="flex items-center gap-2 flex-wrap px-1">
                  <span className="font-semibold text-slate-900">{c.displayName}</span>
                  <span className="text-xs text-slate-500">{c.country}</span>
                  <span className="text-xs text-slate-500 ml-auto">
                    {t('crawler.lastData', { ago: lastData })}
                  </span>
                  <OverflowMenu
                    ariaLabel={t('overflow.menu')}
                    items={[{
                      label: t('menu.deactivate'),
                      icon: <EyeOff className="w-4 h-4" />,
                      onClick: () => handleSetActive(c.id, false),
                    }]}
                  />
                </div>
                {isShazam
                  ? <ChartList crawlerId={c.id} refreshKey={refreshKeys[c.id] ?? 0} chartName={c.displayName} country={c.country} onPlayed={highlights.refetch} />
                  : <ChartPager crawlerId={c.id} refreshKey={refreshKeys[c.id] ?? 0} chartName={c.displayName} country={c.country} onPlayed={highlights.refetch} />}
              </section>
            );
          })}
      </div>
      <Highlights data={highlights.data} onPlayed={highlights.refetch} />
    </div>
  );
}

export default App;
