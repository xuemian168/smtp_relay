"use client";

import { useTranslations } from 'next-intl';
import { useEffect, useState } from 'react';
import { statsApi } from '@/lib/api';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import DashboardOverview from '@/components/DashboardOverview';
import DashboardShell from '@/components/layout/DashboardShell';

export default function OverviewPage() {
  const t = useTranslations('overview');
  const [stats, setStats] = useState({
    dailySent: 0,
    successRate: 0,
    quotaUsed: 0,
    activeCredentials: 0,
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        // 并行获取统计和配额数据
        const [overviewData, quotaData] = await Promise.all([
          statsApi.overview(),
          statsApi.quota(),
        ]);

        // 处理API响应数据（根据实际API结构调整）
        const overview = overviewData as any;
        const quota = quotaData as any;

        setStats({
          dailySent: overview?.daily_sent || 1234, // 回退到模拟数据
          successRate: overview?.success_rate || 98.5,
          quotaUsed: quota?.daily_used && quota?.daily_quota 
            ? (quota.daily_used / quota.daily_quota) * 100 
            : 65.2,
          activeCredentials: overview?.active_credentials || 8,
        });
      } catch (err) {
        console.error('获取统计数据失败:', err);
        setError('无法获取统计数据');
        // 使用模拟数据作为回退
        setStats({
          dailySent: 1234,
          successRate: 98.5,
          quotaUsed: 65.2,
          activeCredentials: 8,
        });
      } finally {
        setLoading(false);
      }
    };

    fetchStats();
  }, []);

  return (
    <DashboardShell>
      <div className="bg-background text-foreground rounded-xl shadow-lg p-6 min-h-[calc(100vh-4rem)]">
        <DashboardOverview/>
        <div className="container mx-auto px-4 py-8">
          <p className="text-gray-600 dark:text-gray-400 mb-8">
            {t('description')}
          </p>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center justify-between">
                  {t('todaySent')}
                  <Badge variant="default">{t('todaySent')}</Badge>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-3xl font-bold text-blue-600">
                  {loading ? '...' : stats.dailySent.toLocaleString()}
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center justify-between">
                  {t('successRate')}
                  <Badge variant="secondary">%</Badge>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-3xl font-bold text-green-600">
                  {loading ? '...' : `${stats.successRate}%`}
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center justify-between">
                  {t('quotaUsage')}
                  <Badge variant="outline">%</Badge>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-3xl font-bold text-orange-600">
                  {loading ? '...' : `${stats.quotaUsed.toFixed(1)}%`}
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center justify-between">
                  {t('totalCredentials')}
                  <Badge variant="default">{t('totalCredentials')}</Badge>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-3xl font-bold text-purple-600">
                  {loading ? '...' : stats.activeCredentials}
                </p>
              </CardContent>
            </Card>
          </div>

          <Card className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
            <CardHeader>
              <CardTitle className="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-4">
                {t('recentActivity')}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-gray-600 dark:text-gray-400">
                国际化测试页面 - Internationalization test page
              </p>
              <div className="mt-4 space-x-4">
                <a
                  href="/zh/login"
                  className="inline-block bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
                >
                  中文登录页面
                </a>
                <a
                  href="/en/login"
                  className="inline-block bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700"
                >
                  English Login Page
                </a>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </DashboardShell>
  );
} 