import React from 'react';
import ViewModeButton, { ViewModeButtonProps } from './ViewModeButton';

const GridIcon = () => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    className="w-5 h-5"
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
  >
    <rect x="4" y="4" width="7" height="7" rx="2" />
    <rect x="13" y="4" width="7" height="7" rx="2" />
    <rect x="4" y="13" width="7" height="7" rx="2" />
    <rect x="13" y="13" width="7" height="7" rx="2" />
  </svg>
);

const GridViewModeButton: React.FC<ViewModeButtonProps> = (props) => {
  return (
    <ViewModeButton {...props} rounded="right">
      <GridIcon />
    </ViewModeButton>
  );
};

export default GridViewModeButton;

