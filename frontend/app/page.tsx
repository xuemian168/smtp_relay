import { redirect } from 'next/navigation';

export default function RootPage() {
  // 重定向到默认语言的概览页面
  redirect('/zh/overview');
} 