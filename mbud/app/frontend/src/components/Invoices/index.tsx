import { useEffect, useState } from 'react';
import { Loader2, X } from 'lucide-react';
import { createInvoice, deleteInvoice, getInvoiceStats, listBusiness, listInvoice, listTag, listUser, updateInvoice, type Business, type CurrencyStats, type Invoice, type InvoiceSortBy, type InvoiceSortDir, type PendingInvoice, type Tag, type User } from '../../api';
import { epochFromDateInput, formatDate, formatMoney } from '../../lib/format';
import { AddButton, DeleteButton, EditButton } from '../../common/Buttons';
import { BusinessLogo } from '../Businesses/BusinessLogo';
import { Pagination } from '../../common/Pagination';
import { Limit } from '../../common/Limit';
import { Dropdown } from '../../common/Dropdown';
import { useConfirm } from '../../common/ConfirmationModal';
import { InvoiceForm } from './InvoiceForm';
import { InvoiceStatsPanel } from './InvoiceStatsPanel';

const INVOICE_LIMITS = [5, 10, 20, 50] as const;

const SORT_OPTIONS: { value: string; label: string }[] = [
  { value: 'issuedAt:desc', label: 'Issued date — newest first' },
  { value: 'issuedAt:asc', label: 'Issued date — oldest first' },
  { value: 'dueAt:asc', label: 'Due date — soonest first' },
  { value: 'dueAt:desc', label: 'Due date — latest first' },
  { value: 'amount:desc', label: 'Amount — high to low' },
  { value: 'amount:asc', label: 'Amount — low to high' },
];

type FilterPreset = 'week' | 'month' | 'year' | 'custom';

interface Range { from: number; to: number; error: string | null; }

function endOfDayEpoch(dateStr: string): number {
  const [yy, mm, dd] = dateStr.split('-').map(Number);
  return Math.floor(new Date(yy, mm - 1, dd, 23, 59, 59).getTime() / 1000);
}

function rangeOf(preset: FilterPreset, fromStr: string, toStr: string): Range {
  if (preset === 'custom') {
    const from = fromStr ? epochFromDateInput(fromStr) : 0;
    const to = toStr ? endOfDayEpoch(toStr) : 0;
    if (from > 0 && to > 0 && from > to) {
      return { from: 0, to: 0, error: 'From must be on or before To' };
    }
    return { from, to, error: null };
  }
  const now = new Date();
  const y = now.getFullYear();
  const m = now.getMonth();
  if (preset === 'week') {
    const day = now.getDay();
    const isoIdx = day === 0 ? 6 : day - 1;
    const monday = new Date(now);
    monday.setDate(now.getDate() - isoIdx);
    monday.setHours(0, 0, 0, 0);
    const sunday = new Date(monday);
    sunday.setDate(monday.getDate() + 6);
    sunday.setHours(23, 59, 59, 0);
    return {
      from: Math.floor(monday.getTime() / 1000),
      to: Math.floor(sunday.getTime() / 1000),
      error: null,
    };
  }
  if (preset === 'month') {
    const start = new Date(y, m, 1, 0, 0, 0).getTime();
    const end = new Date(y, m + 1, 0, 23, 59, 59).getTime();
    return { from: Math.floor(start / 1000), to: Math.floor(end / 1000), error: null };
  }
  const start = new Date(y, 0, 1, 0, 0, 0).getTime();
  const end = new Date(y, 11, 31, 23, 59, 59).getTime();
  return { from: Math.floor(start / 1000), to: Math.floor(end / 1000), error: null };
}

const PRESETS: { value: FilterPreset; label: string }[] = [
  { value: 'week', label: 'This week' },
  { value: 'month', label: 'This month' },
  { value: 'year', label: 'This year' },
  { value: 'custom', label: 'Custom' },
];

const dateInputClass = 'w-full px-3 py-2 rounded-md border border-slate-200 bg-white text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-500';

