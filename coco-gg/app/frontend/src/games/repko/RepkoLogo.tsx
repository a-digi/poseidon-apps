interface RepkoLogoProps {
  className?: string;
}

export function RepkoLogo({ className }: RepkoLogoProps) {
  return (
    <svg
      viewBox="0 0 40 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      aria-hidden="true"
    >
      <polygon
        points="20,3 35,12 35,28 20,37 5,28 5,12"
        stroke="currentColor"
        strokeWidth="2"
        strokeOpacity="0.35"
        fill="none"
      />
      <polygon points="14,12 20,9 26,12 26,18 20,21 14,18" fill="#92400e" />
      <polygon points="9,21 15,18 21,21 21,27 15,30 9,27" fill="#15803d" />
      <polygon points="19,21 25,18 31,21 31,27 25,30 19,27" fill="#ca8a04" />
    </svg>
  );
}
