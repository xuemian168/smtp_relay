"use client";

import { useEffect, useState } from 'react';
import { useTranslations } from 'next-intl';
import { credentialsApi } from '@/lib/api';
import { getSystem } from '@/lib/api/orval/system';
import {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Table,
  TableHeader,
  TableRow,
  TableHead,
  TableBody,
  TableCell
} from '@/components/ui/table';
import { Check, Copy, Info, Shield } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';
import DKIMManager from './DKIMManager';

function AddCredentialModal({ open, onClose, onSuccess }: { open: boolean; onClose: () => void; onSuccess: (result?: any) => void }) {
  const t = useTranslations('credentials');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const res = await credentialsApi.create({ name, description });
      onSuccess(res);
      onClose();
      setName('');
      setDescription('');
    } catch (e: any) {
      setError(e.message || 'Error');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={v => { if (!v) onClose(); }}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('add')}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <Label htmlFor="name">{t('name')}</Label>
            <Input id="name" value={name} onChange={e => setName(e.target.value)} required maxLength={50} />
          </div>
          <div>
            <Label htmlFor="desc">{t('description')}</Label>
            <Input id="desc" value={description} onChange={e => setDescription(e.target.value)} maxLength={200} />
          </div>
          {error && <div className="text-red-500 text-sm">{error}</div>}
          <DialogFooter>
            <Button type="button" variant="secondary" onClick={onClose}>{t('cancel')}</Button>
            <Button type="submit" disabled={loading}>
              {loading ? t('loading') : t('add')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function PasswordModal({ open, username, password, onClose }: { open: boolean; username: string; password: string; onClose: () => void }) {
  const t = useTranslations('credentials');
  const { toast } = useToast();
  const [copied, setCopied] = useState<'username' | 'password' | null>(null);

  const handleCopy = async (text: string, type: 'username' | 'password') => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(type);
      toast({ description: t('copied') });
      setTimeout(() => setCopied(null), 1200);
    } catch {
      toast({ description: t('copyFailed') || '复制失败', variant: 'destructive' });
    }
  };

  return (
    <Dialog open={open} onOpenChange={v => { if (!v) onClose(); }}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('newPassword')}</DialogTitle>
        </DialogHeader>
        <div className="mb-2">
          <div className="text-sm text-gray-500 mb-1">{t('username')}</div>
          <div
            className="mb-2 text-lg font-mono break-all text-green-700 dark:text-green-300 flex items-center gap-2 cursor-pointer hover:bg-accent/40 rounded px-2 py-1 select-all"
            title={t('copy')}
            onClick={() => handleCopy(username, 'username')}
          >
            {username}
            {copied === 'username' ? <Check className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4 opacity-60" />}
          </div>
          <div className="text-sm text-gray-500 mb-1">{t('password')}</div>
          <div
            className="mb-2 text-lg font-mono break-all text-blue-700 dark:text-blue-300 flex items-center gap-2 cursor-pointer hover:bg-accent/40 rounded px-2 py-1 select-all"
            title={t('copy')}
            onClick={() => handleCopy(password, 'password')}
          >
            {password}
            {copied === 'password' ? <Check className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4 opacity-60" />}
          </div>
        </div>
        <div className="mb-4 text-sm text-red-600">{t('passwordShowOnce')}</div>
        <DialogFooter>
          <Button onClick={onClose}>{t('confirm')}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function ConfirmDeleteModal({ open, onClose, onConfirm, name }: { open: boolean; onClose: () => void; onConfirm: () => void; name: string }) {
  const t = useTranslations('credentials');
  return (
    <Dialog open={open} onOpenChange={v => { if (!v) onClose(); }}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('confirmDelete')}</DialogTitle>
        </DialogHeader>
        <div className="mb-4">{t('deleteWarning')}</div>
        <div className="mb-4 font-semibold text-red-600">{name}</div>
        <DialogFooter>
          <Button variant="secondary" onClick={onClose}>{t('cancel')}</Button>
          <Button variant="destructive" onClick={onConfirm}>{t('delete')}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function ComplianceDialog({ open, onClose, relayDomain, relayIP, loading, error }: { open: boolean; onClose: () => void; relayDomain: string; relayIP: string; loading: boolean; error: string | null }) {
  const t = useTranslations('credentials');
  return (
    <Dialog open={open} onOpenChange={v => { if (!v) onClose(); }}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>
            <Info className="inline-block mr-2 text-blue-500" />{t('complianceGuide')} {t('complianceGuideDescription')}
          </DialogTitle>
        </DialogHeader>
        {loading ? (
          <div className="text-gray-500 py-4">{t('loading')}</div>
        ) : error ? (
          <div className="text-red-500 py-4">{error}</div>
        ) : !relayDomain || !relayIP ? (
          <div className="text-gray-500 py-4">{t('relayInfoUnavailable') || 'Relay info unavailable'}</div>
        ) : (
          <div className="space-y-3 text-sm">
            <div>
              <b>{t('spfRecord')}</b><br />
              {t('spfRecordDescription', { relayDomain, relayIP })}<br />
              <code className="block bg-gray-100 p-1 rounded my-1 select-all">v=spf1 ip4:{relayIP} -all</code>
              或<br />
              <code className="block bg-gray-100 p-1 rounded my-1 select-all">v=spf1 include:{relayDomain} -all</code>
            </div>
            <div>
              <b>{t('dkimSignature')}</b><br />
              {t('dkimSignatureDescription', { relayDomain })}<br />
              {t('dkimSignatureDescription2')}
            </div>
            <div>
              <b>{t('reverseDns')}</b><br />
              {t('reverseDnsDescription', { relayIP, relayDomain })}
            </div>
            <div>
              <b>{t('aRecord')}</b><br />
              {t('aRecordDescription', { relayDomain, relayIP })}
            </div>
            <div className="mt-2">
              <b>{t('detectionTools')}</b><br />
              <a href="https://mxtoolbox.com/spf.aspx" target="_blank" className="text-blue-600 underline">{t('spfDetection')}</a>、
              <a href="https://mxtoolbox.com/dkim.aspx" target="_blank" className="text-blue-600 underline">{t('dkimDetection')}</a>
            </div>
          </div>
        )}
        <DialogFooter>
          <Button onClick={onClose}>{t('iKnow')}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default function CredentialList() {
  const t = useTranslations('credentials');
  const [list, setList] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showAdd, setShowAdd] = useState(false);
  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [deleteName, setDeleteName] = useState<string>('');
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [newPassword, setNewPassword] = useState('');
  const [newUsername, setNewUsername] = useState('');
  const [showCompliance, setShowCompliance] = useState(false);
  const [showDKIM, setShowDKIM] = useState(false);
  const [relayDomain, setRelayDomain] = useState('');
  const [relayIP, setRelayIP] = useState('');
  const [relayInfoLoading, setRelayInfoLoading] = useState(true);
  const [relayInfoError, setRelayInfoError] = useState<string | null>(null);

  const fetchList = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await credentialsApi.list();
      setList(res.data || []);
    } catch (e: any) {
      setError(e.message || 'Error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    // 获取 relay info
    setRelayInfoLoading(true);
    getSystem().getApiRelayInfo()
      .then(res => {
        setRelayDomain(res.data?.relayDomain || '');
        setRelayIP(res.data?.relayIP || '');
        setRelayInfoError(null);
      })
      .catch(e => {
        setRelayInfoError(e.message || 'Failed to fetch relay info');
      })
      .finally(() => setRelayInfoLoading(false));
    fetchList();
  }, []);

  const handleDelete = async () => {
    if (!deleteId) return;
    setDeleteLoading(true);
    try {
      await credentialsApi.deleteCredential(deleteId);
      setDeleteId(null);
      setDeleteName('');
      fetchList();
    } catch (e: any) {
      setError(e.message || 'Error');
    } finally {
      setDeleteLoading(false);
    }
  };

  const handleAddSuccess = (res?: any) => {
    fetchList();
    if (res && res.data && res.data.password && res.data.credential && res.data.credential.username) {
      setNewPassword(res.data.password);
      setNewUsername(res.data.credential.username);
      setShowPassword(true);
      setShowCompliance(true); // 新建凭据成功后自动弹出合规性Dialog
    }
  };

  return (
    <div>
      <div className="flex justify-end mb-2 gap-2">
        <Button variant="outline" onClick={() => setShowCompliance(true)}>
          <Info className="inline-block mr-1 h-4 w-4 text-blue-500" />{t('complianceGuide')}
        </Button>
        <Button variant="outline" onClick={() => setShowDKIM(true)}>
          <Shield className="inline-block mr-1 h-4 w-4 text-green-500" />{t('dkimManagement')}
        </Button>
        <Button onClick={() => setShowAdd(true)}>{t('add')}</Button>
      </div>
      <ComplianceDialog open={showCompliance} onClose={() => setShowCompliance(false)} relayDomain={relayDomain} relayIP={relayIP} loading={relayInfoLoading} error={relayInfoError} />
      
      {/* DKIM管理对话框 */}
      <Dialog open={showDKIM} onOpenChange={setShowDKIM}>
        <DialogContent className="max-w-6xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Shield className="h-5 w-5 text-green-500" />
              DKIM密钥管理
            </DialogTitle>
          </DialogHeader>
          <DKIMManager />
        </DialogContent>
      </Dialog>
      
      <AddCredentialModal open={showAdd} onClose={() => setShowAdd(false)} onSuccess={handleAddSuccess} />
      <PasswordModal open={showPassword} username={newUsername} password={newPassword} onClose={() => setShowPassword(false)} />
      <ConfirmDeleteModal
        open={!!deleteId}
        onClose={() => { setDeleteId(null); setDeleteName(''); }}
        onConfirm={handleDelete}
        name={deleteName}
      />
      {loading && <div>{t('loading')}</div>}
      {error && <div className="text-red-500 text-sm">{error}</div>}
      <Table className="mt-2">
        <TableHeader>
          <TableRow>
            <TableHead>{t('name')}</TableHead>
            <TableHead>{t('description')}</TableHead>
            <TableHead>{t('status')}</TableHead>
            <TableHead>{t('actions')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {list.map((item) => (
            <TableRow key={item.id}>
              <TableCell>{item.name}</TableCell>
              <TableCell>{item.description}</TableCell>
              <TableCell>{item.status}</TableCell>
              <TableCell className="space-x-2">
                <Button
                  size="sm"
                  variant="destructive"
                  onClick={() => { setDeleteId(item.id); setDeleteName(item.name); }}
                  disabled={deleteLoading}
                >
                  {t('delete')}
                </Button>
                <Button size="sm" variant="secondary" onClick={() => alert('TODO: reset')}>{t('resetPassword')}</Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
} 