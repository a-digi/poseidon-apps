import React from 'react';
import { BUTTON_BASE_CLASSES, BUTTON_DELETE_GRADIENT } from '../../ui';

export interface DeleteButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  label?: string;
}

const DeleteButton: React.FC<DeleteButtonProps> = ({ label = 'Delete', ...props }) => {
  return (
    <button
      type="button"
      className={`${BUTTON_BASE_CLASSES} ${BUTTON_DELETE_GRADIENT}`}
      {...props}
    >
      {label}
    </button>
  );
};

export default DeleteButton;
