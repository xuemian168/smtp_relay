import axios, { AxiosError, AxiosRequestConfig, AxiosResponse } from 'axios';
import { useAuthStore } from '../stores/auth';

// API基础配置
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// 创建axios实例
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
    'Accept-Language': 'zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6',
    'withCredentials': true,
  },
});

// 请求拦截器 - 添加认证token
apiClient.interceptors.request.use(
  (config) => {
    // 从Zustand store获取token
    const token = useAuthStore.getState().token;
    
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    
    // 添加请求时间戳用于调试
    config.metadata = { requestStartTime: Date.now() };
    
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器 - 处理错误和token过期
apiClient.interceptors.response.use(
  (response: AxiosResponse) => {
    // 记录响应时间用于性能监控
    const requestStartTime = response.config.metadata?.requestStartTime;
    if (requestStartTime) {
      const duration = Date.now() - requestStartTime;
      console.log(`API响应时间: ${duration}ms - ${response.config.url}`);
    }
    
    return response;
  },
  (error: AxiosError) => {
    // 处理401错误 - token过期或无效
    if (error.response?.status === 401) {
      const authStore = useAuthStore.getState();
      authStore.logout();
      
      // 如果不是登录页面，重定向到登录页
      if (typeof window !== 'undefined' && !window.location.pathname.includes('/login')) {
        window.location.href = '/login';
      }
    }
    
    // 处理网络错误
    if (!error.response) {
      console.error('网络错误:', error.message);
    }
    
    // 格式化错误响应，支持国际化
    const errorResponse = {
      message: (error.response?.data as any)?.message || error.message || '请求失败',
      status: error.response?.status,
      code: (error.response?.data as any)?.code,
      details: (error.response?.data as any)?.details,
    };
    
    return Promise.reject(errorResponse);
  }
);

// 自定义axios函数，供Orval使用
export const customAxios = <T = any>(config: AxiosRequestConfig): Promise<T> => {
  return apiClient(config).then((response) => response.data);
};

// 导出配置好的axios实例
export default apiClient; 