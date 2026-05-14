import React, { createContext, useContext, useState, ReactNode } from 'react';

interface SidebarContextType {
  wide: boolean;
  toggleWide: () => void;
  setWide: (wide: boolean) => void;
}

const SidebarContext = createContext<SidebarContextType | undefined>(undefined);

export const SidebarProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [wide, setWide] = useState(true);
  const toggleWide = () => setWide((w) => !w);

  return (
    <SidebarContext.Provider value={{ wide, toggleWide, setWide }}>
      {children}
    </SidebarContext.Provider>
  );
};

export function useSidebar() {
  const ctx = useContext(SidebarContext);
  if (!ctx) throw new Error('useSidebar must be used within a SidebarProvider');
  return ctx;
}

