import React from 'react';
import Alert, { AlertProps } from './Alert';

const InfoAlert: React.FC<Omit<AlertProps, 'type'>> = (props) => (
  <Alert type="info" {...props} />
);

export default InfoAlert;
