import { useEffect, useRef, useState } from 'react';
import { Check, ChevronDown, Search } from 'lucide-react';

export interface DropdownOption {
  value: string;
  label: string;
}

interface BaseProps {
  options: DropdownOption[];
  placeholder?: string;
  disabled?: boolean;
}

interface SingleProps extends BaseProps {
  mode: 'single';
  value: string;
  onChange: (value: string) => void;
}

interface MultiProps extends BaseProps {
  mode: 'multi';
  value: string[];
  onChange: (value: string[]) => void;
}

export type DropdownProps = SingleProps | MultiProps;

export function Dropdown(props: DropdownProps) {
  const { options, placeholder = 'Select...', disabled = false } = props;
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState('');
  const containerRef = useRef<HTMLDivElement>(null);
  const searchRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (!open) return;
    const onMouseDown = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', onMouseDown);
    return () => document.removeEventListener('mousedown', onMouseDown);
  }, [open]);

  useEffect(() => {
    if (open) {
      setSearch('');
      searchRef.current?.focus();
    }
  }, [open]);

  const filtered = options.filter(o =>
    o.label.toLowerCase().includes(search.toLowerCase()),
  );

  const isSelected = (v: string) =>
    props.mode === 'multi' ? props.value.includes(v) : props.value === v;

  const handleSelect = (v: string) => {
    if (props.mode === 'multi') {
      props.onChange(
        props.value.includes(v) ? props.value.filter(x => x !== v) : [...props.value, v],
      );
    } else {
      props.onChange(v);
      setOpen(false);
    }
  };

  let triggerLabel: string;
  if (props.mode === 'multi') {
    const n = props.value.length;
    triggerLabel =
      n === 0
        ? placeholder
        : n === 1
          ? (options.find(o => o.value === props.value[0])?.label ?? placeholder)
          : `${n} selected`;
  } else {
    triggerLabel = options.find(o => o.value === props.value)?.label ?? placeholder;
  }

  const hasValue = props.mode === 'multi' ? props.value.length > 0 : props.value !== '';

  return (
    <div ref={containerRef} className="relative">
      <button
        type="button"
        disabled={disabled}
        onClick={() => setOpen(prev => !prev)}
        aria-haspopup="listbox"
        aria-expanded={open}
        className={`w-full flex items-center justify-between gap-2 px-3 py-2 rounded-md border bg-white text-sm text-left transition-all duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-900 focus-visible:ring-offset-1 disabled:opacity-40 disabled:cursor-not-allowed ${
          open ? 'border-slate-400 shadow-sm' : 'border-slate-200 hover:border-slate-300'
        }`}
      >
        <span className={`truncate ${hasValue ? 'text-slate-900' : 'text-slate-400'}`}>
          {triggerLabel}
        </span>
        <ChevronDown
          className={`w-4 h-4 shrink-0 text-slate-400 transition-transform duration-200 ${open ? 'rotate-180' : ''}`}
        />
      </button>

      {open && (
        <div className="absolute z-50 top-full left-0 right-0 mt-1 bg-white rounded-xl shadow-xl border border-slate-100 overflow-hidden">
          <div className="p-2 border-b border-slate-100">
            <div className="relative">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-400 pointer-events-none" />
              <input
                ref={searchRef}
                type="text"
                value={search}
                onChange={e => setSearch(e.target.value)}
                placeholder="Search…"
                aria-label="Search options"
                onKeyDown={e => e.key === 'Escape' && setOpen(false)}
                className="w-full pl-7 pr-3 py-1.5 text-sm bg-slate-50 border border-slate-200 rounded-md focus:outline-none focus:ring-2 focus:ring-slate-900 focus:ring-offset-0"
              />
            </div>
          </div>

          <ul role="listbox" aria-multiselectable={props.mode === 'multi'} className="max-h-52 overflow-y-auto py-1">
            {filtered.length === 0 ? (
              <li className="px-3 py-2 text-sm text-slate-400 text-center">No results</li>
            ) : (
              filtered.map(option => {
                const selected = isSelected(option.value);
                return (
                  <li
                    key={option.value}
                    role="option"
                    aria-selected={selected}
                    tabIndex={0}
                    onClick={() => handleSelect(option.value)}
                    onKeyDown={e => (e.key === 'Enter' || e.key === ' ') && handleSelect(option.value)}
                    className={`flex items-center gap-2.5 px-3 py-2 text-sm cursor-pointer transition-colors focus:outline-none focus:bg-slate-100 ${
                      selected ? 'bg-slate-50 text-slate-900' : 'text-slate-700 hover:bg-slate-50'
                    }`}
                  >
                    {props.mode === 'multi' && (
                      <span
                        className={`shrink-0 w-4 h-4 rounded border flex items-center justify-center transition-colors ${
                          selected ? 'bg-slate-900 border-slate-900' : 'border-slate-300 bg-white'
                        }`}
                      >
                        {selected && <Check className="w-2.5 h-2.5 text-white" />}
                      </span>
                    )}
                    <span className="flex-1 truncate">{option.label}</span>
                    {props.mode === 'single' && selected && (
                      <Check className="w-4 h-4 shrink-0 text-slate-900" />
                    )}
                  </li>
                );
              })
            )}
          </ul>

          {props.mode === 'multi' && props.value.length > 0 && (
            <div className="border-t border-slate-100 px-3 py-1.5 flex justify-end">
              <button
                type="button"
                onClick={() => props.onChange([])}
                className="text-xs text-slate-400 hover:text-slate-700 transition-colors"
              >
                Clear all
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
