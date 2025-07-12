"use client";

import { useEffect, useState } from 'react';
import { useTranslations } from 'next-intl';
import { useLocale } from 'next-intl';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import {
  Table,
  TableHeader,
  TableRow,
  TableHead,
  TableBody,
  TableCell
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Check, Copy, Key, Shield, AlertCircle, CheckCircle, RotateCw, Trash2, Plus } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';
import { getDkim } from '@/lib/api/orval/dkim';
import type {
  ModelsDKIMKeyPair,
  ModelsDNSRecord,
  ModelsDKIMValidationResult
} from '@/lib/api/orval/smtpApi.schemas';

// DKIM密钥对类型定义
interface DKIMKeyPair {
  id: string;
  domain: string;
  selector: string;
  key_size: number;
  algorithm: string;
  status: string;
  dns_record: string;
  dns_verified: boolean;
  last_verified?: string;
  created_at: string;
  updated_at: string;
  expires_at?: string;
  public_key?: string; // Added for CreateSuccessDialog
}

// DNS记录类型定义
interface DNSRecord {
  type: string;
  name: string;
  value: string;
  ttl: number;
  priority?: number;
}

// DKIM验证结果类型定义
interface DKIMValidationResult {
  domain: string;
  selector: string;
  valid: boolean;
  dns_found: boolean;
  dns_record?: string;
  expected_dns: string;
  error_message?: string;
  checked_at: string;
}

// 用 orval 生成的 dkim API 替换原有 dkimApi 对象
const dkimApi = getDkim();

// 类型映射函数
function toDKIMKeyPair(model: ModelsDKIMKeyPair): DKIMKeyPair {
  return {
    id: model.id || '',
    domain: model.domain || '',
    selector: model.selector || '',
    key_size: model.key_size || 2048,
    algorithm: model.algorithm || '',
    status: model.status || '',
    dns_record: model.dns_record || '',
    dns_verified: !!model.dns_verified,
    last_verified: model.last_verified,
    created_at: model.created_at || '',
    updated_at: model.updated_at || '',
    expires_at: model.expires_at,
    public_key: model.public_key, // Added for CreateSuccessDialog
  };
}

function toDNSRecord(model: ModelsDNSRecord): DNSRecord {
  return {
    type: model.type || '',
    name: model.name || '',
    value: model.value || '',
    ttl: model.ttl || 0,
    priority: model.priority,
  };
}

function toDKIMValidationResult(model: ModelsDKIMValidationResult): DKIMValidationResult {
  return {
    domain: model.domain || '',
    selector: model.selector || '',
    valid: !!model.valid,
    dns_found: !!model.dns_found,
    dns_record: model.dns_record,
    expected_dns: model.expected_dns || '',
    error_message: model.error_message,
    checked_at: model.checked_at || '',
  };
}

// 适配 orval API 返回结构
async function listKeys(): Promise<DKIMKeyPair[]> {
  const res = await dkimApi.getApiV1DkimKeys();
  const data: any = res.data;
  if (Array.isArray(data)) {
    return (data as any[]).map(toDKIMKeyPair);
  } else if (data && Array.isArray((data as any).data)) {
    return (data as any).data.map(toDKIMKeyPair);
  } else {
    throw new Error('API返回格式不正确');
  }
}

// createKey 只抛出英文或默认错误
async function createKey(domain: string, selector: string, keySize: number): Promise<DKIMKeyPair> {
  const res = await dkimApi.postApiV1DkimKeys({ domain, selector, key_size: keySize });
  const data = res.data as any;
  // 兼容直接返回对象
  if (data && data.id && data.domain) {
    return toDKIMKeyPair(data);
  }
  if (!data || data.success === false) throw new Error(data.error || 'Failed to create DKIM key pair');
  if (!data.data) throw new Error('Failed to create DKIM key pair');
  return toDKIMKeyPair(data.data);
}

async function deleteKey(keyId: string) {
  const res = await dkimApi.deleteApiV1DkimKeysId(keyId);
  const data = res.data as any;
  if (!data.success) throw new Error(data.error || '删除DKIM密钥失败');
}

async function rotateKey(keyId: string): Promise<DKIMKeyPair> {
  const res = await dkimApi.postApiV1DkimKeysIdRotate(keyId);
  const data = res.data as any;
  if (!data.success) throw new Error(data.error || '轮换DKIM密钥失败');
  return toDKIMKeyPair(data.data!);
}

