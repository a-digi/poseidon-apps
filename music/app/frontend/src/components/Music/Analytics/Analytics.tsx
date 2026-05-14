import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardsContainer } from '@/components/Card';
import { BarChart } from './BarChart';
import type { AnalyticsOverview } from './types';

const wails = (window as any).go?.main?.App;

const WEEKDAY_LABELS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

type TopTab = 'tracks' | 'playlists' | 'artists' | 'albums' | 'genres';

const tabBtnClass = (active: boolean) =>
  `inline-flex items-center gap-2 px-5 py-3 text-sm border-b-2 -mb-px transition-colors ${
    active
      ? 'border-blue-600 text-blue-700 font-semibold'
      : 'border-transparent text-slate-600 hover:text-slate-900 hover:border-slate-300'
  }`;

export function Analytics() {
  const { t } = useTranslation();
  const [data, setData] = useState<AnalyticsOverview | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTopTab, setActiveTopTab] = useState<TopTab>('tracks');

  async function fetchData() {
    setLoading(true);
    setError(null);
    try {
      const result = await wails?.GetAnalyticsOverview?.();
      if (!result) {
        setError('Not available');
        return;
      }
      const parsed = JSON.parse(result as string);
      if (parsed.status !== 'success') {
        setError(parsed.message);
        return;
      }
      setData(parsed.message as AnalyticsOverview);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    fetchData();
  }, []);

  function formatDate(ts?: number): string {
    if (!ts) return t('analytics.summary.never', 'Never');
    return new Date(ts * 1000).toLocaleDateString();
  }

  return (
    <div className="w-full space-y-8 pb-8">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-slate-800">{t('analytics.title', 'Analytics')}</h1>
        <button
          onClick={fetchData}
          disabled={loading}
          className="px-4 py-2 text-sm font-medium bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
        >
          {t('analytics.refresh', 'Refresh')}
        </button>
      </div>

      {loading && <p className="text-slate-500 text-sm">Loading…</p>}

      {error && <p className="text-red-500 text-sm">{error}</p>}

      {!loading && !error && data && data.summary.totalPlays === 0 && (
        <p className="text-slate-400 italic">{t('analytics.empty', 'No plays recorded yet.')}</p>
      )}

      {data && data.summary.totalPlays > 0 && (
        <>
          <CardsContainer columns="1fr 1fr 1fr 1fr" gap="1rem">
            <Card color="darkGradient" shadow="md" centered title={t('analytics.summary.totalPlays', 'Total Plays')}>
              <span className="text-3xl font-bold text-white">{data.summary.totalPlays}</span>
            </Card>
            <Card color="darkGradient" shadow="md" centered title={t('analytics.summary.uniqueItems', 'Unique Tracks')}>
              <span className="text-3xl font-bold text-white">{data.summary.uniqueItemsPlayed}</span>
            </Card>
            <Card color="darkGradient" shadow="md" centered title={t('analytics.summary.uniquePlaylists', 'Unique Playlists')}>
              <span className="text-3xl font-bold text-white">{data.summary.uniquePlaylistsPlayed}</span>
            </Card>
            <Card color="darkGradient" shadow="md" centered title={t('analytics.summary.lastPlay', 'Last Played')}>
              <span className="text-lg font-semibold text-white">{formatDate(data.summary.lastPlayAt)}</span>
            </Card>
          </CardsContainer>

          {(() => {
            const tabs: { id: TopTab; label: string; count: number }[] = [
              { id: 'tracks',    label: t('analytics.topItems',     'Top Tracks'),    count: data.topItems.length },
              { id: 'playlists', label: t('analytics.topPlaylists', 'Top Playlists'), count: data.topPlaylists.length },
              { id: 'artists',   label: t('analytics.topArtists',   'Top Artists'),   count: data.topArtists.length },
              { id: 'albums',    label: t('analytics.topAlbums',    'Top Albums'),    count: data.topAlbums.length },
              { id: 'genres',    label: t('analytics.topGenres',    'Top Genres'),    count: data.topGenres.length },
            ];
            const visible = tabs.filter((tab) => tab.count > 0);
            if (visible.length === 0) return null;
            const active: TopTab = visible.some((tab) => tab.id === activeTopTab)
              ? activeTopTab
              : visible[0].id;
            return (
              <div className="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
                <nav
                  aria-label={t('analytics.topNav', 'Top rankings')}
                  className="border-b border-slate-200 px-2 flex gap-1 overflow-x-auto"
                >
                  {visible.map((tab) => (
                    <button
                      key={tab.id}
                      type="button"
                      onClick={() => setActiveTopTab(tab.id)}
                      className={tabBtnClass(tab.id === active)}
                    >
                      {tab.label}
                      <span className="text-xs text-slate-400 tabular-nums">{tab.count}</span>
                    </button>
                  ))}
                </nav>

                {active === 'tracks' && (
                  <table className="w-full text-sm">
                    <thead className="bg-slate-50 text-slate-500 text-xs uppercase tracking-wide">
                      <tr>
                        <th className="px-5 py-2 text-left">#</th>
                        <th className="px-5 py-2 text-left">{t('analytics.title', 'Title')}</th>
                        <th className="px-5 py-2 text-left">{t('analytics.artistAlbum', 'Artist / Album')}</th>
                        <th className="px-5 py-2 text-right">{t('analytics.plays', 'Plays')}</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100">
                      {data.topItems.map((item, i) => {
                        const isFile = item.itemId.startsWith('file:');
                        return (
                          <tr key={item.itemId} className="hover:bg-slate-50">
                            <td className="px-5 py-2 text-slate-400 tabular-nums">{i + 1}</td>
                            <td className="px-5 py-2 font-medium text-slate-900">{item.title}</td>
                            <td className="px-5 py-2 text-slate-500">
                              {isFile
                                ? <span className="italic text-xs">{t('analytics.singleFile', 'Single file')}</span>
                                : <>{item.artist}{item.album ? ` · ${item.album}` : ''}</>
                              }
                            </td>
                            <td className="px-5 py-2 text-right font-semibold tabular-nums">{item.plays}</td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                )}

                {active === 'playlists' && (
                  <table className="w-full text-sm">
                    <thead className="bg-slate-50 text-slate-500 text-xs uppercase tracking-wide">
                      <tr>
                        <th className="px-5 py-2 text-left">#</th>
                        <th className="px-5 py-2 text-left">{t('analytics.name', 'Name')}</th>
                        <th className="px-5 py-2 text-right">{t('analytics.plays', 'Plays')}</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100">
                      {data.topPlaylists.map((pl, i) => (
                        <tr key={pl.playlistId} className="hover:bg-slate-50">
                          <td className="px-5 py-2 text-slate-400 tabular-nums">{i + 1}</td>
                          <td className="px-5 py-2 font-medium text-slate-900">{pl.name}</td>
                          <td className="px-5 py-2 text-right font-semibold tabular-nums">{pl.plays}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                )}

                {active === 'artists' && (
                  <table className="w-full text-sm">
                    <thead className="bg-slate-50 text-slate-500 text-xs uppercase tracking-wide">
                      <tr>
                        <th className="px-5 py-2 text-left">{t('analytics.artist', 'Artist')}</th>
                        <th className="px-5 py-2 text-right">{t('analytics.plays', 'Plays')}</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100">
                      {data.topArtists.map((a) => (
                        <tr key={a.artist} className="hover:bg-slate-50">
                          <td className="px-5 py-2 font-medium text-slate-900">{a.artist}</td>
                          <td className="px-5 py-2 text-right font-semibold tabular-nums">{a.plays}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                )}

                {active === 'albums' && (
                  <table className="w-full text-sm">
                    <thead className="bg-slate-50 text-slate-500 text-xs uppercase tracking-wide">
                      <tr>
                        <th className="px-5 py-2 text-left">{t('analytics.albumArtist', 'Album — Artist')}</th>
                        <th className="px-5 py-2 text-right">{t('analytics.plays', 'Plays')}</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100">
                      {data.topAlbums.map((a) => (
                        <tr key={`${a.artist}|${a.album}`} className="hover:bg-slate-50">
                          <td className="px-5 py-2 font-medium text-slate-900">{a.album} — {a.artist}</td>
                          <td className="px-5 py-2 text-right font-semibold tabular-nums">{a.plays}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                )}

                {active === 'genres' && (
                  <table className="w-full text-sm">
                    <thead className="bg-slate-50 text-slate-500 text-xs uppercase tracking-wide">
                      <tr>
                        <th className="px-5 py-2 text-left">{t('analytics.genre', 'Genre')}</th>
                        <th className="px-5 py-2 text-right">{t('analytics.plays', 'Plays')}</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100">
                      {data.topGenres.map((g) => (
                        <tr key={g.genre} className="hover:bg-slate-50">
                          <td className="px-5 py-2 font-medium text-slate-900">{g.genre}</td>
                          <td className="px-5 py-2 text-right font-semibold tabular-nums">{g.plays}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                )}
              </div>
            );
          })()}

          <div className="grid grid-cols-2 gap-6">
            <div className="bg-white rounded-xl border border-slate-200 shadow-sm p-5">
              <h3 className="text-sm font-semibold text-slate-700 mb-4">{t('analytics.byHour', 'By Hour of Day')}</h3>
              <BarChart ariaLabel={t('analytics.byHour', 'By Hour of Day')} data={data.byHour.map(b => ({ label: b.key, value: b.plays }))} />
            </div>
            <div className="bg-white rounded-xl border border-slate-200 shadow-sm p-5">
              <h3 className="text-sm font-semibold text-slate-700 mb-4">{t('analytics.byWeekday', 'By Weekday')}</h3>
              <BarChart ariaLabel={t('analytics.byWeekday', 'By Weekday')} data={data.byWeekday.map(b => ({ label: WEEKDAY_LABELS[parseInt(b.key)] ?? b.key, value: b.plays }))} />
            </div>
            <div className="bg-white rounded-xl border border-slate-200 shadow-sm p-5">
              <h3 className="text-sm font-semibold text-slate-700 mb-4">{t('analytics.byMonth', 'By Month')}</h3>
              <BarChart ariaLabel={t('analytics.byMonth', 'By Month')} data={data.byMonth.map(b => ({ label: b.key, value: b.plays }))} />
            </div>
            <div className="bg-white rounded-xl border border-slate-200 shadow-sm p-5">
              <h3 className="text-sm font-semibold text-slate-700 mb-4">{t('analytics.byYear', 'By Year')}</h3>
              <BarChart ariaLabel={t('analytics.byYear', 'By Year')} data={data.byYear.map(b => ({ label: b.key, value: b.plays }))} />
            </div>
          </div>
        </>
      )}
    </div>
  );
}
