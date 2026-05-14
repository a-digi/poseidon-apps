import React from 'react';
import Alert, { AlertProps } from './Alert';

const SuccessAlert: React.FC<Omit<AlertProps, 'type'>> = (props) => (
  <Alert type="success" {...props} />
);

export default SuccessAlert;
