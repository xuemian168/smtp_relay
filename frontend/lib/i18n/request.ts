import { notFound } from 'next/navigation';
import { getRequestConfig } from 'next-intl/server';
import { locales } from './config';

export default getRequestConfig(async ({ requestLocale }) => {
  // 获取当前请求的语言
  const locale = await requestLocale;
  
  // 验证传入的语言是否在支持列表中
  if (!locale || !locales.includes(locale as any)) notFound();

  return {
    locale,
    messages: (await import(`../../messages/${locale}.json`)).default
  };
}); 