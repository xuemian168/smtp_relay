import { redirect } from 'next/navigation';

type Props = {
  params: { locale: string };
};

export default function LocaleHomePage({ params: { locale } }: Props) {
  // 重定向到概览页面
  redirect(`/${locale}/overview`);
} 