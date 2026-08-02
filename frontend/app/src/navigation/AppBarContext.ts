import * as React from 'react';

export const AppBarRightContext = React.createContext<{
  right: React.ReactNode;
  setRight: (node: React.ReactNode) => void;
}>({ right: null, setRight: () => {} });
