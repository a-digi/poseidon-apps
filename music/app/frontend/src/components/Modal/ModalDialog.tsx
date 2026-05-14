import React from 'react';

export interface ModalDialogProps {
  open: boolean;
  onClose: () => void;
  title?: string;
  children?: React.ReactNode;
  actions?: React.ReactNode;
  errorMessage?: string;
}

const ModalDialog: React.FC<ModalDialogProps> = ({
  open,
  onClose,
  title,
  children,
  actions,
  errorMessage,
}) => {
  if (!open) return null;
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-40">
      <div className="bg-white rounded-lg shadow-lg w-full max-w-md p-6 relative">
        {title && <h2 className="text-xl font-bold mb-4">{title}</h2>}
        <button
          className="absolute top-4 right-4 text-gray-500 hover:text-gray-700 text-xl"
          onClick={onClose}
          aria-label="Schließen"
        >
          &times;
        </button>
        <div className="mb-4">{children}</div>
        {errorMessage && <div className="text-red-600 text-sm mb-2">{errorMessage}</div>}
        {actions && <div className="flex justify-end gap-2">{actions}</div>}
      </div>
    </div>
  );
};

export default ModalDialog;
