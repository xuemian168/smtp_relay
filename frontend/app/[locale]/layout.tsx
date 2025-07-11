import { NextIntlClientProvider } from 'next-intl';
import { getMessages } from 'next-intl/server';
import { notFound } from 'next/navigation';
import { locales } from '@/lib/i18n/config';
import { ThemeProvider } from '@/components/layout/ThemeProvider';
import { Toaster } from '@/components/ui/toaster';
import React from 'react';

type Props = {
  children: React.ReactNode;
  params: Promise<{ locale: string }>;
};

export default async function LocaleLayout({
  children,
  params
}: Props) {
  const { locale } = await params;
  
  // 验证语言是否有效
  if (!locales.includes(locale as any)) {
    notFound();
  }

  // 获取当前语言的消息
  const messages = await getMessages();

  return (
    <NextIntlClientProvider messages={messages}>
      <ThemeProvider>
        {children}
        <Toaster />
      </ThemeProvider>
    </NextIntlClientProvider>
  );
} 