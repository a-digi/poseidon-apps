import React, { useEffect, useRef, useState } from 'react';
import ModalDialog from './ModalDialog';
import CreateButton from '../ui/Button/CreateButton';
import CancelButton from '../ui/Button/CancelButton';

export interface PromptDialogProps {
  open: boolean;
  title: string;
  placeholder?: string;
  initialValue?: string;
  errorMessage?: string;
  onCancel: () => void;
  onSubmit: (value: string) => void;
}

const PromptDialog: React.FC<PromptDialogProps> = ({
  open,
  title,
  placeholder,
  initialValue = '',
  errorMessage,
  onCancel,
  onSubmit,
}) => {
  const [value, setValue] = useState(initialValue);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (open) {
      setValue(initialValue);
      setTimeout(() => inputRef.current?.focus(), 0);
    }
  }, [open, initialValue]);

  const trySubmit = () => {
    const trimmed = value.trim();
    if (!trimmed) return;
    onSubmit(trimmed);
  };

  return (
    <ModalDialog
      open={open}
      onClose={onCancel}
      title={title}
      errorMessage={errorMessage}
      actions={
        <>
          <CancelButton className="p-2" onClick={onCancel} label="Cancel" />
          <CreateButton onClick={trySubmit} label="OK" />
        </>
      }
    >
      <input
        ref={inputRef}
        type="text"
        value={value}
        onChange={(e) => setValue(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter') {
            e.preventDefault();
            trySubmit();
          }
        }}
        placeholder={placeholder}
        className="w-full border border-blue-300 px-3 py-2 rounded focus:outline-none focus:ring-2 focus:ring-blue-400 bg-white"
      />
    </ModalDialog>
  );
};

export default PromptDialog;
