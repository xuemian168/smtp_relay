import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'SMTP Relay Service Dashboard',
  description: 'SMTP中继服务控制台',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <head>
        <script dangerouslySetInnerHTML={{
          __html: `
            (function() {
              try {
                var theme = localStorage.getItem('theme');
                if (theme === 'dark' || (!theme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
                  document.documentElement.classList.add('dark');
                } else {
                  document.documentElement.classList.remove('dark');
                }
              } catch (_) {}
            })();
          `
        }} />
      </head>
      <body>
        {children}
      </body>
    </html>
  );
} 