import React from 'react';
import Alert, { AlertProps } from './Alert';

const WarningAlert: React.FC<Omit<AlertProps, 'type'>> = (props) => (
  <Alert type="warning" {...props} />
);

export default WarningAlert;
