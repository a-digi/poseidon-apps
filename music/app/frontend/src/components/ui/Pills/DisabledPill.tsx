import React from 'react';
import { BUTTON_BASE_CLASSES } from '../../ui';

export interface DisabledPillProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  label?: string;
}

const DisabledPill: React.FC<DisabledPillProps> = ({
  label = 'Disabled',
  className = '',
  children,
  ...props
}) => {
  return (
    <button
      type="button"
      disabled
      className={`inline-block rounded-full ${BUTTON_BASE_CLASSES} bg-gray-300 text-black cursor-not-allowed ${className}`.trim()}
      {...props}
    >
      {children ?? label}
    </button>
  );
};

export default DisabledPill;
