import React, { createContext, useContext, useState, ReactNode } from 'react';

export interface TitleObject {
  text: string;
  icon?: React.ReactNode;
}

type TitleType = string | TitleObject | null;

interface TopBarContextType {
  components: ReactNode[];
  addComponent: (component: ReactNode) => void;
  removeComponent: (component: ReactNode) => void;
  title: TitleType;
  setTitle: (title: TitleType) => void;
}

const TopBarContext = createContext<TopBarContextType | undefined>(undefined);

export const TopBarProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [components, setComponents] = useState<ReactNode[]>([]);
  const [title, setTitle] = useState<TitleType>(null);

  const addComponent = (component: ReactNode) => {
    setComponents((prev) => [...prev, component]);
  };

  const removeComponent = (component: ReactNode) => {
    setComponents((prev) => prev.filter((c) => c !== component));
  };

  return (
    <TopBarContext.Provider value={{ components, addComponent, removeComponent, title, setTitle }}>
      {children}
    </TopBarContext.Provider>
  );
};

export const useTopBar = () => {
  const context = useContext(TopBarContext);
  if (!context) throw new Error('useTopBar must be used within a TopBarProvider');
  return context;
};
