import { t } from '@/i18n';
import { useState, useMemo, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { rpcCall } from '@/lib/rpc-client';
import type { SkillItem, SkillDetailItem, IssueItem } from '@/lib/types';
import { cn } from '@/lib/utils';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';

export default function Skills() {
  const { toast } = useToast();
  const qc = useQueryClient();

  const [selectedName, setSelectedName] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [showImport, setShowImport] = useState(false);
  const [newSkillName, setNewSkillName] = useState('');
  const [newSkillDesc, setNewSkillDesc] = useState('');
  const [importURL, setImportURL] = useState('');
  const [importLog, setImportLog] = useState('');
  const [search, setSearch] = useState('');
  const [checked, setChecked] = useState<Set<string>>(new Set());
  const [ctxMenu, setCtxMenu] = useState<{ x: number; y: number; name: string } | null>(null);

  // Close context menu on any click outside
  useEffect(() => {
    function close() { setCtxMenu(null); }
    if (ctxMenu) document.addEventListener('click', close);
    return () => document.removeEventListener('click', close);
  }, [ctxMenu]);

  const { data: skills = [], isLoading } = useQuery<SkillItem[]>({
    queryKey: ['skills'],
    queryFn: () => rpcCall('skills.list'),
  });

  const { data: issues = [] } = useQuery<IssueItem[]>({
    queryKey: ['skills', 'validate'],
    queryFn: () => rpcCall('skills.validate'),
  });

  const { data: detail } = useQuery<SkillDetailItem | null>({
    queryKey: ['skills', 'detail', selectedName],
    queryFn: () => rpcCall('skills.get_errors', { name: selectedName }),
    enabled: !!selectedName,
  });

  const selected = useMemo(
    () => skills.find((s) => s.name === selectedName) || null,
    [skills, selectedName],
  );

  const filtered = useMemo(() => {
    if (!search) return skills;
    const q = search.toLowerCase();
    return skills.filter(
      (s) =>
        s.name.toLowerCase().includes(q) ||
        (s.description || '').toLowerCase().includes(q),
    );
  }, [skills, search]);

  const allChecked =
    filtered.length > 0 && filtered.every((s) => checked.has(s.name));

  function toggleAll() {
    if (allChecked) {
      setChecked(new Set());
    } else {
      setChecked(new Set(filtered.map((s) => s.name)));
    }
  }

  function toggleOne(name: string) {
    const n = new Set(checked);
    if (n.has(name)) n.delete(name);
    else n.add(name);
    setChecked(n);
  }

  const toggleMut = useMutation({
    mutationFn: (name: string) => rpcCall('skills.toggle', { name }),
    onSuccess: (result: string) => {
      toast({ title: result });
      qc.invalidateQueries({ queryKey: ['skills'] });
      qc.invalidateQueries({ queryKey: ['skills', 'validate'] });
    },
    onError: (err: Error) => toast({ title: err.message, variant: 'destructive' }),
  });

  const createMut = useMutation({
    mutationFn: () => rpcCall('skills.create', { name: newSkillName, description: newSkillDesc }),
    onSuccess: (result: string) => {
      toast({ title: result });
      setShowCreate(false);
      setNewSkillName('');
      setNewSkillDesc('');
      qc.invalidateQueries({ queryKey: ['skills'] });
    },
    onError: (err: Error) => toast({ title: err.message, variant: 'destructive' }),
  });

  const importMut = useMutation({
    mutationFn: (url: string) => {
      setImportLog('正在克隆...\n');
      return rpcCall('skills.import', { url });
    },
    onSuccess: (result: string) => {
      toast({ title: result });
      qc.invalidateQueries({ queryKey: ['skills'] });
    },
    onError: (err: Error) => toast({ title: err.message, variant: 'destructive' }),
  });

  const deleteMut = useMutation({
    mutationFn: async (names: string[]) => {
      for (const name of names) {
        await rpcCall('skills.delete', { name });
      }
    },
    onSuccess: () => {
      toast({ title: '已删除' });
      setChecked(new Set());
      setCtxMenu(null);
      qc.invalidateQueries({ queryKey: ['skills'] });
      qc.invalidateQueries({ queryKey: ['skills', 'validate'] });
    },
    onError: (err: Error) => toast({ title: err.message, variant: 'destructive' }),
  });

  function onContextMenu(e: React.MouseEvent, name: string) {
    e.preventDefault();
    e.stopPropagation();
    setCtxMenu({ x: e.clientX, y: e.clientY, name });
  }

  function ctxToggle(name: string, disabled: boolean | undefined) {
    toggleMut.mutate(name);
    setCtxMenu(null);
  }

  function ctxDelete(name: string) {
    deleteMut.mutate([name]);
  }

  const statusLabel = (status: string) =>
    status === 'ok' ? '正常' : status === 'broken' ? '损坏' : '警告';

  return (
    <div className="content">
      <div className="page-header">
        <h2>{t('skills.title')}</h2>
        <div className="flex gap-2">
          <Button
            size="sm"
            onClick={() => {
              setShowCreate(!showCreate);
              setShowImport(false);
            }}
          >
            {showCreate ? t('common.cancel') : '+ 新建'}
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={() => {
              setShowImport(!showImport);
              setShowCreate(false);
            }}
          >
            {showImport ? t('common.cancel') : t('common.import')}
          </Button>
          <span className="refresh-hint self-center">{skills.length} 个 Skill</span>
        </div>
      </div>

      {/* Import Form */}
      {showImport && (
        <div className="card mb-4">
          <h3 className="text-sm font-semibold mb-2">从 GitHub 导入</h3>
          <div className="flex gap-2 mb-2">
            <Input
              value={importURL}
              onChange={(e) => setImportURL(e.target.value)}
              placeholder="https://github.com/user/skill-repo"
              disabled={importMut.isPending}
            />
          </div>
          <Button
            size="sm"
            onClick={() => importMut.mutate(importURL)}
            disabled={importMut.isPending || !importURL}
          >
            {importMut.isPending ? '导入中...' : '导入'}
          </Button>
          {importLog && (
            <pre className="bg-[#0a0a0a] text-[#00ff41] p-3 rounded-md text-xs max-h-[160px] overflow-y-auto mt-2 font-mono">
              {importLog}
            </pre>
          )}
          {importMut.data && <div className="text-sm text-[var(--text-secondary)] mt-2">{importMut.data}</div>}
        </div>
      )}

      {/* Create Form */}
      {showCreate && (
        <div className="card mb-4">
          <h3 className="text-sm font-semibold mb-2">新建 Skill</h3>
          <div className="flex flex-col gap-2 mb-2">
            <Input
              value={newSkillName}
              onChange={(e) => setNewSkillName(e.target.value)}
              placeholder="Skill 名称（英文）"
            />
            <Input
              value={newSkillDesc}
              onChange={(e) => setNewSkillDesc(e.target.value)}
              placeholder="描述 / 触发关键词"
            />
          </div>
          <Button
            size="sm"
            onClick={() => createMut.mutate()}
            disabled={createMut.isPending || !newSkillName}
          >
            创建
          </Button>
        </div>
      )}

      {/* Search + Batch bar */}
      {!isLoading && (
        <div className="flex items-center gap-3 mb-2">
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t('common.searchSkills')}
            className="search-input flex-1"
          />
          {checked.size > 0 && (
            <div className="flex items-center gap-2">
              <span className="text-xs text-[var(--text-secondary)]">{checked.size} 个已选</span>
              <Button
                size="sm"
                variant="destructive"
                onClick={() => deleteMut.mutate(Array.from(checked))}
              >
                删除所选
              </Button>
            </div>
          )}
        </div>
      )}

      {isLoading ? (
        <div className="p-8 text-center text-[var(--text-secondary)]">{t('common.loading')}</div>
      ) : (
        <div className="master-detail">
          <div className="master-list">
            {filtered.length === 0 ? (
              <div className="p-4 text-[var(--text-secondary)] text-sm">
                {search ? '无匹配结果' : '无已安装 Skill'}
              </div>
            ) : (
              <>
                <div className="flex items-center px-3 py-2">
                  <Checkbox checked={allChecked} onCheckedChange={toggleAll} />
                </div>
                {filtered.map((s) => (
                  <div
                    key={s.name}
                    className={cn('master-item', selectedName === s.name && 'active')}
                    onClick={() => setSelectedName(s.name)}
                    onContextMenu={(e) => onContextMenu(e, s.name)}
                  >
                    <div className="master-item-row">
                      <Checkbox
                        checked={checked.has(s.name)}
                        onCheckedChange={() => toggleOne(s.name)}
                        onClick={(e) => e.stopPropagation()}
                      />
                      <span className={cn(
                        'w-2 h-2 rounded-full',
                        s.status === 'ok' ? 'bg-[var(--success)]' : s.status === 'broken' ? 'bg-[var(--danger)]' : 'bg-[#d2a8ff]',
                      )} />
                      <span className="master-item-name">{s.name}</span>
                    </div>
                    <div className="master-item-sub">{s.type}</div>
                  </div>
                ))}
              </>
            )}
          </div>

          {/* Detail Panel */}
          {selected ? (
            <div className="detail-panel">
              <h3>{selected.name}</h3>
              <div className="grid grid-cols-2 gap-3 mt-3">
                <div>
                  <label className="text-xs text-[var(--text-secondary)]">调用指令</label>
                  <div><code>{selected.invocation || '/' + selected.name}</code></div>
                </div>
                <div>
                  <label className="text-xs text-[var(--text-secondary)]">类型</label>
                  <div className="text-sm">{selected.type}</div>
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    size="sm"
                    variant={selected.disabled ? 'outline' : 'destructive'}
                    onClick={() => toggleMut.mutate(selected.name)}
                  >
                    {selected.disabled ? '启用' : '禁用'}
                  </Button>
                  <label className="text-xs text-[var(--text-secondary)]">状态</label>
                  <span className={cn(
                    'tag',
                    selected.status === 'ok' ? 'tag-ok' : selected.status === 'broken' ? 'tag-err' : 'tag-warn',
                  )}>
                    {statusLabel(selected.status)}
                  </span>
                </div>
              </div>
              {selected.description && (
                <div className="mt-3">
                  <label className="text-xs text-[var(--text-secondary)]">描述</label>
                  <p className="text-sm mt-1">{selected.description}</p>
                  {selected.descriptionCN && (
                    <p className="text-xs text-[var(--text-secondary)] mt-1">{selected.descriptionCN}</p>
                  )}
                  {!selected.descriptionCN && (
                    <p className="text-xs text-[var(--text-secondary)] italic mt-1">（原文为中文）</p>
                  )}
                </div>
              )}
              {(selected.triggers && selected.triggers.length > 0) && (
                <div className="mt-3">
                  <label className="text-xs text-[var(--text-secondary)]">触发关键词</label>
                  <div className="flex flex-wrap gap-1 mt-1">
                    {selected.triggers.map((t) => (
                      <span key={t} className="bg-[var(--bg-tertiary)] px-2 py-0.5 rounded text-xs">{t}</span>
                    ))}
                  </div>
                </div>
              )}
              {selected.target && (
                <div className="mt-3">
                  <label className="text-xs text-[var(--text-secondary)]">目标路径</label>
                  <div><code className="text-xs text-[var(--text-secondary)]">{selected.target}</code></div>
                </div>
              )}
              {detail?.errors?.length ? (
                <div className="mt-3 p-3 bg-[var(--danger-bg)] border border-[var(--danger)] rounded-md">
                  <h4 className="text-sm font-semibold text-[var(--danger)] mb-1">{detail.errors.length} 个错误</h4>
                  <ul className="list-disc list-inside text-xs space-y-1">
                    {detail.errors.map((e, i) => <li key={i}>{e}</li>)}
                  </ul>
                </div>
              ) : null}
            </div>
          ) : (
            <div className="detail-panel empty-panel">选择一个 Skill 查看详情</div>
          )}
        </div>
      )}

      {issues.length > 0 && (
        <>
          <h3 className="text-sm font-semibold text-[var(--text-primary)] mt-6 mb-2">验证问题 ({issues.length})</h3>
          {issues.map((iss, i) => (
            <div key={i} className={cn(
              'p-3 rounded-md mb-2 text-sm',
              iss.severity === 'error' ? 'bg-[var(--danger-bg)] border border-[var(--danger)]' : 'bg-[var(--active-bg)] border border-[var(--text-link)]',
            )}>
              <span className="font-semibold text-xs">[{iss.severity?.toUpperCase()}]</span>{' '}
              {iss.domain}: {iss.message}
              {iss.fix && <div className="text-xs text-[var(--text-secondary)] mt-1">{iss.fix}</div>}
            </div>
          ))}
        </>
      )}

      {/* Context Menu */}
      {ctxMenu && (() => {
        const s = skills.find((sk) => sk.name === ctxMenu.name);
        return (
          <div
            className="context-menu"
            style={{ left: ctxMenu.x, top: ctxMenu.y }}
            onClick={(e) => e.stopPropagation()}
          >
            <button onClick={() => ctxToggle(ctxMenu.name, s?.disabled)}>
              {s?.disabled ? '启用' : '禁用'}
            </button>
            <button className="danger" onClick={() => ctxDelete(ctxMenu.name)}>
              删除
            </button>
          </div>
        );
      })()}
    </div>
  );
}
