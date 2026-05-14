import { useEffect, useState } from 'react';
import { Loader2 } from 'lucide-react';
import {
  createBusiness,
  deleteBusiness,
  deleteBusinessLogo,
  listBusiness,
  listTag,
  updateBusiness,
  uploadBusinessLogo,
  type Business,
  type Tag,
} from '../../api';
import { AddButton, DeleteButton, EditButton } from '../../common/Buttons';
import { Card, CardGrid } from '../../common/CardGrid';
import { useConfirm } from '../../common/ConfirmationModal';
import { Search } from '../../common/Search';
import { BusinessForm, type LogoAction } from './BusinessForm';
import { BusinessLogo } from './BusinessLogo';

export function Businesses() {
  const [items, setItems] = useState<Business[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [editing, setEditing] = useState<Business | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [query, setQuery] = useState('');
  const { confirm, dialog } = useConfirm();

  const refresh = async () => {
    setLoading(true);
    setError(null);
    try {
      const [biz, tgs] = await Promise.all([listBusiness(), listTag()]);
      setItems(biz ?? []);
      setTags(tgs ?? []);
    } catch (err) {
      console.error('list_business failed', err);
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { void refresh(); }, []);

  const handleSave = async (
    payload: Omit<Business, 'id' | 'createdAt' | 'updatedAt' | 'logoType' | 'logo'>,
    logoAction: LogoAction,
  ) => {
    const saved = editing
      ? await updateBusiness(editing.id, { ...editing, ...payload })
      : await createBusiness(payload);
    if (logoAction.type === 'upload') {
      await uploadBusinessLogo(saved.id, logoAction.dataUrl);
    } else if (logoAction.type === 'delete') {
      await deleteBusinessLogo(saved.id);
    }
    await refresh();
  };

  const handleDelete = async (b: Business) => {
    if (!await confirm(`Delete business "${b.name}"?`)) return;
    try {
      await deleteBusiness(b.id);
      await refresh();
    } catch (err) {
      console.error('delete_business failed', err);
      setError(err instanceof Error ? err.message : String(err));
    }
  };

  const q = query.trim().toLowerCase();
  const visible = q ? items.filter(b => b.name.toLowerCase().includes(q)) : items;

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-slate-900">Businesses</h2>
        <AddButton onClick={() => { setEditing(null); setModalOpen(true); }} />
      </div>
      <Search value={query} onChange={setQuery} placeholder="Search businesses…" />
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
          emptyMessage={q ? `No matches for "${query.trim()}"` : 'No businesses yet.'}
          renderCard={b => (
            <Card>
              <div className="flex items-start gap-3">
                <BusinessLogo name={b.name} logo={b.logo} size={48} />
                <div className="min-w-0 flex-1">
                  <p className="font-semibold text-slate-900 truncate">{b.name}</p>
                  {(b.taxId || b.email) && (
                    <div className="mt-1 flex flex-col gap-0.5">
                      {b.taxId && <p className="text-xs text-slate-500">Tax ID: {b.taxId}</p>}
                      {b.email && <p className="text-xs text-slate-500">{b.email}</p>}
                    </div>
                  )}
                  {b.address && (
                    <p className="mt-2 text-xs text-slate-400 line-clamp-2">{b.address}</p>
                  )}
                  {b.tagIds && b.tagIds.length > 0 && (
                    <div className="mt-2 flex flex-wrap gap-1">
                      {b.tagIds
                        .map(id => tags.find(t => t.id === id))
                        .filter((t): t is Tag => t !== undefined)
                        .map(t => (
                          <span key={t.id} className="text-xs bg-slate-100 text-slate-600 px-2 py-0.5 rounded-full"># {t.name}</span>
                        ))}
                    </div>
                  )}
                </div>
                <div className="flex items-center gap-1 shrink-0">
                  <EditButton onClick={() => { setEditing(b); setModalOpen(true); }} />
                  <DeleteButton onClick={() => handleDelete(b)} />
                </div>
              </div>
            </Card>
          )}
        />
      )}
      {modalOpen && (
        <BusinessForm initial={editing} tags={tags} onSave={handleSave} onClose={() => setModalOpen(false)} />
      )}
      {dialog}
    </div>
  );
}
