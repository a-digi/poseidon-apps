interface InfoAlertProps {
  message: string;
}

const InfoIcon = () => (
  <svg
    className="w-5 h-5 text-blue-500 shrink-0 mt-0.5"
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
    aria-hidden="true"
  >
    <circle cx="12" cy="12" r="10" strokeWidth="2" />
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 16v-4m0-4h.01" />
  </svg>
);

export function InfoAlert({ message }: InfoAlertProps) {
  return (
    <div
      className="flex items-start gap-3 bg-blue-50 border-l-4 border-blue-300 text-blue-800 rounded p-3"
      role="alert"
    >
      <InfoIcon />
      <p className="text-sm leading-relaxed">{message}</p>
    </div>
  );
}
