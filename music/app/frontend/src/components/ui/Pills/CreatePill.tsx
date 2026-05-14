import React from 'react';
import { BUTTON_BASE_CLASSES, BUTTON_CREATE_GRADIENT } from '../../ui';

export interface CreatePillProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  label?: string;
}

const CreatePill: React.FC<CreatePillProps> = ({
  label = 'Create',
  className = '',
  children,
  ...props
}) => {
  return (
    <button
      type="button"
      className={`inline-block rounded-full ${BUTTON_BASE_CLASSES} ${BUTTON_CREATE_GRADIENT} ${className}`.trim()}
      {...props}
    >
      {children ?? label}
    </button>
  );
};

export default CreatePill;
