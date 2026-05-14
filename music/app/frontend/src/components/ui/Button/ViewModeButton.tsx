import React from 'react';
import { BUTTON_BASE_CLASSES, BUTTON_CREATE_GRADIENT } from '../../ui';

export interface ViewModeButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  active?: boolean;
  rounded?: 'left' | 'right' | 'full' | 'none';
}

const getRoundedClass = (rounded: 'left' | 'right' | 'full' | 'none' = 'left') => {
  switch (rounded) {
    case 'right':
      return 'rounded-r';
    case 'full':
      return 'rounded-full';
    case 'none':
      return '';
    case 'left':
    default:
      return 'rounded-l';
  }
};

const ViewModeButton: React.FC<ViewModeButtonProps> = ({
  active = false,
  rounded = 'left',
  className = '',
  children,
  ...props
}) => {
  const base = `${BUTTON_BASE_CLASSES}`;
  const gradientClass = active ? BUTTON_CREATE_GRADIENT : '';
  const inactiveClass = !active ? 'bg-gray-200 text-gray-800' : '';
  const roundedClass = getRoundedClass(rounded);
  return (
    <button
      className={`${base} ${roundedClass} ${gradientClass} ${inactiveClass} ${className}`.trim()}
      aria-pressed={active}
      {...props}
    >
      {children}
    </button>
  );
};

export default ViewModeButton;
