import React from 'react';
import Alert, { AlertProps } from './Alert';

const ErrorAlert: React.FC<Omit<AlertProps, 'type'>> = (props) => (
  <Alert type="error" {...props} />
);

export default ErrorAlert;
