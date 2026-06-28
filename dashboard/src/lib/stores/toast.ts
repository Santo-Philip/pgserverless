import { writable } from 'svelte/store';

export type ToastType = 'success' | 'error' | 'warning' | 'info';

export interface Toast {
  id: string;
  message: string;
  type: ToastType;
  duration?: number;
}

function createToastStore() {
  const { subscribe, update } = writable<Toast[]>([]);

  let counter = 0;

  function add(message: string, type: ToastType = 'info', duration = 4000) {
    const id = `toast-${++counter}`;
    update((t) => [...t, { id, message, type, duration }]);
    if (duration > 0) {
      setTimeout(() => remove(id), duration);
    }
    return id;
  }

  function remove(id: string) {
    update((t) => t.filter((toast) => toast.id !== id));
  }

  function success(message: string) { return add(message, 'success'); }
  function error(message: string) { return add(message, 'error', 6000); }
  function warning(message: string) { return add(message, 'warning'); }
  function info(message: string) { return add(message, 'info'); }

  return {
    subscribe,
    add,
    remove,
    success,
    error,
    warning,
    info,
  };
}

export const toast = createToastStore();
