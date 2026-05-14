import { pluginApi, type PlayerState } from '@/api/pluginApi';

declare global {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  interface Window {
    go?: any;
  }
}

const App = {
  GetAudioDataUrl: (path: string) => pluginApi.getAudioDataUrl(path),
  ListPlaylists: () => pluginApi.listPlaylists(),
  GetPlaylistByID: (id: string) => pluginApi.getPlaylistById(id),
  CreatePlaylist: (name: string) => pluginApi.createPlaylist(name),
  EditPlaylist: (id: string, name: string) => pluginApi.editPlaylist(id, name),
  DeletePlaylist: (id: string) => pluginApi.deletePlaylist(id),
  AddPlaylistItem: (playlistId: string, item: unknown) => pluginApi.addPlaylistItem(playlistId, item),
  DeletePlaylistItem: (playlistId: string, itemId: string) => pluginApi.deletePlaylistItem(playlistId, itemId),
  GetPlayerState: () => pluginApi.getPlayerState(),
  SavePlayerState: (state: PlayerState) => pluginApi.savePlayerState(state),
  RecordPlay: (args: { itemId: string; playlistId?: string; title: string; artist?: string; album?: string }) =>
    pluginApi.recordPlay(args),
  GetAnalyticsOverview: () => pluginApi.getAnalyticsOverview(),
  OpenFileDialog: () => pluginApi.openAudioFileDialog(),
};

window.go = { main: { App: App as unknown as Record<string, (...args: unknown[]) => Promise<unknown>> } };

export {};
