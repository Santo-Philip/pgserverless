import { writable } from 'svelte/store';

export type Theme = 'dark' | 'light';

function getInitialTheme(): Theme {
  if (typeof localStorage !== 'undefined') {
    return (localStorage.getItem('theme') as Theme) || 'dark';
  }
  return 'dark';
}

export const theme = writable<Theme>(getInitialTheme());

theme.subscribe((value) => {
  if (typeof document !== 'undefined') {
    document.documentElement.setAttribute('data-theme', value);
    const meta = document.querySelector('meta[name="theme-color"]');
    if (meta) {
      meta.setAttribute('content', value === 'dark' ? '#0f1117' : '#f8f9fa');
    }
    localStorage.setItem('theme', value);
  }
});

export function toggleTheme() {
  theme.update((t) => (t === 'dark' ? 'light' : 'dark'));
}
