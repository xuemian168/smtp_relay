'use client';

import { useEffect } from 'react';
import { useTranslations } from 'next-intl';
import { useAuth } from '@/lib/api/auth';
import { useRedirect } from '@/lib/hooks/useRedirect';
import LoginForm from '@/components/auth/LoginForm';

export default function LoginPage() {
  const t = useTranslations('auth');
  const { isAuthenticated } = useAuth();
  const { redirectIfAuthenticated } = useRedirect();

  // 如果已经登录，重定向到首页
  useEffect(() => {
    if (isAuthenticated) {
      redirectIfAuthenticated();
    }
  }, [isAuthenticated, redirectIfAuthenticated]);

  // 如果已登录，不渲染登录表单
  if (isAuthenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
        <div className="text-center">
          <div className="animate-spin h-8 w-8 border-b-2 border-blue-600 rounded-full mx-auto"></div>
          <p className="mt-2 text-gray-600 dark:text-gray-400">正在重定向...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900 dark:text-gray-100">
            {t('loginToAccount')}
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600 dark:text-gray-400">
            {t('loginDescription')}
          </p>
        </div>
        <div className="bg-white dark:bg-gray-800 py-8 px-4 shadow sm:rounded-lg sm:px-10">
          <LoginForm />
        </div>
      </div>
    </div>
  );
} 