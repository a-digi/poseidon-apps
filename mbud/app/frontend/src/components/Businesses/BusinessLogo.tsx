interface BusinessLogoProps {
  name: string;
  logo?: string;
  size?: number;
}

function initials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return '?';
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

export function BusinessLogo({ name, logo, size = 48 }: BusinessLogoProps) {
  const style = { width: `${size}px`, height: `${size}px` };
  if (logo) {
    return (
      <img
        src={logo}
        alt={`${name} logo`}
        style={style}
        className="rounded-lg object-cover shrink-0 border border-slate-200"
      />
    );
  }
  return (
    <div
      style={style}
      aria-label={`${name} initials`}
      className="rounded-lg shrink-0 flex items-center justify-center bg-gradient-to-br from-slate-700 to-slate-900 text-white font-semibold select-none"
    >
      <span className={size >= 56 ? 'text-base' : 'text-sm'}>{initials(name)}</span>
    </div>
  );
}
