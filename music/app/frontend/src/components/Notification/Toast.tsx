import React, { useEffect } from 'react';

interface ToastProps {
  id?: string;
  message: string;
  type?: 'success' | 'error' | 'info' | 'warning';
  duration?: number; // in ms
  onClose?: () => void;
  position?:
    | 'top-right'
    | 'top-left'
    | 'bottom-right'
    | 'bottom-left'
    | 'top-center'
    | 'bottom-center';
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
}

const typeStyles: Record<string, string> = {
  success: 'bg-green-500 text-white',
  error: 'bg-red-500 text-white',
  info: 'bg-blue-500 text-white',
  warning: 'bg-yellow-400 text-black',
};

const positionStyles: Record<string, string> = {
  'top-right': 'top-6 right-6',
  'top-left': 'top-6 left-6',
  'bottom-right': 'bottom-6 right-6',
  'bottom-left': 'bottom-6 left-6',
  'top-center': 'top-6 left-1/2 -translate-x-1/2',
  'bottom-center': 'bottom-6 left-1/2 -translate-x-1/2',
};

const sizeStyles: Record<string, string> = {
  sm: 'max-w-xs w-64',
  md: 'max-w-sm w-80',
  lg: 'max-w-md w-96',
  xl: 'max-w-lg w-[32rem]',
  full: 'w-full max-w-full',
};

const Toast: React.FC<ToastProps> = ({
  message,
  type = 'info',
  duration = 3000,
  onClose,
  position = 'top-center',
  size = 'xl',
}) => {
  const [progress, setProgress] = React.useState(100);

  useEffect(() => {
    if (!onClose) return;
    setProgress(100);
    const start = Date.now();
    const timer = setInterval(() => {
      const elapsed = Date.now() - start;
      const percent = Math.max(0, 100 - (elapsed / duration) * 100);
      setProgress(percent);
      if (percent <= 0) {
        clearInterval(timer);
      }
    }, 30);
    const timeout = setTimeout(() => {
      onClose();
    }, duration);
    return () => {
      clearTimeout(timeout);
      clearInterval(timer);
    };
  }, [onClose, duration]);

  return (
    <div
      className={`fixed z-50 px-4 py-3 rounded shadow-lg flex items-center gap-2 transition-all animate-fade-in-up ${typeStyles[type]} ${positionStyles[position]} ${sizeStyles[size]}`}
      role="alert"
    >
      <span>{message}</span>
      {onClose && (
        <button
          className="ml-2 text-lg font-bold focus:outline-none"
          onClick={onClose}
          aria-label="Schließen"
        >
          ×
        </button>
      )}
      <div className="absolute left-0 bottom-0 h-1 w-full">
        <div
          className="h-full bg-white/60 transition-all duration-75"
          style={{ width: `${progress}%` }}
        />
      </div>
    </div>
  );
};

export default Toast;
