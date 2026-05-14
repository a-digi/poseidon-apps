export function formatMoney(amount: number, currency: string): string {
  try {
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency,
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(amount);
  } catch {
    return `${amount.toFixed(2)} ${currency}`;
  }
}

export function formatDate(epochSec: number): string {
  return new Date(epochSec * 1000).toLocaleDateString();
}

export function epochFromDateInput(s: string): number {
  if (!s) return 0;
  const [y, m, d] = s.split('-').map(Number);
  if (!y || !m || !d) return 0;
  return Math.floor(Date.UTC(y, m - 1, d) / 1000);
}

export function dateInputFromEpoch(e: number): string {
  if (e === 0) return '';
  return new Date(e * 1000).toISOString().slice(0, 10);
}
