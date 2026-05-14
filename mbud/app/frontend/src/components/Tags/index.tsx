import { useEffect, useState } from 'react';
import { Loader2 } from 'lucide-react';
import {
  createTag,
  deleteTag,
  listTag,
  updateTag,
  type Tag,
} from '../../api';
import { AddButton, DeleteButton, EditButton } from '../../common/Buttons';
import { Card, CardGrid } from '../../common/CardGrid';
import { useConfirm } from '../../common/ConfirmationModal';
import { Search } from '../../common/Search';
import { TagForm } from './TagForm';

export function Tags() {
  const [items, setItems] = useState<Tag[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [editing, setEditing] = useState<Tag | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [query, setQuery] = useState('');
  const { confirm, dialog } = useConfirm();

  const refresh = async () => {
    setLoading(true);
    setError(null);
    try {
      setItems(await listTag() ?? []);
    } catch (err) {
      console.error('list_tag failed', err);
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { void refresh(); }, []);

  const handleSave = async (payload: Omit<Tag, 'id' | 'createdAt' | 'updatedAt'>) => {
    if (editing) {
      await updateTag(editing.id, { ...editing, ...payload });
    } else {
      await createTag(payload);
    }
    await refresh();
  };

  const handleDelete = async (t: Tag) => {
    if (!await confirm(`Delete tag "${t.name}"?`)) return;
    try {
      await deleteTag(t.id);
      await refresh();
    } catch (err) {
      console.error('delete_tag failed', err);
      setError(err instanceof Error ? err.message : String(err));
    }
  };

  const q = query.trim().toLowerCase();
  const visible = q ? items.filter(t => t.name.toLowerCase().includes(q)) : items;

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-slate-900">Tags</h2>
        <AddButton onClick={() => { setEditing(null); setModalOpen(true); }} />
      </div>
      <Search value={query} onChange={setQuery} placeholder="Search tags…" />
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
          emptyMessage={q ? `No matches for "${query.trim()}"` : 'No tags yet.'}
          renderCard={t => (
            <Card>
              <div className="flex items-start gap-3">
                <div className="min-w-0 flex-1">
                  <p className="font-semibold text-slate-900 truncate"># {t.name}</p>
                </div>
                <div className="flex items-center gap-1 shrink-0">
                  <EditButton onClick={() => { setEditing(t); setModalOpen(true); }} />
                  <DeleteButton onClick={() => handleDelete(t)} />
                </div>
              </div>
            </Card>
          )}
        />
      )}
      {modalOpen && (
        <TagForm initial={editing} onSave={handleSave} onClose={() => setModalOpen(false)} />
      )}
      {dialog}
    </div>
  );
}
