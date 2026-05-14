interface Props {
  data: Array<{ label: string; value: number }>;
  ariaLabel: string;
}

export function BarChart({ data, ariaLabel }: Props) {
  const hasData = data.length > 0 && data.some((d) => d.value > 0);

  if (!hasData) {
    return (
      <p className="text-sm text-slate-400 py-8 text-center">No data yet</p>
    );
  }

  const max = Math.max(...data.map((d) => d.value), 1);

  return (
    <div role="img" aria-label={ariaLabel} className="w-full">
      <div className="flex items-end gap-1 h-32">
        {data.map(({ label, value }) => (
          <div key={label} className="flex-1 flex flex-col items-center gap-1">
            <div
              style={{ height: `${Math.round((value / max) * 100)}%` }}
              className="w-full bg-blue-500 rounded-t min-h-[2px] transition-all"
            />
            <span className="text-[10px] text-slate-500 truncate w-full text-center">
              {label}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
