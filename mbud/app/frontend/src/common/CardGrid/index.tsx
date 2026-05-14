import { useState } from 'react';
import { Limit } from '../Limit';
import { Pagination } from '../Pagination';

export interface CardProps {
  children: React.ReactNode;
}

export function Card({ children }: CardProps) {
  return (
    <div className="relative bg-white rounded-xl border border-slate-100 shadow-sm hover:shadow-xl hover:-translate-y-1 transition-all duration-200 overflow-hidden group">
      <div
        aria-hidden={true}
        className="absolute left-0 top-0 bottom-0 w-1 bg-gradient-to-b from-slate-500 to-slate-900 pointer-events-none"
      />
      <div
        aria-hidden={true}
        className="absolute left-0 right-0 top-0 h-px bg-gradient-to-r from-white via-slate-200 to-white pointer-events-none"
      />
      <div className="pl-5 pr-4 py-4">{children}</div>
    </div>
  );
}

interface CardGridProps<T> {
  items: T[];
  renderCard: (item: T) => React.ReactNode;
  defaultLimit?: number;
  columns?: 3 | 4;
  emptyMessage?: string;
}

export function CardGrid<T>({
  items,
  renderCard,
  defaultLimit = 12,
  columns = 3,
  emptyMessage = 'No items.',
}: CardGridProps<T>) {
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(defaultLimit);

  const totalPages = Math.max(1, Math.ceil(items.length / limit));
  const safePage = Math.min(page, totalPages);
  const visible = items.slice((safePage - 1) * limit, safePage * limit);

  const handleLimitChange = (newLimit: number) => {
    setLimit(newLimit);
    setPage(1);
  };

  const gridCls =
    columns === 4
      ? 'grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4'
      : 'grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4';

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <span className="text-sm text-slate-400 tabular-nums">
          {items.length} {items.length === 1 ? 'item' : 'items'}
        </span>
        <Limit value={limit} onChange={handleLimitChange} />
      </div>
      {items.length === 0 ? (
        <div className="py-16 text-center text-sm text-slate-400">{emptyMessage}</div>
      ) : (
        <>
          <div className={gridCls}>
            {visible.map((item, i) => (
              <div key={i}>{renderCard(item)}</div>
            ))}
          </div>
          <Pagination page={safePage} totalPages={totalPages} onPageChange={setPage} />
        </>
      )}
    </div>
  );
}
