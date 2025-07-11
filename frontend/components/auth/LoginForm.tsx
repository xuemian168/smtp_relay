'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { useTranslations } from 'next-intl';
import { useAuth } from '@/lib/api/auth';
import { useRedirect } from '@/lib/hooks/useRedirect';
import { LoginRequest, ApiErrorResponse } from '@/lib/api/types';

interface LoginFormData {
  username: string;
  password: string;
  rememberMe: boolean;
}

export default function LoginForm() {
  const t = useTranslations('auth');
  const tValidation = useTranslations('validation');
  const { login, isLoading } = useAuth();
  const { redirectAfterLogin } = useRedirect();
  const [loginError, setLoginError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormData>({
    defaultValues: {
      username: '',
      password: '',
      rememberMe: false,
    },
  });

  const onSubmit = async (data: LoginFormData) => {
    try {
      setLoginError(null);
      
      const credentials: LoginRequest = {
        username: data.username,
        password: data.password,
        remember_me: data.rememberMe,
      };

      const result = await login(credentials);

      // 检查登录结果
      if ('success' in result && result.success) {
        // 登录成功，重定向
        redirectAfterLogin();
      } else {
        // 登录失败，显示错误信息
        const errorResult = result as ApiErrorResponse;
        setLoginError(errorResult.message || t('loginFailed'));
      }
    } catch (error) {
      console.error('登录表单错误:', error);
      setLoginError(t('loginError'));
    }
  };

  const isFormLoading = isLoading || isSubmitting;

  return (
    <form className="space-y-6" onSubmit={handleSubmit(onSubmit)}>
      {/* 错误提示 */}
      {loginError && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-md dark:bg-red-900/20 dark:border-red-800 dark:text-red-400">
          {loginError}
        </div>
      )}

      {/* 用户名字段 */}
      <div>
        <label 
          htmlFor="username" 
          className="block text-sm font-medium text-gray-700 dark:text-gray-300"
        >
          {t('usernameLabel')}
        </label>
        <div className="mt-1">
          <input
            id="username"
            type="text"
            autoComplete="username"
            disabled={isFormLoading}
            className="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm dark:bg-gray-700 dark:border-gray-600 dark:text-white disabled:opacity-50 disabled:cursor-not-allowed"
            {...register('username', {
              required: t('usernameRequired'),
              minLength: {
                value: 2,
                message: tValidation('minLength', { min: 2 }),
              },
            })}
          />
          {errors.username && (
            <p className="mt-1 text-sm text-red-600 dark:text-red-400">
              {errors.username.message}
            </p>
          )}
        </div>
      </div>

      {/* 密码字段 */}
      <div>
        <label 
          htmlFor="password" 
          className="block text-sm font-medium text-gray-700 dark:text-gray-300"
        >
          {t('passwordLabel')}
        </label>
        <div className="mt-1">
          <input
            id="password"
            type="password"
            autoComplete="current-password"
            disabled={isFormLoading}
            className="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm dark:bg-gray-700 dark:border-gray-600 dark:text-white disabled:opacity-50 disabled:cursor-not-allowed"
            {...register('password', {
              required: t('passwordRequired'),
              minLength: {
                value: 6,
                message: tValidation('minLength', { min: 6 }),
              },
            })}
          />
          {errors.password && (
            <p className="mt-1 text-sm text-red-600 dark:text-red-400">
              {errors.password.message}
            </p>
          )}
        </div>
      </div>

      {/* 记住我复选框 */}
      <div className="flex items-center justify-between">
        <div className="flex items-center">
          <input
            id="remember-me"
            type="checkbox"
            disabled={isFormLoading}
            className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded disabled:opacity-50"
            {...register('rememberMe')}
          />
          <label 
            htmlFor="remember-me" 
            className="ml-2 block text-sm text-gray-900 dark:text-gray-300"
          >
            {t('rememberMe')}
          </label>
        </div>
      </div>

      {/* 提交按钮 */}
      <div>
        <button
          type="submit"
          disabled={isFormLoading}
          className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:bg-blue-600"
        >
          {isFormLoading ? (
            <div className="flex items-center">
              <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              {t('loggingIn')}
            </div>
          ) : (
            t('login')
          )}
        </button>
      </div>
    </form>
  );
} 