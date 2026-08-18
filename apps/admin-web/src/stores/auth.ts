import { defineStore } from 'pinia';
import { ref, computed, watch } from 'vue';
import router from '@/router';
import type { User, LoginResponse, AuthState, LoginRequest } from '@/types';
import { parseJwt } from '@/utils/auth';
import apiClient, { setCsrfToken } from '@/api/client';

const STORAGE_KEY = 'eregen_admin_auth_state';

export const useAuthStore = defineStore('auth', () => {

  const stored = localStorage.getItem(STORAGE_KEY);
  let initialState: AuthState = { token: null, user: null, expiresAt: null };
  if (stored) {
    const parts = stored.split(':::');
    initialState = {
      token: parts[0] || null,
      user: parts[1] === 'null' ? null : (parts[1] ? JSON.parse(parts[1]) : null),
      expiresAt: parts[2] ? Number(parts[2]) : null,
    };
  }

  const state = ref<AuthState>(initialState);
  const loading = ref(false);
  const error = ref<string | null>(null);

  const isExpired = computed(() => state.value.expiresAt && Date.now() >= state.value.expiresAt * 1000);
  const isLoggedIn = computed(() => !!state.value.token && !isExpired.value);
  const user = computed(() => state.value.user);
  const getUser = computed(() => state.value.user || null);
  const getToken = () => state.value.token;
  const checkLoggedIn = () => isLoggedIn.value;

  const persist = () => {
    const uStr = state.value.user ? JSON.stringify(state.value.user) : 'null';
    const eStr = state.value.expiresAt ? state.value.expiresAt.toString() : '';
    localStorage.setItem(STORAGE_KEY, `${state.value.token || ''}:::${uStr}:::${eStr}`);
  };

  watch(state.value, () => { persist(); }, { deep: true });

  const login = async (input: LoginRequest | LoginResponse) => {
    loading.value = true;
    error.value = null;
    try {
      let token: string;
      let userId: string;
      let role: string;

      if ('identifier' in input) {
        const res = await apiClient.post('/auth/login', {
          method: 'email',
          credential: input.identifier,
          secret: input.password,
        }, { timeout: 10000 });
        // Backend response: { code: "OK", data: { token: "...", user: {...} } }
        const respData = res as any;
        token = respData?.data?.token || respData?.access_token;
        const backendUser = respData?.data?.user || {};
        userId = backendUser.id || respData?.data?.user_id || '';
        role = backendUser.role || respData?.role || 'admin';
      } else {
        token = input.access_token;
        userId = input.user_id;
        role = input.role;
      }

      const exp = Math.floor(Date.now() / 1000) + (token ? parseJwt(token)?.exp || 7200 : 7200);
      state.value = {
        token,
        user: {
          id: userId,
          name: role === 'super_admin' ? '管理员' : '用户',
          role: role as any,
          phone: '',
          created_at: '',
        },
        expiresAt: Number(exp),
      };
      persist();
      const rd = (router.currentRoute.value.query.redirect as string) || '/dashboard';
      router.push({ path: rd });
      return state.value.user;
    } catch (e: any) {
      error.value = e?.response?.data?.msg || e?.response?.data?.message || '登录失败';
      throw e;
    } finally {
      loading.value = false;
    }
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
      // Token expired — clear both state and storage
      state.value = { token: null, user: null, expiresAt: null };
      localStorage.removeItem(STORAGE_KEY);
    } else if ((jp?.user_id || jp?.sub) && !state.value.user) {
      state.value.user = {
        id: (jp?.user_id || jp?.sub || '') as string,
        name: (jp?.name || '用户') as string,
        role: (jp?.role || 'family') as any,
        phone: '',
        created_at: '',
      };
      persist();
    }
  } else if (state.value.user) {
    // Token was cleared but user remained in state — reset user too
    state.value.user = null;
  }

  return { state, user, isLoggedIn, getUser, getToken, checkLoggedIn, isExpired, loading, error, login, logout, hasPermission, refreshToken, parseJwt };
});
