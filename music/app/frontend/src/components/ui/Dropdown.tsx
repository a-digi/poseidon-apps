import React, { useState } from 'react';

export interface DropdownItem {
  label: string;
  value: string;
}

interface DropdownProps {
  label: string;
  items: DropdownItem[];
  selectedValue: string;
  onSelect: (value: string) => void;
  disabled?: boolean;
}

const Dropdown: React.FC<DropdownProps> = ({ label, items, selectedValue, onSelect, disabled }) => {
  const [open, setOpen] = useState(false);
  const selected = items.find((i) => i.value === selectedValue);
  return (
    <div
      className={`relative w-64 my-2 ${disabled ? 'opacity-60 cursor-not-allowed' : ''}`}
      tabIndex={-1}
      onClick={(e) => e.stopPropagation()}
    >
      <button
        className={`w-full bg-white text-gray-900 border border-gray-300 rounded px-4 py-2 text-left shadow flex justify-between items-center hover:border-blue-400 transition
          ${disabled ? 'bg-gray-100 text-gray-400 cursor-not-allowed' : ''}`}
        onClick={() => !disabled && setOpen((v) => !v)}
        type="button"
        disabled={disabled}
        style={disabled ? { pointerEvents: 'auto' } : {}}
      >
        <span>{selected ? selected.label : label}</span>
        <svg
          className={`w-4 h-4 ml-2 transition-transform ${open ? 'rotate-180' : ''}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path d="M19 9l-7 7-7-7" />
        </svg>
      </button>
      {open && !disabled && (
        <ul className="absolute text-gray-900 z-10 w-full bg-white border border-gray-200 rounded shadow mt-1 max-h-60 overflow-auto">
          {items.map((item) => (
            <li
              key={item.value}
              className={`px-4 py-2 cursor-pointer hover:bg-blue-100 ${item.value === selectedValue ? 'bg-blue-50 font-bold' : ''}`}
              onClick={() => {
                onSelect(item.value);
                setOpen(false);
              }}
            >
              {item.label}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

export default Dropdown;
