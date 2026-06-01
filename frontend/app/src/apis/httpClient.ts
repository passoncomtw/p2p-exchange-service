import axios, { AxiosInstance, AxiosResponse } from 'axios';
import logger from '@pkg/logger';

// 使用 Expo 環境變數，從 .env 讀取 BASE_URL
// Expo 需要使用 EXPO_PUBLIC_ 前綴才能在 JavaScript 中使用
const baseURL =
  process.env.EXPO_PUBLIC_API_BASE_URL || 'https://token-app-api.passon.tw';

// 儲存導航引用（需要在 App 初始化時設定）
let navigationRef: any = null;

// 儲存 Redux store 引用（需要在 App 初始化時設定）
let storeRef: any = null;

/**
 * 設定導航引用（用於未授權時導航到登入頁）
 */
export function setNavigationRef(ref: any) {
  navigationRef = ref;
}

/**
 * 設定 Redux store 引用（用於從 state 讀取 token）
 */
export function setStoreRef(store: any) {
  storeRef = store;
}

// HTTP 客戶端配置
const httpClient: AxiosInstance = axios.create({
  baseURL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 請求攔截器
httpClient.interceptors.request.use(
  (config) => {
    // 從 Redux store 讀取 token
    try {
      if (storeRef) {
        const state = storeRef.getState();
        const token = state.auth?.accessToken || null;

        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
          logger.debug('HTTP Request - Authorization header set', {
            url: config.url,
            method: config.method,
            hasToken: !!token,
          });
        }
      }
    } catch (error) {
      logger.warn('HTTP Request - Failed to get token from store', { error });
    }
    
    // 記錄請求
    logger.info(`HTTP ${config.method?.toUpperCase()} ${config.url}`);
    
    return config;
  },
  (error) => {
    logger.warn('HTTP Request error', { error });
    return Promise.reject(error);
  }
);

// 回應攔截器
httpClient.interceptors.response.use(
  (response: AxiosResponse) => {
    // 記錄成功回應
    logger.info(`HTTP ${response.status} ${response.config.url}`);
    return response;
  },
  (error) => {
    // 統一錯誤處理
    logger.warn('HTTP Response error', {
      status: error.response?.status,
      url: error.config?.url,
      message: error.message,
    });

    const isAuthLogin =
      typeof error?.config?.url === 'string' &&
      (error.config.url.includes('/auth/login') ||
       error.config.url.includes('/api/v1/auth/login'));

    if (error.response?.status === 401 && !isAuthLogin) {
      // 未授權時：導航到登入頁
      logger.warn('HTTP 401 Unauthorized - Redirecting to login', {
        url: error.config?.url,
      });

      if (navigationRef) {
        import('@react-navigation/native').then(({ CommonActions }) => {
          navigationRef.dispatch(
            CommonActions.reset({
              index: 0,
              routes: [{ name: 'Login' }],
            })
          );
        }).catch((navError) => {
          logger.warn('HTTP Failed to navigate to login', { error: navError });
        });
      }
    }

    return Promise.reject(error);
  }
);

/**
 * 檢查 token 是否存在
 * @throws Error 如果 token 不存在
 */
function ensureToken(): string {
  if (!storeRef) {
    const error = new Error('[HTTP] Store reference not set. Please call setStoreRef() first.');
    logger.warn('ensureToken failed', { error: error.message });
    throw error;
  }

  const state = storeRef.getState();
  const token = state.auth?.accessToken;
  
  logger.debug('ensureToken check', {
    hasToken: !!token,
    isAuthenticated: state.auth?.isAuthenticated,
  });

  if (!token) {
    const error = new Error('[HTTP] No authentication token available. Please login first.');
    logger.warn('ensureToken failed', { error: error.message });
    throw error;
  }

  return token;
}

/**
 * 需要認證的 HTTP 方法
 * 這些方法會在發送請求前檢查 token 是否存在
 */
export const httpClientWithAuth = {
  /**
   * GET 請求（需要 token）
   * @throws Error 如果 token 不存在
   */
  getWithToken: <T = any>(url: string, config?: any) => {
    ensureToken(); // 確保 token 存在
    return httpClient.get<T>(url, config);
  },

  /**
   * POST 請求（需要 token）
   * @throws Error 如果 token 不存在
   */
  postWithToken: <T = any>(url: string, data?: any, config?: any) => {
    ensureToken();
    return httpClient.post<T>(url, data, config);
  },

  /**
   * PUT 請求（需要 token）
   * @throws Error 如果 token 不存在
   */
  putWithToken: <T = any>(url: string, data?: any, config?: any) => {
    ensureToken();
    return httpClient.put<T>(url, data, config);
  },

  /**
   * PATCH 請求（需要 token）
   * @throws Error 如果 token 不存在
   */
  patchWithToken: <T = any>(url: string, data?: any, config?: any) => {
    ensureToken();
    return httpClient.patch<T>(url, data, config);
  },

  /**
   * DELETE 請求（需要 token）
   * @throws Error 如果 token 不存在
   */
  deleteWithToken: <T = any>(url: string, config?: any) => {
    ensureToken();
    return httpClient.delete<T>(url, config);
  },
};

export default httpClient;

