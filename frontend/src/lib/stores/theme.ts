import { writable } from 'svelte/store';

export type Theme = 'dark' | 'light' | 'system';

function createThemeStore() {
  // Load initial theme from localStorage or default to system
  const savedTheme = (typeof localStorage !== 'undefined' ? localStorage.getItem('theme') : 'system') as Theme;
  const { subscribe, set } = writable<Theme>(savedTheme || 'system');

  return {
    subscribe,
    set: (value: Theme) => {
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem('theme', value);
      }
      set(value);
      applyTheme(value);
    },
    init: () => {
      applyTheme(savedTheme);
      
      // Watch for system theme changes
      if (typeof window !== 'undefined') {
        window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
          const current = localStorage.getItem('theme') as Theme;
          if (current === 'system') {
            applyTheme('system');
          }
        });
      }
    }
  };
}

function applyTheme(theme: Theme) {
  if (typeof document === 'undefined') return;
  
  const root = document.documentElement;
  let effectiveTheme = theme;

  if (theme === 'system') {
    effectiveTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }

  if (effectiveTheme === 'light') {
    root.classList.add('light');
    root.classList.remove('dark');
    root.style.colorScheme = 'light';
  } else {
    root.classList.add('dark');
    root.classList.remove('light');
    root.style.colorScheme = 'dark';
  }
}

export const theme = createThemeStore();
