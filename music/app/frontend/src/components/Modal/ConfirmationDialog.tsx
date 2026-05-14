import React from 'react';
import ModalDialog from './ModalDialog';
import CreateButton from '../ui/Button/CreateButton';
import CancelButton from '../ui/Button/CancelButton';

export interface ConfirmationDialogProps {
  open: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title?: string;
  message?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  errorMessage?: string;
}

const ConfirmationDialog: React.FC<ConfirmationDialogProps> = ({
  open,
  onClose,
  onConfirm,
  title,
  message,
  confirmLabel = 'OK',
  cancelLabel = 'Cancel',
  errorMessage,
}) => {
  return (
    <ModalDialog
      open={open}
      onClose={onClose}
      title={title}
      errorMessage={errorMessage}
      actions={
        <>
          <CancelButton className={'p-2'} onClick={onClose} label={cancelLabel} />
          <CreateButton onClick={onConfirm} label={confirmLabel} />
        </>
      }
    >
      {message && <div className="mb-2 text-gray-800">{message}</div>}
    </ModalDialog>
  );
};

export default ConfirmationDialog;
