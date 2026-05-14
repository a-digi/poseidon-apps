import React from 'react';

export interface CardProps {
  icon?: React.ReactNode;
  iconPosition?: 'left' | 'right';
  title?: React.ReactNode;
  children?: React.ReactNode;
  className?: string;
  color?:
    | 'white'
    | 'gray'
    | 'blue'
    | 'green'
    | 'red'
    | 'dark'
    | 'blueGradient'
    | 'greenGradient'
    | 'redGradient'
    | 'darkGradient';
  shadow?: 'none' | 'sm' | 'md' | 'lg' | 'xl' | '2xl'; // NEU: Schattenoption
  centered?: boolean; // NEU: Inhalt zentrieren (horizontal & vertikal)
}

const colorMap: Record<string, string> = {
  white: 'bg-white text-gray-900 border-gray-200',
  gray: 'bg-gray-100 text-gray-900 border-gray-200',
  blue: 'bg-blue-600 text-white border-blue-700',
  green: 'bg-green-600 text-white border-green-700',
  red: 'bg-red-600 text-white border-red-700',
  dark: 'bg-gray-800 text-white border-gray-900',
  // Gradient-Varianten
  blueGradient: 'bg-gradient-to-r from-blue-500 to-blue-700 text-white border-blue-700',
  greenGradient: 'bg-gradient-to-r from-green-500 to-green-700 text-white border-green-700',
  redGradient: 'bg-gradient-to-r from-red-500 to-red-700 text-white border-red-700',
  darkGradient: 'bg-gradient-to-r from-gray-800 to-gray-900 text-white border-gray-900',
};

const shadowMap: Record<string, string> = {
  none: 'shadow-none',
  sm: 'shadow-sm',
  md: 'shadow-md',
  lg: 'shadow-lg',
  xl: 'shadow-xl',
  '2xl': 'shadow-2xl',
};

const Card: React.FC<CardProps> = ({
  icon,
  iconPosition = 'left',
  title,
  children,
  className = '',
  color = 'white',
  shadow = 'md',
  centered = false, // NEU: Standard nicht zentriert
}) => {
  return (
    <div
      className={`rounded-lg ${shadowMap[shadow] || shadowMap.md} p-4 flex items-center justify-center border relative overflow-visible min-h-[120px] ${colorMap[color] || colorMap.white} ${className}`}
    >
      <div
        className={`flex flex-col w-full${centered ? ' items-center justify-center h-full flex-1' : ''}`}
        style={centered ? { minHeight: '120px' } : {}}
      >
        {title && (
          <div
            className={`flex items-center gap-2 justify-center mb-1 ${iconPosition === 'left' ? 'flex-row' : 'flex-row-reverse'}`}
          >
            {icon && <span className="text-xl">{icon}</span>}
            <span className="font-semibold text-base">{title}</span>
          </div>
        )}
        <div
          className={`w-full${centered ? ' flex flex-col items-center justify-center flex-1' : ''}`}
          style={centered ? { flex: 1, height: '100%' } : {}}
        >
          {children}
        </div>
      </div>
    </div>
  );
};

export default Card;
