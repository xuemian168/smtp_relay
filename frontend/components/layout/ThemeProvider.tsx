"use client";
import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';

// 主题类型
type Theme = 'light' | 'dark' | 'system';

interface ThemeContextProps {
  theme: Theme;
  setTheme: (theme: Theme) => void;
}

const ThemeContext = createContext<ThemeContextProps | undefined>(undefined);

export const useTheme = () => {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
};

interface Props {
  children: ReactNode;
}

function getInitialTheme(): Theme {
  if (typeof window !== 'undefined') {
    const saved = localStorage.getItem('theme');
    if (saved === 'light' || saved === 'dark' || saved === 'system') {
      return saved;
    }
  }
  return 'system';
}

export const ThemeProvider = ({ children }: Props) => {
  const [theme, setThemeState] = useState<Theme>(getInitialTheme);
  const [mounted, setMounted] = useState(false);

  // 挂载后标记mounted，防止闪色
  useEffect(() => {
    setMounted(true);
  }, []);

  // 应用主题到html标签
  useEffect(() => {
    const root = window.document.documentElement;
    let appliedTheme = theme;
    if (theme === 'system') {
      const mql = window.matchMedia('(prefers-color-scheme: dark)');
      appliedTheme = mql.matches ? 'dark' : 'light';
    }
    root.classList.remove('light', 'dark');
    root.classList.add(appliedTheme);
    localStorage.setItem('theme', theme);
  }, [theme]);

  const setTheme = (t: Theme) => {
    setThemeState(t);
  };

  if (!mounted) return null;

  return (
    <ThemeContext.Provider value={{ theme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}; 