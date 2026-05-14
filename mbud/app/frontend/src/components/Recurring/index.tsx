import { useEffect, useState } from 'react';
import { Loader2 } from 'lucide-react';
import { createRecurring, deleteRecurring, listBusiness, listRecurring, listTag, listUser, updateRecurring, type Business, type RecurringInvoice, type Tag, type User } from '../../api';
import { formatDate, formatMoney } from '../../lib/format';
import { AddButton, DeleteButton, EditButton } from '../../common/Buttons';
import { Pagination } from '../../common/Pagination';
import { useConfirm } from '../../common/ConfirmationModal';
import { RecurringForm } from './RecurringForm';

const PAGE_SIZE = 10;

const WEEKDAY_NAMES = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'];
const MONTH_NAMES = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'];

function scheduleLabel(r: RecurringInvoice): string {
  switch (r.frequency) {
    case 'daily':
      return 'Daily';
    case 'weekly':
      return r.issueDayOfWeek
        ? `Weekly on ${WEEKDAY_NAMES[r.issueDayOfWeek - 1]}`
        : 'Weekly';
    case 'monthly':
      return r.issueDayOfMonth
        ? `Monthly on day ${r.issueDayOfMonth}`
        : 'Monthly';
    case 'yearly':
      return r.issueMonthOfYear && r.issueDayOfMonth
        ? `Yearly on ${MONTH_NAMES[r.issueMonthOfYear - 1]} ${r.issueDayOfMonth}`
        : 'Yearly';
  }
}

export function Recurring() {
  const [items, setItems] = useState<RecurringInvoice[]>([]);
  const [businesses, setBusinesses] = useState<Business[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [editing, setEditing] = useState<RecurringInvoice | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [page, setPage] = useState(1);
  const { confirm, dialog } = useConfirm();

  const refresh = async () => {
    setLoading(true);
    setError(null);
    try {
      const [rec, biz, usrs, tgs] = await Promise.all([listRecurring(), listBusiness(), listUser(), listTag()]);
      setItems(rec ?? []);
      setBusinesses(biz ?? []);
      setUsers(usrs ?? []);
      setTags(tgs ?? []);
      setPage(1);
    } catch (err) {
      console.error('list_recurring failed', err);
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { void refresh(); }, []);

  const handleSave = async (payload: Omit<RecurringInvoice, 'id' | 'createdAt' | 'updatedAt'>) => {
    if (editing) {
      await updateRecurring(editing.id, { ...editing, ...payload });
    } else {
      await createRecurring(payload);
    }
    await refresh();
  };

  const handleDelete = async (r: RecurringInvoice) => {
    if (!await confirm('Delete this recurring invoice?')) return;
    try {
      await deleteRecurring(r.id);
      await refresh();
    } catch (err) {
      console.error('delete_recurring failed', err);
      setError(err instanceof Error ? err.message : String(err));
    }
  };

  const businessName = (id: string) => businesses.find(b => b.id === id)?.name;
  const totalPages = Math.ceil(items.length / PAGE_SIZE);
  const visible = items.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-slate-900">Recurring invoices</h2>
        <AddButton onClick={() => { setEditing(null); setModalOpen(true); }} />
      </div>
      {error && <div className="text-sm text-red-700 bg-red-50 border border-red-200 rounded-md px-3 py-2">{error}</div>}
      {loading ? (
        <div className="flex items-center gap-2 text-sm text-slate-500"><Loader2 className="w-4 h-4 animate-spin" />Loading...</div>
      ) : items.length === 0 ? (
        <div className="text-sm text-slate-500">No recurring invoices yet.</div>
      ) : (
        <>
          <div className="grid gap-3">
            {visible.map(r => {
              const name = businessName(r.businessId);
              return (
                <div key={r.id} className="bg-white rounded-xl shadow-sm border border-slate-100 p-4 flex items-center gap-4">
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2 flex-wrap">
                      {r.businessId === ''
                        ? <span className="font-medium text-slate-500 italic">(No business)</span>
                        : name
                          ? <span className="font-semibold text-slate-900">{name}</span>
                          : <span className="font-semibold text-red-600">&lt;unknown business&gt;</span>}
                      <span className="text-slate-900 font-medium">{formatMoney(r.amount, r.currency)}</span>
                      <span className="text-xs bg-slate-100 text-slate-600 px-2 py-0.5 rounded-full">{scheduleLabel(r)}</span>
                      {!r.active && <span className="text-xs bg-slate-200 text-slate-500 px-2 py-0.5 rounded-full">Inactive</span>}
                    </div>
                    <div className="text-xs text-slate-500 mt-1">
                      Starts {formatDate(r.startAt)}{r.endAt ? ` · ends ${formatDate(r.endAt)}` : ''}{r.description && <> · {r.description}</>}
                      {r.userIds && r.userIds.length > 0 && (
                        <> · with {r.userIds.map(id => users.find(u => u.id === id)?.name).filter(Boolean).join(', ')}</>
                      )}
                      {r.tagIds && r.tagIds.length > 0 && (
                        <> · {r.tagIds.map(id => tags.find(t => t.id === id)?.name).filter(Boolean).map(n => `#${n}`).join(', ')}</>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-1">
                    <EditButton onClick={() => { setEditing(r); setModalOpen(true); }} />
                    <DeleteButton onClick={() => handleDelete(r)} />
                  </div>
                </div>
              );
            })}
          </div>
          <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
        </>
      )}
      {modalOpen && (
        <RecurringForm initial={editing} businesses={businesses} users={users} tags={tags} onSave={handleSave} onClose={() => setModalOpen(false)} />
      )}
      {dialog}
    </div>
  );
}
