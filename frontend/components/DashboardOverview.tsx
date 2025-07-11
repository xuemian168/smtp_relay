
import React from 'react';
import { Mail, TrendingUp, AlertTriangle, Eye, Activity } from 'lucide-react';
import { LineChart, Line, AreaChart, Area, XAxis, YAxis, CartesianGrid, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import StatCard from './StatCard';
import { useTranslations } from 'next-intl';

const DashboardOverview = () => {
  const t = useTranslations('overview');

  const emailData = [
    { name: 'Mon', sent: 4000, delivered: 3800, opened: 1900 },
    { name: 'Tue', sent: 3000, delivered: 2900, opened: 1400 },
    { name: 'Wed', sent: 2000, delivered: 1950, opened: 980 },
    { name: 'Thu', sent: 2780, delivered: 2700, opened: 1350 },
    { name: 'Fri', sent: 1890, delivered: 1800, opened: 900 },
    { name: 'Sat', sent: 2390, delivered: 2300, opened: 1150 },
    { name: 'Sun', sent: 3490, delivered: 3350, opened: 1675 },
  ];

  const statusData = [
    { name: 'Delivered', value: 85, color: '#10b981' },
    { name: 'Bounced', value: 8, color: '#ef4444' },
    { name: 'Pending', value: 5, color: '#f59e0b' },
    { name: 'Failed', value: 2, color: '#6b7280' },
  ];

  const recentActivities = [
    { id: 1, type: 'success', message: 'Welcome email sent to user@example.com', time: '2 minutes ago' },
    { id: 2, type: 'warning', message: 'High bounce rate detected for campaign "Newsletter #42"', time: '15 minutes ago' },
    { id: 3, type: 'info', message: 'New API key generated', time: '1 hour ago' },
    { id: 4, type: 'success', message: 'Domain verification completed for example.com', time: '2 hours ago' },
  ];

  const getActivityIcon = (type: string) => {
    switch (type) {
      case 'success': return <div className="w-2 h-2 bg-green-500 rounded-full" />;
      case 'warning': return <div className="w-2 h-2 bg-yellow-500 rounded-full" />;
      case 'error': return <div className="w-2 h-2 bg-red-500 rounded-full" />;
      default: return <div className="w-2 h-2 bg-blue-500 rounded-full" />;
    }
  };

  return (
    <div className="space-y-6">
      {/* Welcome Section */}
      <div className="bg-gradient-to-r from-blue-600 to-purple-600 rounded-xl p-6 text-white">
        <h1 className="text-2xl font-bold mb-2">{t('welcome')}</h1>
        <p className="text-blue-100">{t('manageEmailDelivery')}</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title={t('totalSent')}
          value="24,567"
          change="+12.5% from last week"
          changeType="positive"
          icon={Mail}
          gradient="from-blue-600 to-blue-700"
        />
        <StatCard
          title={t('deliveryRate')}
          value="98.2%"
          change="+0.3% from last week"
          changeType="positive"
          icon={TrendingUp}
          gradient="from-green-600 to-green-700"
        />
        <StatCard
          title={t('bounceRate')}
          value="1.8%"
          change="-0.2% from last week"
          changeType="positive"
          icon={AlertTriangle}
          gradient="from-yellow-600 to-orange-600"
        />
        <StatCard
          title={t('openRate')}
          value="45.6%"
          change="+2.1% from last week"
          changeType="positive"
          icon={Eye}
          gradient="from-purple-600 to-purple-700"
        />
      </div>

      {/* Charts Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Activity className="h-5 w-5" />
              <span>{t('sendingVolume')}</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-80">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={emailData}>
                  <defs>
                    <linearGradient id="colorSent" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.8}/>
                      <stop offset="95%" stopColor="#3b82f6" stopOpacity={0.1}/>
                    </linearGradient>
                    <linearGradient id="colorDelivered" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#10b981" stopOpacity={0.8}/>
                      <stop offset="95%" stopColor="#10b981" stopOpacity={0.1}/>
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
                  <XAxis dataKey="name" className="text-sm" />
                  <YAxis className="text-sm" />
                  <Area
                    type="monotone"
                    dataKey="sent"
                    stroke="#3b82f6"
                    fillOpacity={1}
                    fill="url(#colorSent)"
                  />
                  <Area
                    type="monotone"
                    dataKey="delivered"
                    stroke="#10b981"
                    fillOpacity={1}
                    fill="url(#colorDelivered)"
                  />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t('deliveryStatus')}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-80 flex items-center justify-center">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={statusData}
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={120}
                    paddingAngle={5}
                    dataKey="value"
                  >
                    {statusData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                </PieChart>
              </ResponsiveContainer>
            </div>
            <div className="grid grid-cols-2 gap-4 mt-4">
              {statusData.map((item, index) => (
                <div key={index} className="flex items-center space-x-2">
                  <div 
                    className="w-3 h-3 rounded-full" 
                    style={{ backgroundColor: item.color }}
                  />
                  <span className="text-sm text-muted-foreground">{item.name}</span>
                  <span className="text-sm font-medium">{item.value}%</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Recent Activity */}
      <Card>
        <CardHeader>
          <CardTitle>{t('recentActivity')}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {recentActivities.map((activity) => (
              <div key={activity.id} className="flex items-start space-x-3 p-3 rounded-lg hover:bg-accent/50 transition-colors">
                {getActivityIcon(activity.type)}
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-foreground">{activity.message}</p>
                  <p className="text-xs text-muted-foreground mt-1">{activity.time}</p>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default DashboardOverview;
