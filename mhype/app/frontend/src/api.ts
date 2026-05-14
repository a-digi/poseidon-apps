const params = new URLSearchParams(window.location.search);
const PLUGIN_ID = params.get('pluginId') ?? 'mhype';
const BACKEND_URL = (params.get('backendUrl') ?? window.location.origin).replace(/\/$/, '');

export type OrchestratorStatus = {
  running: boolean;
  pid?: number;
  startedAt?: number;
  exitCode?: number;
};

export type CrawlerInfo = {
  id: string;
  displayName: string;
  source: string;
  country: string;
  intervalSec: number;
  fileCount: number;
  lastSuccessAt: number;
  active: boolean;
};

export type ItemKind = 'track' | 'chart-entry' | 'release' | 'news';

export interface ChartContext {
  name: string;
  position: number;
  prevPosition?: number;
  peakPosition?: number;
  weeksOnChart?: number;
  periodStart?: string;
  periodEnd?: string;
}

export interface NewsContext {
  summary?: string;
  publishedAt?: string;
  relatedArtists?: string[];
}

export interface Item {
  id: string;
  kind: ItemKind;
  crawlerId: string;
  sourceUrl: string;
  title: string;
  scrapedAt: number;

  artists?: string[];
  album?: string;
  label?: string;
  genres?: string[];
  durationSec?: number;
  releasedAt?: string;

  sourceId?: string;
  externalIds?: Record<string, string>;

  artworkUrl?: string;
  previewUrl?: string;

  chart?: ChartContext;
  news?: NewsContext;

  extra?: Record<string, string>;
}

export async function callPlugin<T>(action: string, params: Record<string, unknown> = {}): Promise<T> {
  const res = await fetch(`${BACKEND_URL}/api/plugins/execute`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pluginName: PLUGIN_ID, params: { action, ...params } }),
  });
  if (!res.ok) throw new Error(`HTTP ${res.status}: ${res.statusText}`);
  const json = (await res.json()) as { result?: T; error?: string };
  if (json.error) throw new Error(json.error);
  return json.result as T;
}

async function callLongRunning<T>(method: 'POST' | 'GET', path: string, body?: unknown): Promise<T> {
  const res = await fetch(`${BACKEND_URL}${path}`, {
    method,
    headers: body ? { 'Content-Type': 'application/json' } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  });
  const env = (await res.json()) as { status?: string; message?: T | string };
  if (!res.ok || env?.status === 'error') {
    throw new Error(typeof env?.message === 'string' ? env.message : `HTTP ${res.status}: ${res.statusText}`);
  }
  return env.message as T;
}

export const ensureOrchestrator = () =>
  callLongRunning<OrchestratorStatus>('POST', '/api/plugins/long-running/start', {
    pluginId: PLUGIN_ID,
    args: ['--orchestrator'],
  });
export const stopOrchestrator = () =>
  callLongRunning<{ stopped: true }>('POST', '/api/plugins/long-running/stop', { pluginId: PLUGIN_ID });
export const restartOrchestrator = () =>
  callLongRunning<OrchestratorStatus>('POST', '/api/plugins/long-running/restart', {
    pluginId: PLUGIN_ID,
    args: ['--orchestrator'],
  });
export const orchestratorStatus = () =>
  callLongRunning<OrchestratorStatus>(
    'GET',
    `/api/plugins/long-running/status?pluginId=${encodeURIComponent(PLUGIN_ID)}`,
  );
export const listCrawlers = () => callPlugin<CrawlerInfo[]>('list_crawlers');
export const listItems = (crawlerId: string, limit = 50) => callPlugin<Item[]>('list_items', { crawlerId, limit });
export const triggerCrawl = (crawlerId: string) => callPlugin<{ ok: true }>('trigger_crawl', { crawlerId });
export const setCrawlerActive = (crawlerId: string, active: boolean) =>
  callPlugin<{ ok: true }>('set_crawler_active', { crawlerId, active });

export interface YouTubeResult {
  videoId: string;
  title: string;
  channelTitle: string;
  thumbnailUrl: string;
}

export interface YouTubeSuggestionsResult {
  found: boolean;
  results: YouTubeResult[];
}

export const getYouTubeSuggestions = (artist: string, title: string) =>
  callPlugin<YouTubeSuggestionsResult>('get_youtube_suggestions', { artist, title });

export interface FindYouTubeVideoResult {
  videoId: string;
  title: string;
  channelTitle: string;
  thumbnailUrl: string;
}

export const findYouTubeVideo = (artist: string, title: string) =>
  callPlugin<FindYouTubeVideoResult>('find_youtube_video', { artist, title });

export interface TrackPlayInput {
  title: string;
  artists: string[];
  artworkUrl?: string;
  chartName?: string;
  crawlerId?: string;
  position?: number;
  country?: string;
}

export interface HighlightArtist {
  name: string;
  count: number;
  latestArtworkUrl?: string;
  latestPlayedAt: number;
}

export interface HighlightSong {
  title: string;
  artists: string[];
  artworkUrl?: string;
  chartName?: string;
  position?: number;
  country?: string;
  count: number;
  latestPlayedAt: number;
}

export interface HighlightList {
  crawlerId: string;
  displayName: string;
  country?: string;
  count: number;
  latestPlayedAt: number;
}

export interface Highlights {
  artists: HighlightArtist[];
  songs: HighlightSong[];
  playlists: HighlightList[];
}

export const trackPlay = (input: TrackPlayInput) =>
  callPlugin<{ ok: true }>('track_play', input as unknown as Record<string, unknown>);

export const getHighlights = (limit = 10) =>
  callPlugin<Highlights>('get_highlights', { limit });
