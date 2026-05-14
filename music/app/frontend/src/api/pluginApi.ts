const params = new URLSearchParams(window.location.search);
const PLUGIN_ID = params.get('pluginId') ?? 'music';
const BACKEND_URL = (params.get('backendUrl') ?? window.location.origin).replace(/\/$/, '');

async function execute<T>(action: string, data: Record<string, unknown> = {}): Promise<T> {
  const res = await fetch(`${BACKEND_URL}/api/plugins/execute`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pluginName: PLUGIN_ID, params: { action, ...data } }),
  });
  const json = (await res.json()) as { result?: T; error?: string };
  if (json.error) throw new Error(json.error);
  return json.result as T;
}

export interface PlayerState {
  selectedPlaylistId: string;
  playMode: 'playlist' | 'repeat' | 'shuffle';
  currentItemId: string;
}

const toEnvelope = async (action: string, data: Record<string, unknown> = {}): Promise<string> => {
  try {
    const result = await execute<unknown>(action, data);
    if (typeof result === 'string') return result;
    return JSON.stringify(result);
  } catch (err) {
    return JSON.stringify({ status: 'error', message: err instanceof Error ? err.message : String(err) });
  }
};

export const pluginApi = {
  initTables: () => execute<{ ok: boolean }>('init_tables'),

  listPlaylists: () => toEnvelope('list_playlists'),
  createPlaylist: (name: string) => toEnvelope('create_playlist', { name }),
  editPlaylist: (id: string, name: string) => toEnvelope('edit_playlist', { id, name }),
  deletePlaylist: (id: string) => toEnvelope('delete_playlist', { id }),
  getPlaylistById: (id: string) => toEnvelope('get_playlist_by_id', { id }),
  addPlaylistItem: (playlistId: string, item: unknown) => toEnvelope('add_playlist_item', { playlistId, item }),
  deletePlaylistItem: (playlistId: string, itemId: string) => toEnvelope('delete_playlist_item', { playlistId, itemId }),

  getPlayerState: () => toEnvelope('get_player_state'),
  savePlayerState: (state: PlayerState) =>
    toEnvelope('save_player_state', state as unknown as Record<string, unknown>),

  recordPlay: (args: {
    itemId: string;
    playlistId?: string;
    title: string;
    artist?: string;
    album?: string;
  }) => toEnvelope('record_play', args as Record<string, unknown>),

  getAnalyticsOverview: () => toEnvelope('analytics_overview'),

  getAudioDataUrl: async (path: string): Promise<string> => {
    const res = await execute<{ dataUrl: string }>('get_audio_data_url', { path });
    return res.dataUrl;
  },

  openFileDialog: async (): Promise<string> => {
    const res = await fetch(`${BACKEND_URL}/api/system/open-file-dialog`, { method: 'POST' });
    const json = (await res.json()) as { path?: string };
    return json.path ?? '';
  },

  openAudioFileDialog: async (): Promise<string> => {
    const res = await fetch(`${BACKEND_URL}/api/system/open-audio-file-dialog`, { method: 'POST' });
    const json = (await res.json()) as { path?: string };
    return json.path ?? '';
  },
};
