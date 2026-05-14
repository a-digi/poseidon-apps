import React, { useEffect, useRef, useState } from 'react';
import CreatePill from './ui/Pills/CreatePill';
import DeletePill from './ui/Pills/DeletePill';
import { pluginApi } from '@/api/pluginApi';
import type { AccountInfo, OAuthClient, OauthRequestView } from '@/api/types';
import { useToast } from '@/contexts/ToastContext';

interface CloudProviderProps {
  onBrowse: (tokenId: string, account: AccountInfo) => void;
}

type AuthStatus =
  | { state: 'idle' }
  | { state: 'awaiting'; clientId: string; clientName: string }
  | { state: 'exchanging' };

const POLL_INTERVAL_MS = 2000;
const POLL_TIMEOUT_MS = 5 * 60 * 1000;

export const CloudProvider: React.FC<CloudProviderProps> = ({ onBrowse }) => {
  const { addToast } = useToast();
  const [clients, setClients] = useState<OAuthClient[]>([]);
  const [authorizations, setAuthorizations] = useState<OauthRequestView[]>([]);
  const [loading, setLoading] = useState(true);
  const [auth, setAuth] = useState<AuthStatus>({ state: 'idle' });
  const pollRef = useRef<number | null>(null);
  const pollDeadlineRef = useRef<number>(0);

  const stopPolling = () => {
    if (pollRef.current !== null) {
      window.clearInterval(pollRef.current);
      pollRef.current = null;
    }
  };

  const loadAll = async () => {
    setLoading(true);
    try {
      const [clientsRes, authsRes] = await Promise.all([
        pluginApi.listClients(),
        pluginApi.listAuthorizations(),
      ]);
      setClients(clientsRes ?? []);
      setAuthorizations(authsRes ?? []);
    } catch (err) {
      addToast({
        message: err instanceof Error ? err.message : 'Failed to load',
        type: 'error',
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadAll();
    return () => stopPolling();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleAuthorize = async (client: OAuthClient) => {
    try {
      await pluginApi.createAuthLink(client.id);
      setAuth({ state: 'awaiting', clientId: client.id, clientName: client.name });
      addToast({
        message: `Browser opened for ${client.name}. Complete the authorization and return here.`,
        type: 'info',
      });
      pollDeadlineRef.current = Date.now() + POLL_TIMEOUT_MS;
      pollRef.current = window.setInterval(() => void pollForCode(client), POLL_INTERVAL_MS);
    } catch (err) {
      addToast({
        message: err instanceof Error ? err.message : 'Failed to start authorization',
        type: 'error',
      });
    }
  };

  const pollForCode = async (client: OAuthClient) => {
    if (Date.now() > pollDeadlineRef.current) {
      stopPolling();
      setAuth({ state: 'idle' });
      addToast({ message: 'Authorization timed out. Please try again.', type: 'warning' });
      return;
    }
    try {
      const res = await pluginApi.getOAuthCode();
      if ('pending' in res) return;

      stopPolling();
      setAuth({ state: 'exchanging' });
      await pluginApi.exchangeCode({ state: res.state, code: res.code });
      addToast({ message: `${client.name} authorized successfully.`, type: 'success' });
      setAuth({ state: 'idle' });
      await loadAll();
    } catch (err) {
      stopPolling();
      setAuth({ state: 'idle' });
      addToast({
        message: err instanceof Error ? err.message : 'Authorization failed',
        type: 'error',
      });
    }
  };

  const handleDisconnect = async (authId: string) => {
    try {
      await pluginApi.deleteAuthorization(authId);
      addToast({ message: 'Account disconnected.', type: 'success' });
      await loadAll();
    } catch (err) {
      addToast({
        message: err instanceof Error ? err.message : 'Failed to disconnect',
        type: 'error',
      });
    }
  };

  if (loading) return <div className="text-blue-600 p-4">Loading…</div>;

  if (clients.length === 0) {
    return (
      <div className="text-gray-500 p-4">
        No OAuth clients registered yet. Add one in the <strong>OAuth Apps</strong> tab first.
      </div>
    );
  }

  const isAuthorizing = auth.state !== 'idle';
  const authsByClient = (internalClientId: string) =>
    authorizations.filter((a) => a.oauthClientId === internalClientId);

  return (
    <div className="p-4 space-y-6">
      {isAuthorizing && (
        <div className="bg-blue-100 border border-blue-300 text-blue-800 rounded p-3">
          {auth.state === 'awaiting'
            ? `Waiting for browser authorization (${auth.clientName})… Polling for code.`
            : 'Exchanging authorization code…'}
        </div>
      )}

      {clients.map((client) => {
        const linked = authsByClient(client.id);
        return (
          <div key={client.id} className="bg-blue-50 rounded p-4 shadow-sm">
            <div className="flex items-center justify-between mb-3">
              <div>
                <span className="text-lg font-semibold text-blue-900">{client.name}</span>
                <span className="ml-2 text-xs text-blue-500">
                  ({client.provider === 'dropbox' ? 'Dropbox' : 'Google Drive'})
                </span>
              </div>
              <CreatePill
                onClick={() => void handleAuthorize(client)}
                disabled={isAuthorizing}
                className="px-4 py-2 text-sm"
              >
                + Connect
              </CreatePill>
            </div>
            {linked.length === 0 ? (
              <div className="text-sm text-gray-500 italic">No account connected yet.</div>
            ) : (
              <ul className="space-y-2">
                {linked.map((a) => (
                  <li
                    key={a.id}
                    className="bg-white border border-blue-100 rounded p-3 flex items-center justify-between"
                  >
                    <div>
                      <div className="font-semibold text-blue-900">
                        {a.oauthProfile?.displayName || a.oauthProfile?.email || 'Unknown account'}
                      </div>
                      <div className="text-xs text-gray-500">
                        {a.oauthProfile?.email}
                      </div>
                    </div>
                    <div className="flex gap-2">
                      <CreatePill
                        className="px-3 py-1 text-xs"
                        onClick={() =>
                          onBrowse(a.oauthToken.id, {
                            displayName: a.oauthProfile?.displayName,
                            email: a.oauthProfile?.email,
                            provider: a.oauthToken.provider,
                          })
                        }
                      >
                        Browse
                      </CreatePill>
                      <DeletePill
                        className="px-3 py-1 text-xs"
                        onClick={() => void handleDisconnect(a.id)}
                      >
                        Disconnect
                      </DeletePill>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </div>
        );
      })}
    </div>
  );
};

export default CloudProvider;
