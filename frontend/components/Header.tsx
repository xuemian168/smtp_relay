
import React, { useState } from 'react';
import { Moon, Sun, Globe, Bell, Settings, User } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { useTranslations } from 'next-intl';
import { useTheme } from '@/components/layout/ThemeProvider';
import LanguageToggle from '@/components/layout/LanguageToggle';
import { useAuth } from '@/lib/api/auth';
import { useRedirect } from '@/lib/hooks/useRedirect';
// import { useToast } from '@/components/ui/use-toast';

const Header = () => {
  const t = useTranslations('header');
  const { theme, setTheme } = useTheme();
  const { logout } = useAuth();
  const { redirectToLogin } = useRedirect();
  // const { toast } = useToast ? useToast() : { toast: () => {} };

  const handleLogout = async () => {
    try {
      // TODO: 预留后端API调用
      // await api.logout();
    } catch (e) {
      // 可在此处理API错误
    }
    logout();
    redirectToLogin();
    // toast && toast({ description: t('logoutSuccess') || '已退出登录' });
  };

  return (
    <header className="h-16 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 sticky top-0 z-50">
      <div className="flex h-full items-center justify-between px-6">
        <div className="flex items-center space-x-3">
          <div className="w-8 h-8 bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg flex items-center justify-center">
            <span className="text-white font-bold text-sm">SR</span>
          </div>
          <div>
            <h1 className="font-semibold text-lg">SMTP Relay</h1>
            <p className="text-xs text-muted-foreground">Professional Email Delivery</p>
          </div>
        </div>

        <div className="flex items-center space-x-2">
          <LanguageToggle />

          <Button
            variant="ghost"
            size="icon"
            onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}
            className="h-8 w-8"
          >
            {theme === 'light' ? <Moon className="h-4 w-4" /> : <Sun className="h-4 w-4" />}
          </Button>

          <Button variant="ghost" size="icon" className="h-8 w-8 relative">
            <Bell className="h-4 w-4" />
            <Badge className="absolute -top-1 -right-1 h-3 w-3 p-0 bg-red-500 text-[10px]">3</Badge>
          </Button>

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <User className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-56">
              <DropdownMenuLabel>My Account</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem>
                <Settings className="mr-2 h-4 w-4" />
                <span>{t('settings')}</span>
              </DropdownMenuItem>
              <DropdownMenuItem onClick={handleLogout}>
                <span>{t('logout')}</span>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </header>
  );
};

export default Header;
