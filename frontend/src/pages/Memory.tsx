import { t } from '@/i18n';
import { useState, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { rpcCall } from '@/lib/rpc-client';
import type { MemoryFileItem, MemoryStats, IssueItem } from '@/lib/types';
import { cn } from '@/lib/utils';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

const typeLabels: Record<string, string> = {
  user: '用户',
  feedback: '反馈',
  project: '项目',
  reference: '参考',
};

function typeLabel(tp: string) {
  return typeLabels[tp] || tp;
}

export default function Memory() {
  const { toast } = useToast();
  const qc = useQueryClient();

  const [showCreate, setShowCreate] = useState(false);
  const [memName, setMemName] = useState('');
  const [memType, setMemType] = useState('feedback');
  const [memDesc, setMemDesc] = useState('');
  const [memContent, setMemContent] = useState('');
  const [search, setSearch] = useState('');
  const [viewing, setViewing] = useState<MemoryFileItem | null>(null);
  const [viewContent, setViewContent] = useState('');

  const { data: files, isLoading } = useQuery<MemoryFileItem[]>({
    queryKey: ['memory'],
    queryFn: () => rpcCall('memory.list'),
  });
  const safeFiles = files || [];

  const { data: stats } = useQuery<MemoryStats>({
    queryKey: ['memory', 'stats'],
    queryFn: () => rpcCall('memory.stats'),
  });

  const { data: issues } = useQuery<IssueItem[]>({
    queryKey: ['memory', 'validate'],
    queryFn: () => rpcCall('memory.validate'),
  });
  const safeIssues = issues || [];

  const filtered = useMemo(() => {
    if (!search) return safeFiles;
    const q = search.toLowerCase();
    return safeFiles.filter(
      (f) =>
        (f.name || '').toLowerCase().includes(q) ||
        (f.description || '').toLowerCase().includes(q) ||
        (f.file || '').toLowerCase().includes(q),
    );
  }, [safeFiles, search]);

  function invalidate() {
    qc.invalidateQueries({ queryKey: ['memory'] });
    qc.invalidateQueries({ queryKey: ['memory', 'stats'] });
    qc.invalidateQueries({ queryKey: ['memory', 'validate'] });
  }

  const createMut = useMutation({
    mutationFn: () => rpcCall('memory.create', { name: memName, type: memType, description: memDesc, content: memContent }),
    onSuccess: (result: string) => {
      toast({ title: result });
      setShowCreate(false);
      setMemName(''); setMemDesc(''); setMemContent('');
      invalidate();
    },
  });

  const viewMut = useMutation({
    mutationFn: (file: string) => rpcCall('memory.get_content', { file }),
    onSuccess: (content: string) => setViewContent(content),
    onError: (err: Error) => toast({ title: err.message, variant: 'destructive' }),
  });

  async function doView(f: MemoryFileItem) {
    if (viewing?.file !== f.file) setViewContent('');
    setViewing(f);
    viewMut.mutate(f.file);
  }

  const deleteMut = useMutation({
    mutationFn: (file: string) => {
      if (!confirm(`确定删除 "${file}"？`)) throw new Error('cancelled');
      return rpcCall('memory.delete', { file });
    },
    onSuccess: () => { toast({ title: '已删除' }); setViewing(null); invalidate(); },
    onError: (err) => { if (err.message !== 'cancelled') toast({ title: String(err), variant: 'destructive' }); },
  });

  return (
    <div className="content">
      <div className="page-header">
        <h2>{t('memory.title')}</h2>
        <div className="flex gap-2">
          <Button size="sm" onClick={() => setShowCreate(!showCreate)}>
            {showCreate ? t('common.cancel') : t('common.createMemory')}
          </Button>
          <span className="refresh-hint self-center">{stats?.total ?? 0} 条</span>
        </div>
      </div>

      {showCreate && (
        <div className="card mb-4">
          <h3 className="text-sm font-semibold mb-2">新建 Memory</h3>
          <div className="flex flex-col gap-2 mb-2">
            <Input value={memName} onChange={(e) => setMemName(e.target.value)} placeholder={t('common.name')} />
            <Select value={memType} onValueChange={setMemType}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="user">用户</SelectItem>
                <SelectItem value="feedback">反馈</SelectItem>
                <SelectItem value="project">项目</SelectItem>
                <SelectItem value="reference">参考</SelectItem>
              </SelectContent>
            </Select>
            <Input value={memDesc} onChange={(e) => setMemDesc(e.target.value)} placeholder="简短描述" />
            <Textarea value={memContent} onChange={(e) => setMemContent(e.target.value)} placeholder="正文内容" rows={4} />
          </div>
          <Button size="sm" onClick={() => createMut.mutate()} disabled={createMut.isPending}>
            {t('common.create')}
          </Button>
        </div>
      )}

      {viewing && (
        <div className="card mb-4">
          <div className="flex justify-between items-center">
            <h3 className="text-sm font-semibold">{viewing.name || viewing.file}</h3>
            <div className="flex gap-2">
              <Button size="sm" variant="destructive" onClick={() => deleteMut.mutate(viewing.file)}>{t('common.delete')}</Button>
              <Button size="sm" variant="outline" onClick={() => setViewing(null)}>{t('common.close')}</Button>
            </div>
          </div>
          <p className="text-xs text-[var(--text-secondary)] my-1">{typeLabel(viewing.type)} / {viewing.description}</p>
          <pre className="bg-[var(--bg-tertiary)] p-3 rounded-md text-xs max-h-[300px] overflow-y-auto whitespace-pre-wrap">{viewContent}</pre>
        </div>
      )}

      {isLoading ? (
        <div className="p-8 text-center text-[var(--text-secondary)]">{t('common.loading')}</div>
      ) : (
        <>
          {stats && (
            <div className="grid grid-cols-2 md:grid-cols-5 gap-3 mb-4">
              <div className="card text-center">
                <h3 className="text-xs font-semibold text-[var(--text-secondary)] uppercase mb-1">总计</h3>
                <div className="text-xl font-bold text-[var(--text-link)]">{stats.total}</div>
              </div>
              {Object.entries(stats.byType || {}).map(([type, count]) => (
                <div key={type} className="card text-center">
                  <h3 className="text-xs font-semibold text-[var(--text-secondary)] uppercase mb-1">{typeLabel(type)}</h3>
                  <div className="text-xl font-bold text-[var(--text-link)]">{count}</div>
                </div>
              ))}
            </div>
          )}

          <div className="mb-2">
            <input value={search} onChange={(e) => setSearch(e.target.value)} placeholder="搜索 Memory..." className="search-input" />
          </div>

          {filtered.length > 0 ? (
            <table>
              <thead>
                <tr>
                  <th>文件</th><th>{t('common.name')}</th><th>{t('common.type')}</th><th>描述</th><th style={{ width: 80 }}>{t('common.actions')}</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((f) => (
                  <tr key={f.file} className={cn(viewing?.file === f.file && 'bg-[var(--active-bg)]')}>
                    <td><code className="text-xs">{f.file}</code></td>
                    <td>{f.name || '-'}</td>
                    <td><span className="tag tag-info">{typeLabel(f.type) || '未知'}</span></td>
                    <td className="text-[var(--text-secondary)] text-xs">{f.description || '-'}</td>
                    <td>
                      <Button size="sm" variant="outline" className="text-xs h-6 px-2 mr-1" onClick={() => doView(f)}>{t('common.view')}</Button>
                      <Button size="sm" variant="destructive" className="text-xs h-6 px-2" onClick={() => deleteMut.mutate(f.file)}>{t('common.delete')}</Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            <div className="text-[var(--text-secondary)] text-sm">暂无 Memory 文件</div>
          )}
        </>
      )}

      {safeIssues.length > 0 && (
        <>
          <h3 className="text-sm font-semibold text-[var(--text-primary)] mt-6 mb-2">验证问题 ({safeIssues.length})</h3>
          {safeIssues.map((iss, i) => (
            <div key={i} className={cn(
              'p-3 rounded-md mb-2 text-sm',
              iss.severity === 'error' ? 'bg-[var(--danger-bg)] border border-[var(--danger)]' : 'bg-[var(--active-bg)] border border-[var(--text-link)]',
            )}>
              <span className="font-semibold text-xs">[{iss.severity?.toUpperCase()}]</span>{' '}
              {iss.domain}: {iss.message}
            </div>
          ))}
        </>
      )}
    </div>
  );
}
