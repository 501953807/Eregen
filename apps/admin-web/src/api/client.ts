import axios, { type AxiosInstance, AxiosResponse, AxiosError } from 'axios';
import { useAuthStore } from '@/stores/auth';
import { ElMessage } from 'element-plus';
import router from '@/router';

const baseURL = import.meta.env.VITE_API_BASE_URL
  ? import.meta.env.VITE_API_BASE_URL
  : (import.meta.env.DEV ? 'http://localhost:8089' : '/api/v1');

const apiClient: AxiosInstance = axios.create({
  baseURL,
  timeout: 30000,
});

const parseJwt = (token: string): Record<string, any> | null => {
  try {
    const base64Url = token.split('.')[1];
    if (!base64Url) return null;
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(atob(base64).split('').map(c => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)).join(''));
    return JSON.parse(jsonPayload);
  } catch (e) {
    return null;
  }
};

apiClient.interceptors.request.use(
  async (config) => {
    const authStore = useAuthStore();
    const token = authStore.getToken();

    if (token && config.headers) {
      config.headers['Authorization'] = `Bearer ${token}`;
    }

    if (config.method === 'post' || config.method === 'put') {
      config.headers['Content-Type'] = 'application/json';
    }

    return config;
  },
  (error) => Promise.reject(error)
);

apiClient.interceptors.response.use(
  (response: AxiosResponse<any>) => {
    const { code, msg, data } = response.data || {};
    if ((response.status >= 200 && response.status < 300) && (!code || (code >= 200 && code < 300))) {
      return { code, msg: msg || '成功', data };
    }
    ElMessage.error(msg || '业务请求失败，请重试');
    return Promise.reject({ code, msg: msg || '业务失败', data });
  },
  (error: AxiosError) => {
    const authStore = useAuthStore();

    if (!error.response) {
      ElMessage.error('网络错误，请检查网络连接');
      return Promise.reject({ code: 0, msg: '网络错误', data: null });
    }

    const status = error.response.status;
    const resp = error.response.data as { code?: number; msg?: string; data?: any };

    if (status === 401 || resp.code === 401) {
      authStore.logout();
      ElMessage.warning('会话已过期，请重新登录');
      const redirectPath = error.config?.params?.redirect || window.location.pathname;
      return router.push({ path: '/login', query: { redirect: redirectPath } });
    }

    if (status === 403 || resp.code === 403) {
      ElMessage.error('无权访问此资源');
      return Promise.reject({ code: 403, msg: '权限不足', data: null });
    }

    const errorMsg = resp?.msg || error.response?.statusText || '请求失败';
    ElMessage.error(errorMsg);
    return Promise.reject({ code: status || (resp.code || 0), msg: errorMsg, data: resp?.data });
  }
);

export default apiClient;

export async function apiCall<T>(func: () => Promise<ApiResponse<T>>): Promise<T> {
  try {
    const res = await func();
    if ((res.code && (res.code < 200 || res.code >= 300)) && (!res.code || res.code < 200)) {
      throw new Error(res.msg || 'API call failed');
    }
    return res.data;
  } catch (err) {
    console.error('API call error:', err);
    throw err;
  }
}