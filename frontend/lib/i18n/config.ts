// 支持的语言列表
export const locales = ['zh', 'en'] as const;
export type Locale = (typeof locales)[number];

// 默认语言
export const defaultLocale: Locale = 'zh';

// 语言显示名称
export const localeNames: Record<Locale, string> = {
  zh: '中文',
  en: 'English',
};

// next-intl配置文件将在安装依赖后单独配置 