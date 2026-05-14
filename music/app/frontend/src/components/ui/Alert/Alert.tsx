import React from 'react';

export type AlertType = 'info' | 'success' | 'warning' | 'error' | 'neutral';

export interface AlertProps extends React.HTMLAttributes<HTMLDivElement> {
  type?: AlertType;
  title?: string;
  message: string;
  icon?: React.ReactNode;
}

const ALERT_STYLES: Record<AlertType, string> = {
  info: 'bg-blue-50 border-blue-300 text-blue-800',
  success: 'bg-green-50 border-green-400 text-green-800',
  warning: 'bg-yellow-50 border-yellow-400 text-yellow-900',
  error: 'bg-red-50 border-red-400 text-red-800',
  neutral: 'bg-gray-50 border-gray-300 text-gray-800',
};

const ICONS: Record<AlertType, React.ReactNode> = {
  info: (
    <svg className="w-5 h-5 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2" fill="none" />
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 16v-4m0-4h.01" />
    </svg>
  ),
  success: (
    <svg className="w-5 h-5 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2" fill="none" />
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4" />
    </svg>
  ),
  warning: (
    <svg className="w-5 h-5 text-yellow-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2" fill="none" />
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4m0 4h.01" />
    </svg>
  ),
  error: (
    <svg className="w-5 h-5 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2" fill="none" />
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4m0 4h.01" />
    </svg>
  ),
  neutral: (
    <svg className="w-5 h-5 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2" fill="none" />
    </svg>
  ),
};

const Alert: React.FC<AlertProps> = ({ type = 'info', title, message, icon, className = '', ...props }) => {
  return (
    <div
      className={`flex items-start gap-3 border-l-4 rounded p-4 shadow-sm ${ALERT_STYLES[type]} ${className}`.trim()}
      role="alert"
      {...props}
    >
      <div className="pt-0.5">{icon ?? ICONS[type]}</div>
      <div>
        {title && <div className="font-semibold mb-1">{title}</div>}
        <div className="text-sm leading-relaxed">{message}</div>
      </div>
    </div>
  );
};

export default Alert;
