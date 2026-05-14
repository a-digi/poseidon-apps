const params = new URLSearchParams(window.location.search);
const PLUGIN_ID = params.get('pluginId') ?? 'mbud';
const BACKEND_URL = (params.get('backendUrl') ?? window.location.origin).replace(/\/$/, '');

function mobileTokenHeader(): Record<string, string> {
  if (typeof window === 'undefined') return {};
  const host = window.location.hostname;
  if (host === 'localhost' || host === '127.0.0.1' || host === '::1') return {};
  try {
    const t = window.localStorage.getItem('mbud_mobile_token');
    return t ? { 'X-Mobile-Token': t } : {};
  } catch { return {}; }
}

export async function callPlugin<T>(action: string, params: Record<string, unknown> = {}): Promise<T> {
  const res = await fetch(`${BACKEND_URL}/api/plugins/execute`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...mobileTokenHeader() },
    body: JSON.stringify({ pluginName: PLUGIN_ID, params: { action, ...params } }),
  });
  if (!res.ok) throw new Error(`HTTP ${res.status}: ${res.statusText}`);
  const json = (await res.json()) as { result?: T; error?: string };
  if (json.error) throw new Error(json.error);
  return json.result as T;
}

export type Frequency = 'daily' | 'weekly' | 'monthly' | 'yearly';

export interface Business {
  id: string;
  name: string;
  taxId?: string;
  email?: string;
  address?: string;
  notes?: string;
  logoType?: string;
  logo?: string;
  createdAt: number;
  updatedAt: number;
  tagIds?: string[];
}

export interface User {
  id: string;
  name: string;
  email?: string;
  notes?: string;
  createdAt: number;
  updatedAt: number;
}

export interface Tag {
  id: string;
  name: string;
  createdAt: number;
  updatedAt: number;
}

function parseDataUrl(dataUrl: string): { contentType: string; base64: string } {
  const m = /^data:([^;]+);base64,(.+)$/.exec(dataUrl);
  if (!m) throw new Error('invalid data URL');
  return { contentType: m[1], base64: m[2] };
}

export interface Invoice {
  id: string;
  businessId: string;
  amount: number;
  currency: string;
  description?: string;
  issuedAt: number;
  dueAt: number;
  paid: boolean;
  paidAt?: number;
  recurringIds?: string[];
  upcomingIds?: string[];
  userIds?: string[];
  tagIds?: string[];
  createdAt: number;
  updatedAt: number;
}

export interface PendingInvoice {
  source: 'recurring' | 'upcoming';
  sourceId: string;
  businessId: string;
  amount: number;
  currency: string;
  description?: string;
  dueAt: number;
  issuedAt: number;
  tagIds?: string[];
  userIds?: string[];
}

export interface RecurringInvoice {
  id: string;
  businessId: string;
  amount: number;
  currency: string;
  description?: string;
  frequency: Frequency;
  startAt: number;
  endAt?: number;
  active: boolean;
  issueDayOfWeek?: number;     // 1..7 (ISO 8601: 1=Mon..7=Sun)
  issueDayOfMonth?: number;    // 1..31
  issueMonthOfYear?: number;   // 1..12
  createdAt: number;
  updatedAt: number;
  userIds?: string[];
  tagIds?: string[];
}

export interface UpcomingInvoice {
  id: string;
  businessId: string;
  amount: number;
  currency: string;
  description?: string;
  dueAt: number;
  createdAt: number;
  updatedAt: number;
  userIds?: string[];
  tagIds?: string[];
}

export const health = () => callPlugin<{ ok: true; version: string }>('health');

export const listBusiness = () => callPlugin<Business[]>('list_business');
export const createBusiness = (business: Omit<Business, 'id' | 'createdAt' | 'updatedAt'>) =>
  callPlugin<Business>('create_business', { business });
export const updateBusiness = (id: string, business: Business) =>
  callPlugin<Business>('update_business', { id, business });
export const deleteBusiness = (id: string) =>
  callPlugin<{ ok: true }>('delete_business', { id });

export const uploadBusinessLogo = (id: string, dataUrl: string) => {
  const { contentType, base64 } = parseDataUrl(dataUrl);
  return callPlugin<Business>('upload_business_logo', { id, dataBase64: base64, contentType });
};

export const deleteBusinessLogo = (id: string) =>
  callPlugin<Business>('delete_business_logo', { id });

export const listUser = () => callPlugin<User[]>('list_user');
export const createUser = (user: Omit<User, 'id' | 'createdAt' | 'updatedAt'>) =>
  callPlugin<User>('create_user', { user });
export const updateUser = (id: string, user: User) =>
  callPlugin<User>('update_user', { id, user });
export const deleteUser = (id: string) =>
  callPlugin<{ ok: true }>('delete_user', { id });

export const listTag = () => callPlugin<Tag[]>('list_tag');
export const createTag = (tag: Omit<Tag, 'id' | 'createdAt' | 'updatedAt'>) =>
  callPlugin<Tag>('create_tag', { tag });
export const updateTag = (id: string, tag: Tag) =>
  callPlugin<Tag>('update_tag', { id, tag });
export const deleteTag = (id: string) =>
  callPlugin<{ ok: true }>('delete_tag', { id });

export interface InvoiceListResult {
  items: Invoice[];
  total: number;
  availableBusinessIds: string[];
  availableUserIds: string[];
  availableTagIds: string[];
  pendingItems: PendingInvoice[];
}

export interface TopBusiness {
  businessId: string;
  amount: number;
}

