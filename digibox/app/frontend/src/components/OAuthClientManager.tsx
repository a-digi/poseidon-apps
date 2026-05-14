import React, { useEffect, useState } from 'react';
import Dropdown from './ui/Dropdown';
import CreateButton from './ui/Button/CreateButton';
import CancelButton from './ui/Button/CancelButton';
import DeletePill from './ui/Pills/DeletePill';
import InfoAlert from './ui/Alert/InfoAlert';
import ConfirmationDialog from './Modal/ConfirmationDialog';
import { pluginApi } from '@/api/pluginApi';
import type { OAuthClient, Provider } from '@/api/types';
import { useToast } from '@/contexts/ToastContext';

const PROVIDERS: { value: Provider; label: string }[] = [
  { value: 'dropbox', label: 'Dropbox' },
  { value: 'googledrive', label: 'Google Drive' },
];

interface FormState {
  name: string;
  provider: Provider;
  clientId: string;
  secret: string;
  description: string;
}

const EMPTY_FORM: FormState = {
  name: '',
  provider: 'dropbox',
  clientId: '',
  secret: '',
  description: '',
};

export const OAuthClientManager: React.FC = () => {
  const { addToast } = useToast();
  const [clients, setClients] = useState<OAuthClient[]>([]);
  const [form, setForm] = useState<FormState>(EMPTY_FORM);
  const [loading, setLoading] = useState(false);
  const [showForm, setShowForm] = useState(false);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);

  const loadClients = async () => {
    setLoading(true);
    try {
      const res = await pluginApi.listClients();
      setClients(res ?? []);
    } catch (e) {
      addToast({
        message: e instanceof Error ? e.message : 'Failed to load OAuth clients',
        type: 'error',
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadClients();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    setForm((f) => ({ ...f, [e.target.name]: e.target.value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.name || !form.clientId || !form.secret) {
      addToast({ message: 'Name, Client ID and Client Secret are required.', type: 'error' });
      return;
    }
    try {
      await pluginApi.addClient({
        name: form.name,
        provider: form.provider,
        clientId: form.clientId,
        secret: form.secret,
        description: form.description,
      });
      setForm(EMPTY_FORM);
      setShowForm(false);
      addToast({ message: 'OAuth client added.', type: 'success' });
      await loadClients();
    } catch (err) {
      addToast({
        message: err instanceof Error ? err.message : 'Failed to save client',
        type: 'error',
      });
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await pluginApi.deleteClient(id);
      addToast({ message: 'OAuth client deleted.', type: 'success' });
      await loadClients();
    } catch (err) {
      addToast({
        message: err instanceof Error ? err.message : 'Failed to delete client',
        type: 'error',
      });
    }
  };

  const handleCancel = () => {
    setShowForm(false);
    setForm(EMPTY_FORM);
  };

  return (
    <div className="mx-auto p-2">
      <ConfirmationDialog
        open={!!confirmDeleteId}
        onClose={() => setConfirmDeleteId(null)}
        onConfirm={() => {
          if (confirmDeleteId) {
            void handleDelete(confirmDeleteId);
            setConfirmDeleteId(null);
          }
        }}
        title="Delete OAuth client?"
        confirmLabel="Delete"
        cancelLabel="Cancel"
        message="Connected accounts using this client will stop working. This cannot be undone."
      />
      {showForm ? (
        <form onSubmit={handleSubmit} className="space-y-4 mb-8 bg-gray-50 p-4 rounded">
          <InfoAlert
            title="Create OAuth client"
            message="Enter the credentials of an OAuth app you registered with your cloud provider (Dropbox or Google Drive). These are required to connect an account."
            className="mb-4"
          />
          <div>
            <label className="block text-sm font-semibold text-blue-900 mb-1" htmlFor="name">
              Name
            </label>
            <input
              name="name"
              id="name"
              value={form.name}
              onChange={handleChange}
              placeholder="My Dropbox App"
              className="border border-blue-300 px-3 py-2 rounded w-full focus:outline-none focus:ring-2 focus:ring-blue-400 bg-white"
              required
            />
          </div>
          <div>
            <label className="block text-sm font-semibold text-blue-900 mb-1">Provider</label>
            <Dropdown
              label="Select provider"
              items={PROVIDERS}
              selectedValue={form.provider}
              onSelect={(val) => setForm((f) => ({ ...f, provider: val as Provider }))}
            />
          </div>
          <div className="flex gap-4">
            <div className="flex-1">
              <label className="block text-sm font-semibold text-blue-900 mb-1" htmlFor="clientId">
                Client ID
              </label>
              <input
                name="clientId"
                id="clientId"
                value={form.clientId}
                onChange={handleChange}
                placeholder="xxxxxxxxxxxxxxxxxxxx"
                className="border border-blue-300 px-3 py-2 rounded w-full focus:outline-none focus:ring-2 focus:ring-blue-400 bg-white"
                required
              />
            </div>
            <div className="flex-1">
              <label className="block text-sm font-semibold text-blue-900 mb-1" htmlFor="secret">
                Client Secret
              </label>
              <input
                name="secret"
                id="secret"
                type="password"
                value={form.secret}
                onChange={handleChange}
                placeholder="••••••••••••••••••••"
                className="border border-blue-300 px-3 py-2 rounded w-full focus:outline-none focus:ring-2 focus:ring-blue-400 bg-white"
                required
              />
            </div>
          </div>
          <div>
            <label className="block text-sm font-semibold text-blue-900 mb-1" htmlFor="description">
              Description (optional)
            </label>
            <textarea
              name="description"
              id="description"
              value={form.description}
              onChange={handleChange}
              placeholder="Personal Dropbox"
              className="border border-blue-300 px-3 py-2 rounded w-full resize-y min-h-[40px] focus:outline-none focus:ring-2 focus:ring-blue-400 bg-white"
            />
          </div>
          <div className="flex gap-2 mt-2">
            <CreateButton type="submit" label="Save" className="px-6 py-2 rounded-lg" />
            <CancelButton type="button" onClick={handleCancel} label="Cancel" className="px-6 py-2 rounded-lg" />
          </div>
        </form>
      ) : (
        <>
          <CreateButton onClick={() => setShowForm(true)} label="+ Add OAuth client" className="mb-6" />
          <h3 className="font-semibold mb-4 text-gray-800 text-lg border-b border-blue-100 pb-2">
            Existing clients
          </h3>
          {loading ? (
            <div className="text-blue-600">Loading…</div>
          ) : clients.length === 0 ? (
            <div className="text-gray-500">No OAuth clients registered yet.</div>
          ) : (
            <ul className="divide-y divide-blue-100">
              {clients.map((client) => (
                <li
                  key={client.id}
                  className="py-3 flex flex-col md:flex-row md:items-center md:gap-2 bg-white hover:bg-blue-50 transition rounded"
                >
                  <span className="font-mono text-xs text-blue-700">{client.id}</span>
                  <span className="ml-2 font-semibold text-blue-900">
                    {client.name}{' '}
                    <span className="text-xs text-blue-500">
                      ({PROVIDERS.find((p) => p.value === client.provider)?.label ?? client.provider})
                    </span>
                  </span>
                  {!client.builtin && (
                    <div className="ml-auto flex gap-2 mt-2 md:mt-0">
                      <DeletePill
                        className="px-3 py-1 text-xs"
                        onClick={() => setConfirmDeleteId(client.id)}
                      >
                        Delete
                      </DeletePill>
                    </div>
                  )}
                </li>
              ))}
            </ul>
          )}
        </>
      )}
    </div>
  );
};

export default OAuthClientManager;
