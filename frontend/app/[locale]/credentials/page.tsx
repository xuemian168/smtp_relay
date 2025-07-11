"use client";
import { useTranslations } from 'next-intl';
import CredentialList from '@/components/credentials/CredentialList';
import DashboardShell from '@/components/layout/DashboardShell';

export default function CredentialsPage() {
  const t = useTranslations('credentials');

  return (
    <DashboardShell>
      <div className="bg-background text-foreground rounded-xl shadow-lg p-6 min-h-[calc(100vh-4rem)]">
        <div className="flex items-center justify-between mb-4">
          <h1 className="text-2xl font-bold">{t('title')}</h1>
        </div>
        <CredentialList />
      </div>
    </DashboardShell>
  );
} 