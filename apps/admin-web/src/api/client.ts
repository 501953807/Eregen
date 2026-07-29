import axios from 'axios';
import qs from 'qs'; // For proper form encoding

// Create API client with secure defaults
const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'https://localhost:8080/api/v1', // Default to HTTPS
  timeout: 10000,
  withCredentials: true, // Crucial: sends cookies including httpOnly tokens
  headers: {
    'Content-Type': 'application/json',
    'X-Requested-With': 'XMLHttpRequest', // Helps detect AJAX requests
  },
});

// CSRF token handling - get from x-csrf-token header or cookie
function getCSRFToken() {
  // Check for X-CSRF-Token in headers set by server
  const meta = document.querySelector('meta[name="csrf-token"]');
  return meta ? meta.content : null;
}

// Request interceptor – attach CSRF token and ensure HTTPS
apiClient.interceptors.request.use((config) => {
  // Ensure we're using HTTPS in production
  if (import.meta.env.PRODUCTION && !config.baseURL?.startsWith('https://')) {
    console.warn('API URL should use HTTPS in production');
  }

  // Attach httpOnly session cookie via withCredentials (automatically sent)
  config.withCredentials = true;

  // Add CSRF token if available
  const csrfToken = getCSRFToken();
  if (csrfToken) {
    config.headers['X-CSRF-Token'] = csrfToken;
    // Also send as custom header for non-simple methods
    config.headers['X-CSRF-Token'] = csrfToken;
  }

  return config;
});

// Response interceptor – handle 401/403 errors
apiClient.interceptors.response.use(
  (res) => res,
  (err) => {
    const status = err.response?.status;

    if (status === 401 || status === 403) {
      // Clear any cached CSRF token (may be stale)
      const meta = document.querySelector('meta[name="csrf-token"]');
      if (meta) meta.remove();

      // Redirect to login page
      window.location.href = '/login';
      return Promise.reject(err);
    }

    if (status === 403 && err.response?.headers['x-csrf-token']) {
      // Stale CSRF token, refresh it
      const newCsrf = err.response.headers['x-csrf-token'];
      const meta = document.createElement('meta');
      meta.name = 'csrf-token';
      meta.content = newCsrf;
      document.head.appendChild(meta);
    }

    return Promise.reject(err);
  }
);

// Export utility for fetching fresh CSRF token on login
export function setupCSRFTokenFromResponse(response) {
  if (response.headers['x-csrf-token']) {
    const meta = document.querySelector('meta[name="csrf-token"]');
    if (meta) {
      meta.content = response.headers['x-csrf-token'];
    } else {
      const meta = document.createElement('meta');
      meta.name = 'csrf-token';
      meta.content = response.headers['x-csrf-token'];
      document.head.appendChild(meta);
    }
  }
}

export default apiClient;
