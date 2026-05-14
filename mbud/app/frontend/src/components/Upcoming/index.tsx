import { useEffect, useState } from 'react';
import { Loader2 } from 'lucide-react';
import { createUpcoming, deleteUpcoming, listBusiness, listTag, listUpcoming, listUser, updateUpcoming, type Business, type Tag, type UpcomingInvoice, type User } from '../../api';
import { formatDate, formatMoney } from '../../lib/format';
import { AddButton, DeleteButton, EditButton } from '../../common/Buttons';
import { Pagination } from '../../common/Pagination';
import { useConfirm } from '../../common/ConfirmationModal';
import { UpcomingForm } from './UpcomingForm';

const PAGE_SIZE = 10;

export function Upcoming() {
  const [items, setItems] = useState<UpcomingInvoice[]>([]);
  const [businesses, setBusinesses] = useState<Business[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [editing, setEditing] = useState<UpcomingInvoice | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [page, setPage] = useState(1);
  const { confirm, dialog } = useConfirm();

  const refresh = async () => {
    setLoading(true);
    setError(null);
    try {
      const [up, biz, usrs, tgs] = await Promise.all([listUpcoming(), listBusiness(), listUser(), listTag()]);
      setItems(up ?? []);
      setBusinesses(biz ?? []);
      setUsers(usrs ?? []);
      setTags(tgs ?? []);
      setPage(1);
    } catch (err) {
      console.error('list_upcoming failed', err);
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { void refresh(); }, []);

  const handleSave = async (payload: Omit<UpcomingInvoice, 'id' | 'createdAt' | 'updatedAt'>) => {
    if (editing) {
      await updateUpcoming(editing.id, { ...editing, ...payload });
    } else {
      await createUpcoming(payload);
    }
    await refresh();
  };

  const handleDelete = async (u: UpcomingInvoice) => {
    if (!await confirm('Delete this upcoming invoice?')) return;
    try {
      await deleteUpcoming(u.id);
      await refresh();
    } catch (err) {
      console.error('delete_upcoming failed', err);
      setError(err instanceof Error ? err.message : String(err));
    }
  };

  const businessName = (id: string) => businesses.find(b => b.id === id)?.name;
  const totalPages = Math.ceil(items.length / PAGE_SIZE);
  const visible = items.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-slate-900">Upcoming invoices</h2>
        <AddButton onClick={() => { setEditing(null); setModalOpen(true); }} />
      </div>
      {error && <div className="text-sm text-red-700 bg-red-50 border border-red-200 rounded-md px-3 py-2">{error}</div>}
      {loading ? (
        <div className="flex items-center gap-2 text-sm text-slate-500"><Loader2 className="w-4 h-4 animate-spin" />Loading...</div>
      ) : items.length === 0 ? (
        <div className="text-sm text-slate-500">No upcoming invoices yet.</div>
      ) : (
        <>
          <div className="grid gap-3">
            {visible.map(u => {
              const name = businessName(u.businessId);
              return (
                <div key={u.id} className="bg-white rounded-xl shadow-sm border border-slate-100 p-4 flex items-center gap-4">
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2 flex-wrap">
                      {u.businessId === ''
                        ? <span className="font-medium text-slate-500 italic">(No business)</span>
                        : name
                          ? <span className="font-semibold text-slate-900">{name}</span>
                          : <span className="font-semibold text-red-600">&lt;unknown business&gt;</span>}
                      <span className="text-slate-900 font-medium">{formatMoney(u.amount, u.currency)}</span>
                    </div>
                    <div className="text-xs text-slate-500 mt-1">
                      Due {formatDate(u.dueAt)}{u.description && <> · {u.description}</>}
                      {u.userIds && u.userIds.length > 0 && (
                        <> · with {u.userIds.map(id => users.find(usr => usr.id === id)?.name).filter(Boolean).join(', ')}</>
                      )}
                      {u.tagIds && u.tagIds.length > 0 && (
                        <> · {u.tagIds.map(id => tags.find(t => t.id === id)?.name).filter(Boolean).map(n => `#${n}`).join(', ')}</>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-1">
                    <EditButton onClick={() => { setEditing(u); setModalOpen(true); }} />
                    <DeleteButton onClick={() => handleDelete(u)} />
                  </div>
                </div>
              );
            })}
          </div>
          <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
        </>
      )}
      {modalOpen && (
        <UpcomingForm initial={editing} businesses={businesses} users={users} tags={tags} onSave={handleSave} onClose={() => setModalOpen(false)} />
      )}
      {dialog}
    </div>
  );
}
