const DEFAULT_OPTIONS = [12, 24, 48] as const;

interface LimitProps {
  value: number;
  onChange: (limit: number) => void;
  options?: readonly number[];
}

export function Limit({ value, onChange, options = DEFAULT_OPTIONS }: LimitProps) {
  return (
    <div className="flex items-center gap-2 text-sm text-slate-500">
      <span>Show</span>
      <select
        value={value}
        onChange={e => onChange(Number(e.target.value))}
        aria-label="Items per page"
        className="px-2 py-1 rounded-md border border-slate-200 bg-white text-slate-700 text-sm focus:outline-none focus:ring-2 focus:ring-slate-900 focus:ring-offset-1 cursor-pointer"
      >
        {options.map(n => (
          <option key={n} value={n}>{n}</option>
        ))}
      </select>
      <span>per page</span>
    </div>
  );
}
