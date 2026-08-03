import { defineStore } from 'pinia';
import { ref, computed, watch } from 'vue';
import router from '@/router';
import type { User, LoginResponse, AuthState } from '@/types';

const STORAGE_KEY = 'eregen_admin_auth_state';

export const useAuthStore = defineStore('auth', () => {
  const parseJwt = (token: string): Record<string, any> | null => {
    try {
      const base64Url = token.split('.')[1];
      if (!base64Url) return null;
      const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
      const jsonPayload = decodeURIComponent(atob(base64).split('').map(c => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)).join(''));
      return JSON.parse(jsonPayload);
    } catch { return null; }
  };

  const stored = localStorage.getItem(STORAGE_KEY);
  const initialState: AuthState = stored
    ? { token: stored.split(',')[0], user: stored.split(',')[1] === 'null' ? null : JSON.parse(stored.split(',')[1]), expiresAt: stored.split(',')[2] ? Number(stored.split(',')[2]) : null }
    : { token: null, user: null, expiresAt: null };

  const state = ref<AuthState>(initialState);

  const isExpired = computed(() => state.value.expiresAt && Date.now() >= state.value.expiresAt * 1000);
  const isLoggedIn = computed(() => !!state.value.token && !isExpired.value);
  const getUser = computed(() => state.value.user || null);
  const getToken = () => state.value.token;

  const persist = () => {
    const uStr = state.value.user ? JSON.stringify(state.value.user) : 'null';
    const eStr = state.value.expiresAt ? state.value.expiresAt.toString() : '';
    localStorage.setItem(STORAGE_KEY, `${state.value.token || ''},${uStr},${eStr}`);
  };

  watch(state.value, () => { persist(); }, { deep: true });

  const login = async (resp: LoginResponse) => {
    const jwt = parseJwt(resp.token);
    const exp = jwt?.exp || Math.floor(Date.now() / 1000) + 7200;
    state.value = { token: resp.token, user: resp.user, expiresAt: Number(exp) };
    persist();
    const rd = (router.currentRoute.value.query.redirect as string) || '/dashboard';
    router.push({ path: rd });
    return resp.user;
  };

  const logout = () => {
    state.value = { token: null, user: null, expiresAt: null };
    localStorage.removeItem(STORAGE_KEY);
    router.push({ path: '/login' });
  };

  const hasPermission = (role: string): boolean => {
    const u = getUser.value; if (!u) return false;
    if (role === 'admin') return u.role === 'admin';
    return true;
  };

  const refreshToken = async (): Promise<boolean> => {
    if (!state.value.token) return false;
    try {
      const jp = parseJwt(state.value.token);
      if (jp?.exp) { state.value.expiresAt = jp.exp + 7200; persist(); return true; }
    } catch (e) { console.warn('Token refresh failed:', e); }
    return false;
  };

  if (state.value.token) {
    const jp = parseJwt(state.value.token);
    if (jp?.exp && Date.now() >= jp.exp * 1000) {
      state.value = { token: null, user: null, expiresAt: null };
      localStorage.removeItem(STORAGE_KEY);
    } else if (jp?.sub && !state.value.user) {
      state.value.user = { id: jp.sub || '', name: jp.name || '用户', role: jp.role || 'family', phone: jp.phone || '', created_at: jp.iat ? new Date(jp.iat * 1000).toISOString() : '' };
      persist();
    }
  }

  return { state, isLoggedIn, getUser, getToken, isExpired, login, logout, hasPermission, refreshToken, parseJwt };
});
