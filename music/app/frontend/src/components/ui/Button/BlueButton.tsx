import React from 'react';
import { BUTTON_BASE_CLASSES, BUTTON_BLUE_GRADIENT } from '../../ui';

export interface BlueButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  label?: string;
}

const BlueButton: React.FC<BlueButtonProps> = ({ label = 'Blue', ...props }) => {
  return (
    <button
      type="button"
      className={`${BUTTON_BASE_CLASSES} ${BUTTON_BLUE_GRADIENT} px-6 py-2 rounded-lg mb-6`}
      {...props}
    >
      {label}
    </button>
  );
};

export default BlueButton;
