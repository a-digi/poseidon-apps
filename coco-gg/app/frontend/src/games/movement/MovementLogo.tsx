interface MovementLogoProps {
  className?: string;
}

export function MovementLogo({ className }: MovementLogoProps) {
  return (
    <svg
      viewBox="0 0 40 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      aria-hidden="true"
    >
      <rect x="2" y="2" width="36" height="36" rx="7" stroke="currentColor" strokeWidth="2" strokeOpacity="0.25" />
      <circle cx="20" cy="20" r="7" fill="currentColor" />
      <path d="M20 11 L20 6 M17.5 8.5 L20 6 L22.5 8.5" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M20 29 L20 34 M17.5 31.5 L20 34 L22.5 31.5" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M11 20 L6 20 M8.5 17.5 L6 20 L8.5 22.5" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M29 20 L34 20 M31.5 17.5 L34 20 L31.5 22.5" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}
