export type Provider = 'dropbox' | 'googledrive';

export interface AccountInfo {
  displayName?: string;
  email?: string;
  provider: Provider;
}

export interface OAuthClient {
  id: string;
  name: string;
  provider: Provider;
  clientId: string;
  secret: string;
  description?: string;
  builtin?: boolean;
}

export interface OauthToken {
  id: string;
  clientId: string;
  provider: Provider;
  accessToken: string;
  refreshToken?: string;
  expiry?: number;
  requestId?: string;
}

export interface OauthProfile {
  id: string;
  email: string;
  displayName: string;
  givenName: string;
  surname: string;
  profilePhoto: string;
  provider: string;
  accountType: string;
  country: string;
  tokenId: string;
}

export interface OauthRequestView {
  id: string;
  requestedOn: number;
  oauthClientId?: string;
  oauthToken: OauthToken;
  oauthProfile?: OauthProfile;
}

export interface StorageEntry {
  tag: 'folder' | 'file';
  id?: string;
  name: string;
  path?: string;
  downloaded?: boolean;
  targetFolder?: string;
}

export interface AddClientInput {
  name: string;
  provider: Provider;
  clientId: string;
  secret: string;
  description?: string;
}

export interface ExchangeCodeInput {
  state: string;
  code: string;
}

export interface AuthLink {
  url: string;
  redirectUri: string;
}

export type OAuthCodePoll =
  | { pending: true }
  | { code: string; state: string };

export interface ListFilesResult {
  entries: StorageEntry[];
}

export interface ListResult<T> {
  data: T[];
}