async function getDNSRecord(keyId: string): Promise<DNSRecord> {
  const res = await dkimApi.getApiV1DkimKeysIdDns(keyId);
  const data = res.data as any;
  // 兼容 success 字段不存在的情况
  if (data.success === false) throw new Error(data.error || '获取DNS记录失败');
  // 兼容 data.data 或 data
  const record = data.data || data;
  return toDNSRecord(record);
}

async function verifyDNS(keyId: string): Promise<DKIMValidationResult> {
  const res = await dkimApi.postApiV1DkimKeysIdVerify(keyId);
  const data = res.data as any;
  if (!data.success) throw new Error(data.error || '验证DNS记录失败');
  return toDKIMValidationResult(data.data!);
}

// 创建DKIM密钥对对话框
function CreateDKIMKeyDialog({ open, onClose, onSuccess }: { 
  open: boolean; 
  onClose: () => void; 
  onSuccess: (keyPair: DKIMKeyPair) => void 
}) {
  const t = useTranslations('dkim');
  const tCommon = useTranslations('common');
  const [domain, setDomain] = useState('');
  const [selector, setSelector] = useState('default');
  const [keySize, setKeySize] = useState(2048);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const domainRegex = /^(?!-)[A-Za-z0-9-]{1,63}(?<!-)(\.[A-Za-z0-9-]{1,63})+$/;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    if (!domainRegex.test(domain)) {
      setError(t('invalidDomain'));
      setLoading(false);
      return;
    }

    try {
      const keyPair = await createKey(domain, selector, keySize);
      onSuccess(keyPair);
      onClose();
      setDomain('');
      setSelector('default');
      setKeySize(2048);
    } catch (e: any) {
      console.error('createKey error:', e);
      setError(e.message || t('createFailed'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={v => { if (!v) onClose(); }}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Key className="h-5 w-5" />
            {t('add')}
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <Label htmlFor="domain">{t('domain') || '域名'}</Label>
            <Input 
              id="domain" 
              value={domain} 
              onChange={e => {
                setDomain(e.target.value);
                if (!domainRegex.test(e.target.value)) {
                  setError(t('invalidDomain'));
                } else {
                  setError(null);
                }
              }}
              placeholder="example.com"
              required 
            />
          </div>
          <div>
            <Label htmlFor="selector">{t('selector') || '选择器'}</Label>
            <Input 
              id="selector" 
              value={selector} 
              onChange={e => setSelector(e.target.value)} 
              placeholder="default"
              required 
            />
          </div>
          <div>
            <Label htmlFor="keySize">{t('keySize') || '密钥长度'}</Label>
            <Select value={keySize.toString()} onValueChange={v => setKeySize(parseInt(v))}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="1024">1024</SelectItem>
                <SelectItem value="2048">2048</SelectItem>
              </SelectContent>
            </Select>
          </div>
          {error && <div className="text-red-500 text-sm">{error}</div>}
          <DialogFooter>
            <Button type="button" variant="secondary" onClick={onClose}>{tCommon('cancel')}</Button>
            <Button type="submit" disabled={loading}>
              {loading ? tCommon('loading') : tCommon('create')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

// DNS记录显示对话框
function DNSRecordDialog({ open, onClose, keyPair }: { 
  open: boolean; 
  onClose: () => void; 
  keyPair: DKIMKeyPair | null 
}) {
  const { toast } = useToast();
  const [dnsRecord, setDnsRecord] = useState<DNSRecord | null>(null);
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState<'name' | 'value' | null>(null);

  useEffect(() => {
    if (open && keyPair) {
      loadDNSRecord();
    }
  }, [open, keyPair]);

  const loadDNSRecord = async () => {
    if (!keyPair) return;
    setLoading(true);
    try {
      const record = await getDNSRecord(keyPair.id);
      setDnsRecord(record);
    } catch (e: any) {
      toast({ description: e.message || '获取DNS记录失败', variant: 'destructive' });
    } finally {
      setLoading(false);
    }
  };

  const handleCopy = async (text: string, type: 'name' | 'value') => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(type);
      toast({ description: '已复制到剪贴板' });
      setTimeout(() => setCopied(null), 1200);
    } catch {
      toast({ description: '复制失败', variant: 'destructive' });
    }
  };

  return (
    <Dialog open={open} onOpenChange={v => { if (!v) onClose(); }}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>DNS记录配置</DialogTitle>
        </DialogHeader>
        {loading ? (
          <div className="py-8 text-center">加载中...</div>
        ) : dnsRecord ? (
          <div className="space-y-4">
            <div className="bg-blue-50 dark:bg-blue-950 p-4 rounded-lg">
              <h3 className="font-medium mb-2">配置说明</h3>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                请在您的DNS服务商处添加以下TXT记录以启用DKIM签名：
              </p>
            </div>
            <div className="space-y-3">
              <div>
                <Label className="text-sm font-medium">记录类型</Label>
                <div className="mt-1">
                  <Badge variant="secondary">{dnsRecord.type}</Badge>
                </div>
              </div>
              <div>
                <Label className="text-sm font-medium">记录名称</Label>
                <div 
                  className="mt-1 p-2 bg-gray-100 dark:bg-gray-800 rounded border font-mono text-sm break-all cursor-pointer hover:bg-gray-200 dark:hover:bg-gray-700 flex items-center gap-2"
                  onClick={() => handleCopy(dnsRecord.name, 'name')}
                >
                  <span className="flex-1">{dnsRecord.name}</span>
                  {copied === 'name' ? <Check className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4 opacity-60" />}
                </div>
              </div>
              <div>
                <Label className="text-sm font-medium">记录值</Label>
                <div 
                  className="mt-1 p-2 bg-gray-100 dark:bg-gray-800 rounded border font-mono text-sm break-all cursor-pointer hover:bg-gray-200 dark:hover:bg-gray-700 flex items-center gap-2 max-h-32 overflow-y-auto"
                  onClick={() => handleCopy(dnsRecord.value, 'value')}
                >
                  <span className="flex-1">{dnsRecord.value}</span>
                  {copied === 'value' ? <Check className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4 opacity-60" />}
                </div>
              </div>
              <div>
                <Label className="text-sm font-medium">TTL</Label>
                <div className="mt-1">
                  <Badge variant="outline">{dnsRecord.ttl}秒</Badge>
                </div>
              </div>
              {dnsRecord.priority !== undefined && (
                <div>
                  <Label className="text-sm font-medium">Priority</Label>
                  <div className="mt-1">
                    <Badge variant="outline">{dnsRecord.priority}</Badge>
                  </div>
                </div>
              )}
            </div>
            <div className="bg-yellow-50 dark:bg-yellow-950 p-4 rounded-lg">
              <p className="text-sm text-yellow-800 dark:text-yellow-200">
                <strong>注意：</strong>DNS记录配置后可能需要几分钟到几小时才能生效，请耐心等待。
              </p>
            </div>
          </div>
        ) : (
          <div className="py-8 text-center text-gray-500">无法加载DNS记录</div>
        )}
        <DialogFooter>
          <Button onClick={onClose}>关闭</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// 删除确认对话框
function ConfirmDeleteDialog({ open, onClose, onConfirm, keyPair }: { 
  open: boolean; 
  onClose: () => void; 
  onConfirm: () => void; 
  keyPair: DKIMKeyPair | null 
}) {
  const t = useTranslations('dkim');
  const tCommon = useTranslations('common');
  return (
    <Dialog open={open} onOpenChange={v => { if (!v) onClose(); }}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('confirmDelete')}</DialogTitle>
        </DialogHeader>
        <div className="space-y-2">
          <p>{t('confirmDeleteMessage')}</p>
          {keyPair && (
            <div className="bg-gray-100 dark:bg-gray-800 p-3 rounded">
              <div className="text-sm">
                <div><strong>{t('domain')}:</strong>{keyPair.domain}</div>
                <div><strong>{t('selector')}:</strong>{keyPair.selector}</div>
              </div>
            </div>
          )}
          <p className="text-red-600 text-sm">{t('deleteIrreversible')}</p>
        </div>
        <DialogFooter>
          <Button variant="secondary" onClick={onClose}>{tCommon('cancel')}</Button>
          <Button variant="destructive" onClick={onConfirm}>{tCommon('delete')}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// 创建成功后展示 DKIM 信息的 Dialog
function CreateSuccessDialog({ open, onClose, keyPair }: {
  open: boolean;
  onClose: () => void;
  keyPair: DKIMKeyPair | null;
}) {
  const t = useTranslations('dkim');
  const tCommon = useTranslations('common');
  const { toast } = useToast();
  const [copied, setCopied] = useState(false);

  if (!keyPair) return null;

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(keyPair.public_key || '');
      setCopied(true);
      toast({ description: tCommon('copied') });
      setTimeout(() => setCopied(false), 1200);
    } catch {
      toast({ description: tCommon('copyFailed'), variant: 'destructive' });
    }
  };

  return (
    <Dialog open={open} onOpenChange={v => { if (!v) onClose(); }}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{t('createSuccess')}</DialogTitle>
        </DialogHeader>
        <div className="space-y-2 text-sm">
          <div><b>{t('domain')}:</b> {keyPair.domain}</div>
          <div><b>{t('selector')}:</b> {keyPair.selector}</div>
          <div><b>{t('algorithm')}:</b> {keyPair.algorithm}</div>
          <div><b>{t('keySize')}:</b> {keyPair.key_size} {t('bits')}</div>
          <div><b>{t('publicKey') || 'Public Key'}:</b></div>
          <div className="flex items-start gap-2">
            <pre style={{ whiteSpace: 'pre-wrap', fontSize: 12, background: '#f5f5f5', padding: 8, borderRadius: 4, flex: 1 }}>{keyPair.public_key}</pre>
            <Button size="icon" variant="outline" onClick={handleCopy} title={tCommon('copy')}>
              {copied ? <Check className="w-4 h-4 text-green-500" /> : <Copy className="w-4 h-4" />}
            </Button>
          </div>
        </div>
        <DialogFooter>
          <Button onClick={onClose}>{tCommon('close')}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default function DKIMManager() {
  const t = useTranslations('dkim');
  const tCommon = useTranslations('common');
  const locale = useLocale();
  console.log('Current locale:', locale);
  const [keyPairs, setKeyPairs] = useState<DKIMKeyPair[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [showDNS, setShowDNS] = useState(false);
  const [showDelete, setShowDelete] = useState(false);
  const [selectedKeyPair, setSelectedKeyPair] = useState<DKIMKeyPair | null>(null);
  const [verifyingKeys, setVerifyingKeys] = useState<Set<string>>(new Set());
  const [showCreateSuccessDialog, setShowCreateSuccessDialog] = useState(false);
  const [createdKeyPair, setCreatedKeyPair] = useState<DKIMKeyPair | null>(null);

  const fetchKeyPairs = async () => {
    setLoading(true);
    setError(null);
    try {
      const keys = await listKeys();
      setKeyPairs(keys);
    } catch (e: any) {
      setError(e.message || '获取DKIM密钥列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchKeyPairs();
  }, []);

  const handleCreateSuccess = (keyPair: DKIMKeyPair) => {
    setKeyPairs(prev => [...prev, keyPair]);
    setCreatedKeyPair(keyPair);
    setShowCreateSuccessDialog(true);
    toast({ description: t('createSuccess') });
  };

  const handleDelete = async () => {
    if (!selectedKeyPair) return;
    try {
      const res = await dkimApi.deleteApiV1DkimKeysId(selectedKeyPair.id);
      const data = res.data as any;
      if (!data || data.success === false) {
        throw new Error((data && data.error) || t('deleteFailed'));
      }
      setKeyPairs(prev => prev.filter(k => k.id !== selectedKeyPair.id));
      toast({ description: data.message || t('deleteSuccess') });
      setShowDelete(false);
      setSelectedKeyPair(null);
    } catch (e: any) {
      toast({ description: e.message || t('deleteFailed'), variant: 'destructive' });
    }
  };

  const handleRotate = async (keyPair: DKIMKeyPair) => {
    try {
      await rotateKey(keyPair.id);
      await fetchKeyPairs();
      toast({ description: t('rotateSuccess') });
    } catch (e: any) {
      toast({ description: e.message || '密钥轮换失败', variant: 'destructive' });
    }
  };

  const handleVerifyDNS = async (keyPair: DKIMKeyPair) => {
    setVerifyingKeys(prev => new Set(prev).add(keyPair.id));
    try {
      const result = await verifyDNS(keyPair.id);
      if (result.valid) {
        toast({ description: t('verifySuccess') });
        await fetchKeyPairs();
      } else {
        toast({ 
          description: t('verifyFailed', { msg: result.error_message || t('verifyFailedDefault') }), 
          variant: 'destructive' 
        });
      }
    } catch (e: any) {
      toast({ description: e.message || t('verifyDNSFailed'), variant: 'destructive' });
    } finally {
      setVerifyingKeys(prev => {
        const newSet = new Set(prev);
        newSet.delete(keyPair.id);
        return newSet;
      });
    }
  };

  const getStatusBadge = (keyPair: DKIMKeyPair) => {
    if (keyPair.status === 'active') {
      if (keyPair.dns_verified) {
        return <Badge className="bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">{t('verified')}</Badge>;
      } else {
        return <Badge className="bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200">{t('pendingVerification')}</Badge>;
      }
    } else if (keyPair.status === 'expiring') {
      return <Badge className="bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200">{t('expiring')}</Badge>;
    } else {
      return <Badge variant="secondary">{keyPair.status}</Badge>;
    }
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            {t('dkimManagement')}
          </CardTitle>
          <CardDescription>
            {t('dkimManagementDescription')}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex justify-end mb-4">
            <Button onClick={() => setShowCreate(true)}>
              <Plus className="h-4 w-4 mr-2" />
              {t('addDkimKey')}
            </Button>
          </div>

          {loading && <div className="text-center py-8">加载中...</div>}
          {error && <div className="text-red-500 text-sm mb-4">{error}</div>}

          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('domain')}</TableHead>
                <TableHead>{t('selector')}</TableHead>
                <TableHead>{t('algorithm')}</TableHead>
                <TableHead>{t('status')}</TableHead>
                <TableHead>{t('createdAt')}</TableHead>
                <TableHead>{t('actions')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {keyPairs.map((keyPair) => (
                <TableRow key={keyPair.id}>
                  <TableCell className="font-medium">{keyPair.domain}</TableCell>
                  <TableCell>{keyPair.selector}</TableCell>
                  <TableCell>
                    <Badge variant="outline">
                      {keyPair.algorithm} ({keyPair.key_size} bits)
                    </Badge>
                  </TableCell>
                  <TableCell>{getStatusBadge(keyPair)}</TableCell>
                  <TableCell>
                    {new Date(keyPair.created_at).toLocaleDateString()}
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => {
                          setSelectedKeyPair(keyPair);
                          setShowDNS(true);
                        }}
                      >
                        {t('dnsRecord')}
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleVerifyDNS(keyPair)}
                        disabled={verifyingKeys.has(keyPair.id)}
                      >
                        {verifyingKeys.has(keyPair.id) ? (
                          <>
                            <RotateCw className="h-3 w-3 mr-1 animate-spin" />
                            {t('verifying')}
                          </>
                        ) : keyPair.dns_verified ? (
                          <>
                            <CheckCircle className="h-3 w-3 mr-1" />
                            {t('reverify')}
                          </>
                        ) : (
                          <>
                            <AlertCircle className="h-3 w-3 mr-1" />
                            {t('verifyDns')}
                          </>
                        )}
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleRotate(keyPair)}
                      >
                        <RotateCw className="h-3 w-3 mr-1" />
                        {t('rotate')}
                      </Button>
                      <Button
                        size="sm"
                        variant="destructive"
                        onClick={() => {
                          setSelectedKeyPair(keyPair);
                          setShowDelete(true);
                        }}
                      >
                        <Trash2 className="h-3 w-3" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>

          {keyPairs.length === 0 && !loading && (
            <div className="text-center py-8 text-gray-500">
              {t('emptyList')}
            </div>
          )}
        </CardContent>
      </Card>

      <CreateDKIMKeyDialog 
        open={showCreate} 
        onClose={() => setShowCreate(false)} 
        onSuccess={handleCreateSuccess} 
      />
      
      <DNSRecordDialog 
        open={showDNS} 
        onClose={() => setShowDNS(false)} 
        keyPair={selectedKeyPair} 
      />
      
      <ConfirmDeleteDialog 
        open={showDelete} 
        onClose={() => setShowDelete(false)} 
        onConfirm={handleDelete} 
        keyPair={selectedKeyPair} 
      />
      <CreateSuccessDialog
        open={showCreateSuccessDialog}
        onClose={() => setShowCreateSuccessDialog(false)}
        keyPair={createdKeyPair}
      />
    </div>
  );
} 