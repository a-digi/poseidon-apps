import React from 'react';
import { BUTTON_DELETE_GRADIENT, BUTTON_BASE_CLASSES } from '@/components/ui';

interface DeletePillProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  children: React.ReactNode;
}

const DeletePill: React.FC<DeletePillProps> = ({ children, className = '', ...props }) => {
  return (
    <button
      type="button"
      className={`inline-block rounded-full ${BUTTON_BASE_CLASSES} ${BUTTON_DELETE_GRADIENT} ${className}`.trim()}
      {...props}
    >
      {children}
    </button>
  );
};

export default DeletePill;
