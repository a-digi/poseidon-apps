import React, { ReactNode } from 'react';

export interface CardsContainerProps {
  children: ReactNode;
  columns?: string; // z.B. '1fr 1fr 1fr' oder '300px 1fr 2fr'
  gap?: string; // z.B. '1rem', '2rem', '16px'
  className?: string;
}

const CardsContainer: React.FC<CardsContainerProps> = ({
  children,
  columns = '1fr 1fr 1fr',
  gap = '1.5rem',
  className = '',
}) => {
  return (
    <div
      className={`w-full ${className}`}
      style={{
        display: 'grid',
        gridTemplateColumns: columns,
        gap,
        alignItems: 'stretch',
      }}
    >
      {children}
    </div>
  );
};

export default CardsContainer;
