import type { Business, CurrencyStats } from '../../api';
import { formatDate, formatMoney } from '../../lib/format';

interface InvoiceStatsPanelProps {
  stats: CurrencyStats[];
  businesses: Business[];
}

export function InvoiceStatsPanel({ stats, businesses }: InvoiceStatsPanelProps) {
  if (stats.length === 0) return null;
  const businessLabel = (id: string) =>
    id === '' ? '(No business)' : businesses.find(b => b.id === id)?.name ?? '<unknown>';

  return (
    <div className="flex flex-col gap-4">
      {stats.filter(s => s.count > 0 || s.pendingCount > 0).map(s => (
        <div key={s.currency} className="bg-white rounded-xl border border-slate-100 shadow-sm p-4 flex flex-col gap-3">
          <div className="flex items-baseline justify-between">
            <span className="text-xs font-semibold uppercase tracking-wide text-slate-500">
              {s.currency || '—'}
            </span>
            <span className="text-xs text-slate-400 tabular-nums">
              {s.count} {s.count === 1 ? 'invoice' : 'invoices'} · {s.businessCount}{' '}
              {s.businessCount === 1 ? 'biz' : 'biz'}
            </span>
          </div>

          <div className="text-2xl font-semibold text-slate-900 tabular-nums">
            {formatMoney(s.total, s.currency)}
          </div>

          <div className="border-t border-slate-100 pt-2">
            <table className="w-full text-xs">
              <thead>
                <tr>
                  <th className="text-left font-normal text-slate-400 pb-1.5">Status</th>
                  <th className="text-right font-normal text-slate-400 pb-1.5">Count</th>
                  <th className="text-right font-normal text-slate-400 pb-1.5 pl-3">Amount</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td className="py-0.5 text-green-700">Paid</td>
                  <td className="py-0.5 text-right tabular-nums text-slate-400">{s.paidCount}</td>
                  <td className="py-0.5 text-right tabular-nums text-green-700 pl-3">
                    {formatMoney(s.paidAmount, s.currency)}
                  </td>
                </tr>
                <tr>
                  <td className="py-0.5 text-amber-700">Unpaid</td>
                  <td className="py-0.5 text-right tabular-nums text-slate-400">{s.unpaidCount}</td>
                  <td className="py-0.5 text-right tabular-nums text-amber-700 pl-3">
                    {formatMoney(s.unpaidAmount, s.currency)}
                  </td>
                </tr>
                {s.pendingCount > 0 && (
                  <tr>
                    <td className="py-0.5 text-slate-500">Still to be paid</td>
                    <td className="py-0.5 text-right tabular-nums text-slate-400">{s.pendingCount}</td>
                    <td className="py-0.5 text-right tabular-nums text-slate-700 pl-3">
                      {formatMoney(s.pendingAmount, s.currency)}
                    </td>
                  </tr>
                )}
                <tr className="border-t border-slate-100">
                  <td className="pt-1.5 font-medium text-slate-700">Total</td>
                  <td className="pt-1.5 text-right tabular-nums text-slate-500">{s.count}</td>
                  <td className="pt-1.5 text-right tabular-nums font-medium text-slate-900 pl-3">
                    {formatMoney(s.total, s.currency)}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          {s.topBusinesses.length > 0 && (
            <div className="border-t border-slate-100 pt-2 flex flex-col gap-2">
              <div className="text-xs text-slate-400">Top businesses</div>
              {s.topBusinesses.map(b => {
                const pct = s.total > 0 ? (b.amount / s.total) * 100 : 0;
                return (
                  <div key={b.businessId} className="flex flex-col gap-1">
                    <div className="flex items-baseline justify-between text-xs gap-2">
                      <span className="text-slate-700 truncate">{businessLabel(b.businessId)}</span>
                      <span className="whitespace-nowrap tabular-nums">
                        <span className="text-slate-700">{formatMoney(b.amount, s.currency)}</span>
                        <span className="text-slate-400 ml-2">{pct.toFixed(0)}%</span>
                      </span>
                    </div>
                    <div className="h-1.5 bg-slate-100 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-gradient-to-r from-slate-500 to-slate-900 rounded-full"
                        style={{ width: `${pct}%` }}
                        aria-hidden={true}
                      />
                    </div>
                  </div>
                );
              })}
            </div>
          )}

          <div className="border-t border-slate-100 pt-2">
            <table className="w-full text-xs">
              <tbody>
                {s.topDayEpoch > 0 && (
                  <tr>
                    <td className="py-0.5 text-slate-400 pr-2 whitespace-nowrap">Busiest day</td>
                    <td className="py-0.5 text-slate-700">{formatDate(s.topDayEpoch)}</td>
                    <td className="py-0.5 text-right tabular-nums text-slate-900 pl-3 whitespace-nowrap">
                      {formatMoney(s.topDayAmount, s.currency)}
                    </td>
                  </tr>
                )}
                <tr>
                  <td className="py-0.5 text-slate-400 pr-2">Average</td>
                  <td colSpan={2} className="py-0.5 text-right tabular-nums text-slate-700">
                    {formatMoney(s.average, s.currency)}
                  </td>
                </tr>
                <tr>
                  <td className="py-0.5 text-slate-400 pr-2">Largest</td>
                  <td colSpan={2} className="py-0.5 text-right tabular-nums text-slate-700">
                    {formatMoney(s.maxAmount, s.currency)}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      ))}
    </div>
  );
}