export interface CurrencyStats {
  currency: string;
  total: number;
  count: number;
  paidAmount: number;
  paidCount: number;
  unpaidAmount: number;
  unpaidCount: number;
  average: number;
  maxAmount: number;
  businessCount: number;
  topBusinesses: TopBusiness[];
  topDayEpoch: number;
  topDayAmount: number;
  pendingAmount: number;
  pendingCount: number;
}

export interface InvoiceStatsResult {
  stats: CurrencyStats[];
}

export const getInvoiceStats = (from = 0, to = 0, businessIds: string[] = [], userIds: string[] = [], tagIds: string[] = []) =>
  callPlugin<InvoiceStatsResult>('invoice_stats', { from, to, businessIds, userIds, tagIds });

export type InvoiceSortBy  = 'issuedAt' | 'dueAt' | 'amount';
export type InvoiceSortDir = 'asc' | 'desc';

export const listInvoice = (
  from = 0,
  to   = 0,
  businessIds: string[] = [],
  userIds: string[] = [],
  tagIds: string[] = [],
  unpaidOnly = false,
  limit  = 0,
  offset = 0,
  sortBy:  InvoiceSortBy  = 'issuedAt',
  sortDir: InvoiceSortDir = 'desc',
) => callPlugin<InvoiceListResult>('list_invoice', { from, to, businessIds, userIds, tagIds, unpaidOnly, limit, offset, sortBy, sortDir });
export const createInvoice = (invoice: Omit<Invoice, 'id' | 'createdAt' | 'updatedAt'>) =>
  callPlugin<Invoice>('create_invoice', { invoice });
export const updateInvoice = (id: string, invoice: Invoice) =>
  callPlugin<Invoice>('update_invoice', { id, invoice });
export const deleteInvoice = (id: string) =>
  callPlugin<{ ok: true }>('delete_invoice', { id });

export const listRecurring = () => callPlugin<RecurringInvoice[]>('list_recurring');
export const createRecurring = (recurring: Omit<RecurringInvoice, 'id' | 'createdAt' | 'updatedAt'>) =>
  callPlugin<RecurringInvoice>('create_recurring', { recurring });
export const updateRecurring = (id: string, recurring: RecurringInvoice) =>
  callPlugin<RecurringInvoice>('update_recurring', { id, recurring });
export const deleteRecurring = (id: string) =>
  callPlugin<{ ok: true }>('delete_recurring', { id });

export const listUpcoming = () => callPlugin<UpcomingInvoice[]>('list_upcoming');
export const createUpcoming = (upcoming: Omit<UpcomingInvoice, 'id' | 'createdAt' | 'updatedAt'>) =>
  callPlugin<UpcomingInvoice>('create_upcoming', { upcoming });
export const updateUpcoming = (id: string, upcoming: UpcomingInvoice) =>
  callPlugin<UpcomingInvoice>('update_upcoming', { id, upcoming });
export const deleteUpcoming = (id: string) =>
  callPlugin<{ ok: true }>('delete_upcoming', { id });

export interface UploadSession {
  id: string;
  invoiceId?: string;
  status: 'active' | 'consumed' | 'expired';
  createdAt: number;
  expiresAt: number;
}

export interface Attachment {
  id: string;
  invoiceId?: string;
  sessionId?: string;
  mime: string;
  originalFilename: string;
  sizeBytes: number;
  createdAt: number;
}

export interface UploadSessionResult {
  session: UploadSession;
  attachments: Attachment[];
}

export const createUploadSession = (invoiceId?: string, ttlSeconds = 600) =>
  callPlugin<{ token: string; expiresAt: number }>('create_upload_session', { invoiceId, ttlSeconds });

export const getUploadSession = (token: string) =>
  callPlugin<UploadSessionResult>('get_upload_session', { token });

export const mobileUploadFile = (token: string, filename: string, contentType: string, dataBase64: string) =>
  callPlugin<Attachment>('mobile_upload_file', { token, filename, contentType, dataBase64 });

export const listInvoiceAttachments = (invoiceId: string) =>
  callPlugin<Attachment[]>('list_invoice_attachments', { invoiceId });

export const attachSessionToInvoice = (token: string, invoiceId: string) =>
  callPlugin<{ ok: true; count: number }>('attach_session_to_invoice', { token, invoiceId });

export const deleteAttachment = (id: string) =>
  callPlugin<{ ok: true }>('delete_attachment', { id });

export async function getLanAddresses(): Promise<string[]> {
  const res = await fetch(`${BACKEND_URL}/api/system/lan-addresses`, { headers: { ...mobileTokenHeader() } });
  if (!res.ok) return [];
  const json = await res.json() as { addresses?: string[] };
  return json.addresses ?? [];
}

export interface AttachmentBytes {
  id: string;
  mime: string;
  originalFilename: string;
  dataBase64: string;
}

export const getAttachmentBytes = (id: string) =>
  callPlugin<AttachmentBytes>('get_attachment_bytes', { id });

export interface MobileSession { token: string; expiresAt: number; }

export async function createMobileSession(pluginId: string, ttlSeconds = 3600): Promise<MobileSession> {
  const resp = await fetch(`${BACKEND_URL}/api/mobile-sessions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pluginId, ttlSeconds }),
  });
  if (!resp.ok) throw new Error(`Failed to create session: ${resp.statusText}`);
  return resp.json() as Promise<MobileSession>;
}

export async function revokeMobileSession(token: string): Promise<{ ok: true }> {
  const resp = await fetch(`${BACKEND_URL}/api/mobile-sessions/${encodeURIComponent(token)}`, {
    method: 'DELETE',
  });
  if (!resp.ok) throw new Error(`Failed to revoke session: ${resp.statusText}`);
  return resp.json() as Promise<{ ok: true }>;
}
