interface PaginationProps {
  page: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

function pageRange(page: number, total: number): (number | null)[] {
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1);
  const lo = Math.max(2, page - 2);
  const hi = Math.min(total - 1, page + 2);
  const pages: (number | null)[] = [1];
  if (lo > 2) pages.push(null);
  for (let i = lo; i <= hi; i++) pages.push(i);
  if (hi < total - 1) pages.push(null);
  pages.push(total);
  return pages;
}

export function Pagination({ page, totalPages, onPageChange }: PaginationProps) {
  if (totalPages <= 1) return null;

  const btnBase =
    'min-w-[2rem] h-8 px-2 flex items-center justify-center rounded-md text-sm font-medium transition-colors disabled:opacity-40 disabled:cursor-not-allowed';
  const activeBtn = `${btnBase} bg-slate-900 text-white`;
  const inactiveBtn = `${btnBase} text-slate-600 hover:bg-slate-100`;

  return (
    <nav aria-label="Pagination" className="flex items-center justify-center gap-1 pt-2">
      <button
        type="button"
        onClick={() => onPageChange(page - 1)}
        disabled={page === 1}
        aria-label="Previous page"
        className={inactiveBtn}
      >
        ‹
      </button>

      {pageRange(page, totalPages).map((p, i) =>
        p === null ? (
          <span key={`ellipsis-${i}`} className="min-w-[2rem] h-8 flex items-center justify-center text-slate-400 text-sm">
            …
          </span>
        ) : (
          <button
            key={p}
            type="button"
            onClick={() => onPageChange(p)}
            aria-label={`Page ${p}`}
            aria-current={p === page ? 'page' : undefined}
            className={p === page ? activeBtn : inactiveBtn}
          >
            {p}
          </button>
        ),
      )}

      <button
        type="button"
        onClick={() => onPageChange(page + 1)}
        disabled={page === totalPages}
        aria-label="Next page"
        className={inactiveBtn}
      >
        ›
      </button>
    </nav>
  );
}
