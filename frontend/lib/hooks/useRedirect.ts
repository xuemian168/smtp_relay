import { useRouter } from 'next/navigation';
import { useLocale } from 'next-intl';

export const useRedirect = () => {
  const router = useRouter();
  const locale = useLocale();

  // 重定向到指定路径，保持当前语言
  const redirectTo = (path: string, replace = false) => {
    const localizedPath = `/${locale}${path}`;
    if (replace) {
      router.replace(localizedPath);
    } else {
      router.push(localizedPath);
    }
  };

  // 登录成功后重定向逻辑
  const redirectAfterLogin = () => {
    // 检查是否有重定向参数
    const urlParams = new URLSearchParams(window.location.search);
    const redirectParam = urlParams.get('redirect');
    
    if (redirectParam) {
      // 如果有重定向参数，重定向到指定页面
      redirectTo(redirectParam, true);
    } else {
      // 默认重定向到概览页面
      redirectTo('/overview', true);
    }
  };

  // 需要认证的页面重定向到登录页
  const redirectToLogin = (returnUrl?: string) => {
    const loginPath = '/login';
    if (returnUrl) {
      redirectTo(`${loginPath}?redirect=${encodeURIComponent(returnUrl)}`, true);
    } else {
      redirectTo(loginPath, true);
    }
  };

  // 已登录用户访问登录页时重定向到首页
  const redirectIfAuthenticated = () => {
    redirectTo('/overview', true);
  };

  return {
    redirectTo,
    redirectAfterLogin,
    redirectToLogin,
    redirectIfAuthenticated,
  };
}; 