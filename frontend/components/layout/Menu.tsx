
"use client";

import React from 'react';
import Link from 'next/link';
import { 
  BarChart3, 
  Settings, 
  Key, 
  FileText, 
  Home,
  Zap,
  Shield,
  Mail
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useTranslations } from 'next-intl';
import { usePathname } from 'next/navigation';

const Sidebar = () => {
  const t = useTranslations('sidebar');
  const pathname = usePathname();

  const navigation = [
    { name: t('dashboard'), href: '/', icon: Home },
    { name: t('analytics'), href: '/analytics', icon: BarChart3 },
    { name: t('apiKeys'), href: '/apiKeys', icon: Key },
    { name: 'SMTP', href: '/credentials', icon: Mail },
    { name: t('logs'), href: '/maillogs', icon: FileText },
    { name: t('settings'), href: '/settings', icon: Settings },
  ];

  return (
    <aside className="w-64 border-r bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="flex flex-col h-full">
        <nav className="flex-1 p-4 space-y-2">
          {navigation.map((item) => (
            <Link
              key={item.name}
              href={item.href}
              className={cn(
                'flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-all duration-200',
                'hover:bg-accent hover:text-accent-foreground',
                pathname === item.href
                  ? 'bg-gradient-to-r from-blue-600 to-purple-600 text-white shadow-lg'
                  : 'text-foreground/70'
              )}
            >
              <item.icon className="mr-3 h-4 w-4 flex-shrink-0" />
              {item.name}
            </Link>
          ))}
        </nav>

        <div className="p-4 border-t">
          <div className="bg-gradient-to-r from-blue-50 to-purple-50 dark:from-blue-950/30 dark:to-purple-950/30 rounded-lg p-4">
            <div className="flex items-center space-x-2 mb-2">
              <Zap className="h-4 w-4 text-blue-600" />
              <span className="text-sm font-semibold">Pro Plan</span>
            </div>
            <p className="text-xs text-muted-foreground mb-3">
              100K emails/month included
            </p>
            <div className="w-full bg-background rounded-full h-2 mb-2">
              <div className="bg-gradient-to-r from-blue-600 to-purple-600 h-2 rounded-full w-3/4"></div>
            </div>
            <p className="text-xs text-muted-foreground">75K of 100K used</p>
          </div>
        </div>
      </div>
    </aside>
  );
};

export default Sidebar;
