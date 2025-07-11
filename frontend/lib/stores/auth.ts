import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { AuthUser } from '../api/types';

interface AuthState {
  token: string | null;
  user: AuthUser | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

interface AuthActions {
  login: (token: string, user: AuthUser) => void;
  logout: () => void;
  setLoading: (loading: boolean) => void;
  updateUser: (user: Partial<AuthUser>) => void;
}

type AuthStore = AuthState & AuthActions;

export const useAuthStore = create<AuthStore>()(
  persist(
    (set, get) => ({
      // 初始状态
      token: null,
      user: null,
      isAuthenticated: false,
      isLoading: false,

      // 登录操作
      login: (token: string, user: AuthUser) => {
        set({
          token,
          user,
          isAuthenticated: true,
          isLoading: false,
        });
      },

      // 登出操作
      logout: () => {
        set({
          token: null,
          user: null,
          isAuthenticated: false,
          isLoading: false,
        });
      },

      // 设置加载状态
      setLoading: (loading: boolean) => {
        set({ isLoading: loading });
      },

      // 更新用户信息
      updateUser: (userData: Partial<AuthUser>) => {
        const currentUser = get().user;
        if (currentUser) {
          set({
            user: { ...currentUser, ...userData },
          });
        }
      },
    }),
    {
      name: 'auth-storage', // localStorage key
      storage: createJSONStorage(() => localStorage),
      // 只持久化token和user，不持久化loading状态
      partialize: (state) => ({
        token: state.token,
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
); 