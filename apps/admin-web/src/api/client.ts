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
  (response: AxiosResponse<any>): any => {
    const { code, msg, data } = response.data || {};
    // Handle 4xx/5xx even when code field is missing (e.g. {"error":"msg"})
    if (response.status >= 400) {
      const errMsg = msg || response.data?.error || '业务请求失败，请重试';
      if (response.status === 401) {
        const authStore = useAuthStore();
        authStore.logout();
        ElMessage.warning('会话已过期，请重新登录');
        const redirectPath = window.location.pathname;
        router.push({ path: '/login', query: { redirect: redirectPath } });
        return Promise.reject({ code: 401, msg: '未授权', data: null });
      }
      ElMessage.error(errMsg);
      return Promise.reject({ code: response.status, msg: errMsg, data: null });
    }
    return { code, msg: msg || '成功', data };
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
