import { writable } from 'svelte/store';
import { api } from '$lib/api/client';

export const isAuthenticated = writable(api.isAuthenticated);

export function login(email: string, password: string) {
  return api.login(email, password).then((result) => {
    isAuthenticated.set(true);
    return result;
  });
}

export function logout() {
  api.clearToken();
  isAuthenticated.set(false);
}

export function getMe() {
  return api.getMe();
}
