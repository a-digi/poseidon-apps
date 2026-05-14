import React from 'react';
import { BUTTON_BASE_CLASSES, BUTTON_CREATE_GRADIENT } from '../../ui';

export interface CreateButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  label?: string;
  leftIcon?: React.ReactNode;
}

const CreateButton: React.FC<CreateButtonProps> = ({
  label = 'Create',
  leftIcon,
  className = '',
  ...props
}) => {
  return (
    <button
      type="button"
      className={`${BUTTON_BASE_CLASSES} ${BUTTON_CREATE_GRADIENT} ${className}`.trim()}
      {...props}
    >
      {leftIcon && <span className="mr-2 flex items-center">{leftIcon}</span>}
      {label}
    </button>
  );
};

export default CreateButton;
