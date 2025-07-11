import createMiddleware from 'next-intl/middleware';
import { locales, defaultLocale } from './lib/i18n/config';

export default createMiddleware({
  // 支持的语言列表
  locales,
  
  // 默认语言
  defaultLocale,
  
  // 路径前缀策略
  localePrefix: 'as-needed'
});

export const config = {
  // 匹配所有路径，但排除API路由、静态文件等
  matcher: ['/((?!api|_next|_vercel|.*\\..*).*)']
}; 