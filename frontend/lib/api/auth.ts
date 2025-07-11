import { useAuthStore } from '../stores/auth';
import { LoginRequest, LoginResponse, ApiErrorResponse } from './types';

// Token相关工具函数
export class AuthUtils {
  // 检查token是否过期
  static isTokenExpired(token: string): boolean {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      const currentTime = Date.now() / 1000;
      return payload.exp < currentTime;
    } catch {
      return true;
    }
  }

  // 从token中提取用户信息
  static extractUserFromToken(token: string): any | null {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      return payload;
    } catch {
      return null;
    }
  }

  // 检查当前用户是否已认证
  static isAuthenticated(): boolean {
    const { token, isAuthenticated } = useAuthStore.getState();
    if (!token || !isAuthenticated) return false;
    return !this.isTokenExpired(token);
  }

  // 获取当前用户token
  static getToken(): string | null {
    const { token } = useAuthStore.getState();
    if (token && !this.isTokenExpired(token)) {
      return token;
    }
    return null;
  }

  // 清理过期的认证状态
  static clearExpiredAuth(): void {
    const { token, logout } = useAuthStore.getState();
    if (token && this.isTokenExpired(token)) {
      logout();
    }
  }
}

// 认证相关的Hook函数
export const useAuth = () => {
  const authStore = useAuthStore();

  // 登录函数
  const login = async (credentials: LoginRequest): Promise<LoginResponse | ApiErrorResponse> => {
    authStore.setLoading(true);
    
    try {
      // 使用生成的API客户端
      const { authApi } = await import('./index');
      const response = await authApi.login({
        username: credentials.username,
        password: credentials.password,
        remember_me: credentials.remember_me ? 'true' : 'false',
      });

      // 假设API返回的数据结构（需要根据实际API调整）
      const loginData: LoginResponse = {
        success: true,
        token: (response as any).token || 'mock-jwt-token',
        user: (response as any).user || {
          id: '1',
          username: credentials.username,
          email: credentials.username + '@example.com',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
        expires_at: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
      };
      
      // 更新认证状态
      authStore.login(loginData.token, loginData.user);
      
      return loginData;
    } catch (error) {
      const errorResponse: ApiErrorResponse = {
        message: error instanceof Error ? error.message : '登录失败',
        status: 401,
      };
      return errorResponse;
    } finally {
      authStore.setLoading(false);
    }
  };

  // 登出函数
  const logout = () => {
    authStore.logout();
    // 清理任何其他本地存储的数据
    if (typeof window !== 'undefined') {
      // 可以在这里清理其他缓存
      localStorage.removeItem('api-cache');
    }
  };

  // 检查认证状态
  const checkAuth = () => {
    AuthUtils.clearExpiredAuth();
    return AuthUtils.isAuthenticated();
  };

  return {
    // 状态
    token: authStore.token,
    user: authStore.user,
    isAuthenticated: authStore.isAuthenticated,
    isLoading: authStore.isLoading,
    
    // 操作
    login,
    logout,
    checkAuth,
    updateUser: authStore.updateUser,
    setLoading: authStore.setLoading,
  };
};

// 导出认证状态检查中间件，用于保护路由
export const requireAuth = (callback: () => void) => {
  if (!AuthUtils.isAuthenticated()) {
    if (typeof window !== 'undefined') {
      window.location.href = '/login';
    }
    return;
  }
  callback();
}; 