export function Invoices() {
  const [items, setItems] = useState<Invoice[]>([]);
  const [businesses, setBusinesses] = useState<Business[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [editing, setEditing] = useState<Invoice | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [page, setPage] = useState<number>(1);
  const [limit, setLimit] = useState<number>(5);
  const [total, setTotal] = useState<number>(0);
  const [preset, setPreset] = useState<FilterPreset>('month');
  const [customFrom, setCustomFrom] = useState<string>('');
  const [customTo, setCustomTo] = useState<string>('');
  const [customError, setCustomError] = useState<string | null>(null);
  const [sort, setSort] = useState<string>('issuedAt:desc');
  const [stats, setStats] = useState<CurrencyStats[]>([]);
  const [businessIds, setBusinessIds] = useState<string[]>([]);
  const [availableBusinessIds, setAvailableBusinessIds] = useState<string[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [userIds, setUserIds] = useState<string[]>([]);
  const [availableUserIds, setAvailableUserIds] = useState<string[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [tagIds, setTagIds] = useState<string[]>([]);
  const [availableTagIds, setAvailableTagIds] = useState<string[]>([]);
  const [unpaidOnly, setUnpaidOnly] = useState(false);
  const [pendingItems, setPendingItems] = useState<PendingInvoice[]>([]);
  const { confirm, dialog } = useConfirm();

  const refresh = async (from: number, to: number, bizIds: string[], uIds: string[], tIds: string[], unpaid: boolean, p: number, l: number, s: string) => {
    setLoading(true);
    setError(null);
    try {
      const offset = (p - 1) * l;
      const [sortBy, sortDir] = s.split(':') as [InvoiceSortBy, InvoiceSortDir];
      const [inv, biz, usrs, tgs] = await Promise.all([
        listInvoice(from, to, bizIds, uIds, tIds, unpaid, l, offset, sortBy, sortDir),
        listBusiness(),
        listUser(),
        listTag(),
      ]);
      setItems(inv.items ?? []);
      setTotal(inv.total ?? 0);
      setAvailableBusinessIds(inv.availableBusinessIds ?? []);
      setBusinesses(biz ?? []);
      setUsers(usrs ?? []);
      setAvailableUserIds(inv.availableUserIds ?? []);
      setTags(tgs ?? []);
      setAvailableTagIds(inv.availableTagIds ?? []);
      setPendingItems(inv.pendingItems ?? []);
    } catch (err) {
      console.error('list_invoice failed', err);
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  };

  const refreshCurrent = async () => {
    const r = rangeOf(preset, customFrom, customTo);
    setCustomError(r.error);
    if (r.error) return;
    await Promise.all([
      refresh(r.from, r.to, businessIds, userIds, tagIds, unpaidOnly, page, limit, sort),
      getInvoiceStats(r.from, r.to, businessIds, userIds, tagIds)
        .then(res => setStats(res.stats ?? []))
        .catch(err => {
          console.error('invoice_stats failed', err);
          setStats([]);
        }),
    ]);
  };

  useEffect(() => {
    const r = rangeOf(preset, customFrom, customTo);
    setCustomError(r.error);
    if (r.error) return;
    void refresh(r.from, r.to, businessIds, userIds, tagIds, unpaidOnly, page, limit, sort);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [preset, customFrom, customTo, businessIds, userIds, tagIds, unpaidOnly, page, limit, sort]);

  useEffect(() => {
    const r = rangeOf(preset, customFrom, customTo);
    if (r.error) return;
    getInvoiceStats(r.from, r.to, businessIds, userIds, tagIds)
      .then(res => setStats(res.stats ?? []))
      .catch(err => {
        console.error('invoice_stats failed', err);
        setStats([]);
      });
  }, [preset, customFrom, customTo, businessIds, userIds, tagIds, unpaidOnly]);

  const onPreset = (p: FilterPreset) => { setPreset(p); setPage(1); };
  const onCustomFrom = (v: string) => { setCustomFrom(v); setPage(1); };
  const onCustomTo = (v: string) => { setCustomTo(v); setPage(1); };
  const onLimit = (n: number) => { setLimit(n); setPage(1); };
  const onSort = (v: string) => { setSort(v); setPage(1); };
  const onBusinessIds = (ids: string[]) => { setBusinessIds(ids); setPage(1); };
  const onUserIds = (ids: string[]) => { setUserIds(ids); setPage(1); };
  const onTagIds = (ids: string[]) => { setTagIds(ids); setPage(1); };
  const onUnpaidOnly = (v: boolean) => { setUnpaidOnly(v); setPage(1); };

  const handleSave = async (payload: Omit<Invoice, 'id' | 'createdAt' | 'updatedAt'>) => {
    if (editing) {
      await updateInvoice(editing.id, { ...editing, ...payload });
    } else {
      await createInvoice(payload);
    }
    await refreshCurrent();
  };

  const handleDelete = async (i: Invoice) => {
    const linkedUpcomingCount = i.upcomingIds?.length ?? 0;
    const message = linkedUpcomingCount === 0
      ? 'Delete this invoice?'
      : linkedUpcomingCount === 1
        ? 'This invoice is linked to an upcoming reminder. Deleting it will also remove the reminder. Continue?'
        : `This invoice is linked to ${linkedUpcomingCount} upcoming reminders. Deleting it will also remove them. Continue?`;
    if (!await confirm(message)) return;
    setLoading(true);
    setError(null);
    try {
      await deleteInvoice(i.id);
      // Clamp page if the just-deleted item was the last on the current page.
      // `total` here is the pre-delete count; subtract 1 to get the new effective total.
      const newTotal = Math.max(0, total - 1);
      const newTotalPages = Math.max(1, Math.ceil(newTotal / limit));
      if (page > newTotalPages) {
        // setPage triggers the existing useEffect which will re-fetch the correct page.
        // Return early to avoid a double-fetch racing with that effect.
        setPage(newTotalPages);
        return;
      }
      await refreshCurrent();
    } catch (err) {
      console.error('delete_invoice failed', err);
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  };

  const businessOf = (id: string) => businesses.find(b => b.id === id);
  const totalPages = Math.max(1, Math.ceil(total / limit));
  const safePage = Math.min(page, totalPages);
  const businessOptions = availableBusinessIds
    .map(id => businesses.find(b => b.id === id))
    .filter((b): b is Business => b !== undefined)
    .map(b => ({ value: b.id, label: b.name }));
  const activeFilterCount = (preset !== 'month' ? 1 : 0) + (businessIds.length > 0 ? 1 : 0) + (userIds.length > 0 ? 1 : 0) + (tagIds.length > 0 ? 1 : 0) + (unpaidOnly ? 1 : 0);

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-slate-900">Invoices</h2>
        <AddButton onClick={() => { setEditing(null); setModalOpen(true); }} />
      </div>
      <div className="flex flex-col gap-2">
        <div className="flex items-center gap-2 flex-wrap">
          <div className="inline-flex rounded-lg bg-slate-100 p-0.5 gap-0.5">
            {PRESETS.map(p => {
              const active = preset === p.value;
              return (
                <button
                  key={p.value}
                  type="button"
                  onClick={() => onPreset(p.value)}
                  className={`px-2.5 py-1.5 rounded-md text-xs font-medium transition-all duration-150 focus-visible:outline-none whitespace-nowrap ${active
                    ? 'bg-white text-slate-900 shadow-sm'
                    : 'text-slate-500 hover:text-slate-700'
                    }`}
                >
                  {p.label}
                </button>
              );
            })}
          </div>
          {availableBusinessIds.length > 0 && (
            <div className="w-44">
              <Dropdown
                mode="multi"
                options={businessOptions}
                value={businessIds}
                onChange={onBusinessIds}
                placeholder="All businesses"
              />
            </div>
          )}
          {availableUserIds.length > 0 && (
            <div className="w-44">
              <Dropdown
                mode="multi"
                options={availableUserIds
                  .map(id => users.find(u => u.id === id))
                  .filter((u): u is User => u !== undefined)
                  .map(u => ({ value: u.id, label: u.name }))}
                value={userIds}
                onChange={onUserIds}
                placeholder="All users"
              />
            </div>
          )}
          {availableTagIds.length > 0 && (
            <div className="w-44">
              <Dropdown
                mode="multi"
                options={availableTagIds
                  .map(id => tags.find(t => t.id === id))
                  .filter((t): t is Tag => t !== undefined)
                  .map(t => ({ value: t.id, label: t.name }))}
                value={tagIds}
                onChange={onTagIds}
                placeholder="All tags"
              />
            </div>
          )}
          <button
            type="button"
            onClick={() => onUnpaidOnly(!unpaidOnly)}
            className={`px-3 py-1.5 rounded-full text-xs font-medium transition-all duration-150 focus-visible:outline-none whitespace-nowrap ${unpaidOnly
              ? 'bg-slate-900 text-white shadow-sm'
              : 'bg-slate-100 text-slate-700 hover:bg-slate-200'
              }`}
          >
            Still to be paid
          </button>
          {activeFilterCount > 0 && (
            <button
              type="button"
              onClick={() => { setPreset('month'); setBusinessIds([]); setUserIds([]); setTagIds([]); setUnpaidOnly(false); setPage(1); }}
              className="text-xs text-slate-400 hover:text-slate-700 transition-colors whitespace-nowrap"
            >
              Reset
            </button>
          )}
        </div>

        {activeFilterCount > 0 && (
          <div className="flex items-center gap-2 flex-wrap bg-gray-50 rounded-lg px-3 py-2">
            {preset !== 'month' && (
              <span className="inline-flex items-center gap-1.5 pl-2.5 pr-1.5 py-1 rounded-md bg-slate-100 text-slate-700 text-xs font-medium border border-slate-200">
                {PRESETS.find(p => p.value === preset)?.label}
                <button
                  type="button"
                  aria-label="Remove date filter"
                  onClick={() => { setPreset('month'); setPage(1); }}
                  className="text-slate-400 hover:text-slate-700 transition-colors"
                >
                  <X className="w-3 h-3" />
                </button>
              </span>
            )}
            {businessIds.map(id => {
              const biz = businesses.find(b => b.id === id);
              return biz ? (
                <span key={id} className="inline-flex items-center gap-1.5 pl-2.5 pr-1.5 py-1 rounded-md bg-slate-100 text-slate-700 text-xs font-medium border border-slate-200">
                  {biz.name}
                  <button
                    type="button"
                    aria-label={`Remove ${biz.name} filter`}
                    onClick={() => onBusinessIds(businessIds.filter(x => x !== id))}
                    className="text-slate-400 hover:text-slate-700 transition-colors"
                  >
                    <X className="w-3 h-3" />
                  </button>
                </span>
              ) : null;
            })}
            {userIds.map(id => {
              const usr = users.find(u => u.id === id);
              return usr ? (
                <span key={id} className="inline-flex items-center gap-1.5 pl-2.5 pr-1.5 py-1 rounded-md bg-slate-100 text-slate-700 text-xs font-medium border border-slate-200">
                  {usr.name}
                  <button
                    type="button"
                    aria-label={`Remove ${usr.name} filter`}
                    onClick={() => onUserIds(userIds.filter(x => x !== id))}
                    className="text-slate-400 hover:text-slate-700 transition-colors"
                  >
                    <X className="w-3 h-3" />
                  </button>
                </span>
              ) : null;
            })}
            {tagIds.map(id => {
              const tg = tags.find(t => t.id === id);
              return tg ? (
                <span key={id} className="inline-flex items-center gap-1.5 pl-2.5 pr-1.5 py-1 rounded-md bg-slate-100 text-slate-700 text-xs font-medium border border-slate-200">
                  #{tg.name}
                  <button
                    type="button"
                    aria-label={`Remove ${tg.name} filter`}
                    onClick={() => onTagIds(tagIds.filter(x => x !== id))}
                    className="text-slate-400 hover:text-slate-700 transition-colors"
                  >
                    <X className="w-3 h-3" />
                  </button>
                </span>
              ) : null;
            })}
            {unpaidOnly && (
              <span className="inline-flex items-center gap-1.5 pl-2.5 pr-1.5 py-1 rounded-md bg-slate-100 text-slate-700 text-xs font-medium border border-slate-200">
                Still to be paid
                <button
                  type="button"
                  aria-label="Clear unpaid filter"
                  onClick={() => onUnpaidOnly(false)}
                  className="text-slate-400 hover:text-slate-700 transition-colors"
                >
                  <X className="w-3 h-3" />
                </button>
              </span>
            )}
          </div>
        )}

        {preset === 'custom' && (
          <div className="flex items-end gap-3 flex-wrap pl-1">
            <label className="flex flex-col gap-1 text-xs font-medium text-slate-600">
              From
              <input
                type="date"
                value={customFrom}
                onChange={e => onCustomFrom(e.target.value)}
                className={dateInputClass}
              />
            </label>
            <label className="flex flex-col gap-1 text-xs font-medium text-slate-600">
              To
              <input
                type="date"
                value={customTo}
                onChange={e => onCustomTo(e.target.value)}
                className={dateInputClass}
              />
            </label>
            {customError && <span className="text-xs text-red-600 self-center">{customError}</span>}
          </div>
        )}
      </div>

      <div className="grid mt-7 grid-cols-1 lg:grid-cols-[1fr_22rem] gap-6 items-start">
        <div className="flex flex-col gap-4">
          {total > 0 && (
            <div className="flex items-center gap-3 flex-wrap">
              <div className="min-w-[14rem]">
                <Dropdown
                  mode="single"
                  options={SORT_OPTIONS}
                  value={sort}
                  onChange={onSort}
                  placeholder="Sort by"
                />
              </div>
              <span className="text-sm text-slate-400 tabular-nums">
                {total} {total === 1 ? 'invoice' : 'invoices'}
              </span>
              <div className="ml-auto">
                <Limit value={limit} onChange={onLimit} options={INVOICE_LIMITS} />
              </div>
            </div>
          )}
          {error && <div className="text-sm text-red-700 bg-red-50 border border-red-200 rounded-md px-3 py-2">{error}</div>}
          {loading ? (
            <div className="flex items-center gap-2 text-sm text-slate-500"><Loader2 className="w-4 h-4 animate-spin" />Loading...</div>
          ) : items.length === 0 ? (
            <div className="text-sm text-slate-500">No invoices yet.</div>
          ) : (
            <>
              <div className="grid gap-3">
                {items.map(i => {
                  const biz = businessOf(i.businessId);
                  const name = biz?.name;
                  return (
                    <div key={i.id} className="bg-white rounded-xl shadow-sm border border-slate-100 p-4 flex items-center gap-4">
                      {i.businessId !== '' && (
                        <BusinessLogo name={name ?? '?'} logo={biz?.logo} size={40} />
                      )}
                      <div className="min-w-0 flex-1">
                        <div className="flex items-center gap-2 flex-wrap">
                          {i.businessId === ''
                            ? <span className="font-medium text-slate-500 italic">(No business)</span>
                            : name
                              ? <span className="font-semibold text-slate-900">{name}</span>
                              : <span className="font-semibold text-red-600">&lt;unknown business&gt;</span>}
                          <span className="text-slate-900 font-medium">{formatMoney(i.amount, i.currency)}</span>
                          {((i.recurringIds && i.recurringIds.length > 0) || (i.upcomingIds && i.upcomingIds.length > 0)) && (
                            <span className="text-xs font-semibold uppercase tracking-wide bg-amber-100 text-amber-800 px-2 py-0.5 rounded-full border border-amber-200">
                              Auto-generated
                            </span>
                          )}
                          {i.paid
                            ? <span className="text-xs bg-green-100 text-green-700 px-2 py-0.5 rounded-full">Paid</span>
                            : <span className="text-xs bg-slate-100 text-slate-600 px-2 py-0.5 rounded-full">Unpaid</span>}
                        </div>
                        <div className="text-xs text-slate-500 mt-1">
                          Due {formatDate(i.dueAt)}{i.description && <> · {i.description}</>}
                          {i.userIds && i.userIds.length > 0 && (
                            <> · with {i.userIds.map(id => users.find(u => u.id === id)?.name).filter(Boolean).join(', ')}</>
                          )}
                          {i.tagIds && i.tagIds.length > 0 && (
                            <> · {i.tagIds.map(id => tags.find(t => t.id === id)?.name).filter(Boolean).map(n => `#${n}`).join(', ')}</>
                          )}
                        </div>
                      </div>
                      <div className="flex items-center gap-1">
                        <EditButton onClick={() => { setEditing(i); setModalOpen(true); }} />
                        <DeleteButton onClick={() => handleDelete(i)} />
                      </div>
                    </div>
                  );
                })}
              </div>
              <Pagination page={safePage} totalPages={totalPages} onPageChange={setPage} />
            </>
          )}
          {pendingItems.length > 0 && (
            <section className="flex flex-col gap-3 mt-2">
              <h3 className="text-xs font-semibold text-slate-400 uppercase tracking-wider">
                Still to be paid ({pendingItems.length})
              </h3>
              <div className="grid gap-3 opacity-60">
                {pendingItems.map(p => {
                  const biz = businesses.find(b => b.id === p.businessId);
                  const bizName = biz?.name;
                  const tagNames = (p.tagIds ?? []).map(id => tags.find(t => t.id === id)?.name).filter(Boolean);
                  const userNames = (p.userIds ?? []).map(id => users.find(u => u.id === id)?.name).filter(Boolean);
                  return (
                    <div key={`${p.source}-${p.sourceId}-${p.dueAt}`} className="bg-white rounded-xl shadow-sm border border-dashed border-slate-200 p-4 flex items-center gap-4">
                      {p.businessId !== '' && <BusinessLogo name={bizName ?? '?'} logo={biz?.logo} size={40} />}
                      <div className="min-w-0 flex-1">
                        <div className="flex items-center gap-2 flex-wrap">
                          {p.businessId === ''
                            ? <span className="font-medium text-slate-500 italic">(No business)</span>
                            : bizName
                              ? <span className="font-semibold text-slate-900">{bizName}</span>
                              : <span className="font-semibold text-red-600">&lt;unknown business&gt;</span>}
                          <span className="text-slate-900 font-medium">{formatMoney(p.amount, p.currency)}</span>
                          <span className="text-xs font-semibold uppercase tracking-wide bg-slate-100 text-slate-600 px-2 py-0.5 rounded-full border border-slate-200">
                            Pending
                          </span>
                          <span className="text-xs bg-slate-100 text-slate-600 px-2 py-0.5 rounded-full">
                            From {p.source}
                          </span>
                        </div>
                        <div className="text-xs text-slate-500 mt-1">
                          Due {formatDate(p.dueAt)}{p.description && <> · {p.description}</>}
                          {userNames.length > 0 && <> · with {userNames.join(', ')}</>}
                          {tagNames.length > 0 && <> · {tagNames.map(n => `#${n}`).join(', ')}</>}
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            </section>
          )}
        </div>
        <InvoiceStatsPanel stats={stats} businesses={businesses} />
      </div>
      {modalOpen && (
        <InvoiceForm initial={editing} businesses={businesses} users={users} tags={tags} onSave={handleSave} onClose={() => setModalOpen(false)} />
      )}
      {dialog}
    </div>
  );
}
