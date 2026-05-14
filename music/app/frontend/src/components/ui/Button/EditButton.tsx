import React from 'react';
import { BUTTON_BASE_CLASSES, BUTTON_EDIT_GRADIENT } from '../../ui';

export interface EditButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  label?: string;
}

const EditButton: React.FC<EditButtonProps> = ({ label = 'Edit', className = '', ...props }) => {
  return (
    <button
      type="button"
      className={`${BUTTON_BASE_CLASSES} ${BUTTON_EDIT_GRADIENT} ${className}`.trim()}
      {...props}
    >
      {label}
    </button>
  );
};

export default EditButton;
