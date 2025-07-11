"use client";

import { useRouter, usePathname } from 'next/navigation';
import { useLocale } from 'next-intl';
import { localeNames } from '@/lib/i18n/config';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem
} from '@/components/ui/dropdown-menu';
import { Globe, ChevronDown, Check } from 'lucide-react';

export default function LanguageToggle() {
  const router = useRouter();
  const pathname = usePathname();
  const locale = useLocale();

  const switchLanguage = (newLocale: string) => {
    const segments = pathname.split('/');
    segments[1] = newLocale;
    const newPath = segments.join('/');
    router.push(newPath);
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="sm" className="flex items-center h-8 px-2">
          <Globe className="h-4 w-4 mr-1" />
          <span className="uppercase text-sm font-medium">{locale}</span>
          <ChevronDown className="h-4 w-4 ml-1" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {Object.entries(localeNames).map(([code, name]) => (
          <DropdownMenuItem
            key={code}
            onClick={() => switchLanguage(code)}
            className={
              code === locale
                ? 'font-semibold text-blue-600 dark:text-blue-400'
                : ''
            }
          >
            {code === locale && <Check className="h-4 w-4 mr-2 text-blue-600 dark:text-blue-400" />}
            <span className="mr-2 uppercase">{code}</span>
            <span>{name}</span>
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
} 