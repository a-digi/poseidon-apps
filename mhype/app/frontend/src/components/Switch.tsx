interface SwitchProps {
  checked: boolean;
  onChange: (next: boolean) => void;
  ariaLabel: string;
  disabled?: boolean;
}

export function Switch({ checked, onChange, ariaLabel, disabled }: SwitchProps) {
  const pillColor = checked ? 'bg-blue-600' : 'bg-slate-200';
  const disabledClass = disabled ? 'opacity-50 cursor-not-allowed' : '';
  const knobTranslate = checked ? 'translate-x-4' : 'translate-x-0.5';

  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      aria-label={ariaLabel}
      disabled={disabled}
      onClick={() => onChange(!checked)}
      className={`inline-flex items-center w-9 h-5 rounded-full transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 ${pillColor} ${disabledClass}`}
    >
      <span className={`block w-4 h-4 bg-white rounded-full shadow transition-transform ${knobTranslate}`} />
    </button>
  );
}
