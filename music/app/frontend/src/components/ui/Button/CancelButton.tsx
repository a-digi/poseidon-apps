import React from 'react';
import {BUTTON_BASE_CLASSES, BUTTON_CANCEL_GRADIENT} from '@/components/ui';

interface CancelButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  label: string;
}

const CancelButton: React.FC<CancelButtonProps> = ({ label, className = '', ...props }) => {
  return (
    <button
      type="button"
      className={`${BUTTON_BASE_CLASSES} ${BUTTON_CANCEL_GRADIENT} ${className}`.trim()}
      {...props}
    >
      {label}
    </button>
  );
};

export default CancelButton;
