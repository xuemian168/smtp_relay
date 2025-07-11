"use client";
import { useEffect, useState } from "react";
import { getMailLog } from "@/lib/api/orval/mail-log";
import { Table, TableHeader, TableRow, TableHead, TableBody, TableCell } from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert } from "@/components/ui/alert";
import type { ModelsMailLog } from '@/lib/api/orval/smtpApi.schemas';

export default function MailLogList() {
  const [mailLogs, setMailLogs] = useState<ModelsMailLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchLogs = async () => {
      setLoading(true);
      setError(null);
      try {
        const { getApiV1Logs } = getMailLog();
        const res = await getApiV1Logs({ page: 1, page_size: 20 });
        setMailLogs(res.data?.mail_logs || []);
      } catch (e: any) {
        setError(e.message || "加载失败");
      } finally {
        setLoading(false);
      }
    };
    fetchLogs();
  }, []);

  if (loading) return <Skeleton className="h-32 w-full" />;
  if (error) return <Alert variant="destructive">{error}</Alert>;
  if (!mailLogs.length) return <div>暂无邮件日志</div>;

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>ID</TableHead>
          <TableHead>发件人</TableHead>
          <TableHead>收件人</TableHead>
          <TableHead>主题</TableHead>
          <TableHead>时间</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {mailLogs.map((log) => (
          <TableRow key={log.id}>
            <TableCell>{log.id || '-'}</TableCell>
            <TableCell>{log.from || '-'}</TableCell>
            <TableCell>{Array.isArray(log.to) ? log.to.join(', ') : (log.to || '-')}</TableCell>
            <TableCell>{log.subject || '-'}</TableCell>
            <TableCell>{log.created_at || '-'}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}