import type {
  AddClientInput,
  AuthLink,
  ExchangeCodeInput,
  ListFilesResult,
  OAuthClient,
  OAuthCodePoll,
  OauthRequestView,
} from './types';

const params = new URLSearchParams(window.location.search);
const PLUGIN_ID = params.get('pluginId') ?? 'digibox';
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

export const pluginApi = {
  openDirectoryDialog: async (): Promise<string> => {
    const res = await fetch(`${BACKEND_URL}/api/system/open-directory-dialog`, { method: 'POST' });
    const json = (await res.json()) as { path?: string };
    return json.path ?? '';
  },

  openFileDialog: async (): Promise<string> => {
    const res = await fetch(`${BACKEND_URL}/api/system/open-file-dialog`, { method: 'POST' });
    const json = (await res.json()) as { path?: string };
    return json.path ?? '';
  },

  initTables: () => execute<{ ok: boolean }>('init_tables'),

  listClients: () => execute<OAuthClient[]>('list_clients'),
  addClient: (data: AddClientInput) =>
    execute<unknown>('add_client', data as unknown as Record<string, unknown>),
  deleteClient: (id: string) => execute<unknown>('delete_client', { id }),

  listAuthorizations: () => execute<OauthRequestView[]>('list_authorizations'),
  deleteAuthorization: (id: string) => execute<{ ok: boolean }>('delete_authorization', { id }),

  createAuthLink: (recordId: string) =>
    execute<AuthLink>('create_auth_link', { clientId: recordId }),
  getOAuthCode: () => execute<OAuthCodePoll>('get_oauth_code'),
  exchangeCode: (input: ExchangeCodeInput) =>
    execute<unknown>('exchange_code', input as unknown as Record<string, unknown>),

  listFiles: (tokenId: string, path: string) =>
    execute<ListFilesResult>('list_files', { tokenId, path }),

  downloadFile: (tokenId: string, path: string, targetDir: string) =>
    execute<{ ok: boolean; path: string }>('download_file', { tokenId, path, targetDir }),

  createFolder: (tokenId: string, path: string) =>
    execute<{ ok: boolean }>('create_folder', { tokenId, path }),

  uploadFile: (tokenId: string, path: string, sourcePath: string) =>
    execute<{ ok: boolean; path: string }>('upload_file', { tokenId, path, sourcePath }),

  deleteItem: (tokenId: string, path: string) =>
    execute<{ ok: boolean }>('delete_item', { tokenId, path }),
};

export type PluginApi = typeof pluginApi;
