import { useEffect, useState } from 'react';
import { Loader2 } from 'lucide-react';
import {
  createUser,
  deleteUser,
  listUser,
  updateUser,
  type User,
} from '../../api';
import { AddButton, DeleteButton, EditButton } from '../../common/Buttons';
import { Card, CardGrid } from '../../common/CardGrid';
import { useConfirm } from '../../common/ConfirmationModal';
import { Search } from '../../common/Search';
import { UserForm } from './UserForm';

export function Users() {
  const [items, setItems] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [editing, setEditing] = useState<User | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [query, setQuery] = useState('');
  const { confirm, dialog } = useConfirm();

  const refresh = async () => {
    setLoading(true);
    setError(null);
    try {
      setItems(await listUser() ?? []);
    } catch (err) {
      console.error('list_user failed', err);
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { void refresh(); }, []);

  const handleSave = async (payload: Omit<User, 'id' | 'createdAt' | 'updatedAt'>) => {
    if (editing) {
      await updateUser(editing.id, { ...editing, ...payload });
    } else {
      await createUser(payload);
    }
    await refresh();
  };

  const handleDelete = async (u: User) => {
    if (!await confirm(`Delete user "${u.name}"?`)) return;
    try {
      await deleteUser(u.id);
      await refresh();
    } catch (err) {
      console.error('delete_user failed', err);
      setError(err instanceof Error ? err.message : String(err));
    }
  };

  const q = query.trim().toLowerCase();
  const visible = q ? items.filter(u => u.name.toLowerCase().includes(q)) : items;

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-slate-900">Users</h2>
        <AddButton onClick={() => { setEditing(null); setModalOpen(true); }} />
      </div>
      <Search value={query} onChange={setQuery} placeholder="Search users…" />
      {error && (
        <div className="text-sm text-red-700 bg-red-50 border border-red-200 rounded-md px-3 py-2">{error}</div>
      )}
      {loading ? (
        <div className="flex items-center gap-2 text-sm text-slate-500">
          <Loader2 className="w-4 h-4 animate-spin" />Loading...
        </div>
      ) : (
        <CardGrid
          items={visible}
          emptyMessage={q ? `No matches for "${query.trim()}"` : 'No users yet.'}
          renderCard={u => (
            <Card>
              <div className="flex items-start gap-3">
                <div className="min-w-0 flex-1">
                  <p className="font-semibold text-slate-900 truncate">{u.name}</p>
                  {u.email && <p className="text-xs text-slate-500 mt-1">{u.email}</p>}
                  {u.notes && (
                    <p className="mt-2 pt-2 border-t border-slate-100 text-xs text-slate-400 line-clamp-2">{u.notes}</p>
                  )}
                </div>
                <div className="flex items-center gap-1 shrink-0">
                  <EditButton onClick={() => { setEditing(u); setModalOpen(true); }} />
                  <DeleteButton onClick={() => handleDelete(u)} />
                </div>
              </div>
            </Card>
          )}
        />
      )}
      {modalOpen && (
        <UserForm initial={editing} onSave={handleSave} onClose={() => setModalOpen(false)} />
      )}
      {dialog}
    </div>
  );
}
