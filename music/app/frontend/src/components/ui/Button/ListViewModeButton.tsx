import React from 'react';
import ViewModeButton, { ViewModeButtonProps } from './ViewModeButton';

const ListIcon = () => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    className="w-5 h-5"
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
  >
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={2}
      d="M4 6h16M4 12h16M4 18h16"
    />
  </svg>
);

const ListViewModeButton: React.FC<ViewModeButtonProps> = (props) => {
  return (
    <ViewModeButton {...props}>
      <ListIcon />
    </ViewModeButton>
  );
};

export default ListViewModeButton;

