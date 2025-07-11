"use client";
import { useTranslations } from 'next-intl';
import MailLogList from '@/components/maillog/mailloglist';
import DashboardShell from '@/components/layout/DashboardShell';

export default function MailLogsPage() {
  const t = useTranslations('maillogs');

  return (
    <DashboardShell>
      <div className="bg-background text-foreground rounded-xl shadow-lg p-6 min-h-[calc(100vh-4rem)]">
        <div className="flex items-center justify-between mb-4">
          <h1 className="text-2xl font-bold">{t('title')}</h1>
        </div>
        <MailLogList />
      </div>
    </DashboardShell>
  );
} 