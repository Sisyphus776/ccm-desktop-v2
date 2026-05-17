import { t } from '@/i18n';
import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { rpcCall } from '@/lib/rpc-client';
import type { ClaudeMDItem } from '@/lib/types';
import { cn } from '@/lib/utils';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';

const SUGGESTED_PATHS = [
  { label: '用户级 (~/.claude/CLAUDE.md)', path: '' }, // filled at runtime
  { label: '项目级 (当前目录/CLAUDE.md)', path: 'CLAUDE.md' },
];

export default function ClaudeMd() {
  const { toast } = useToast();
  const qc = useQueryClient();

  const [selectedPath, setSelectedPath] = useState<string | null>(null);
  const [editing, setEditing] = useState(false);
  const [editContent, setEditContent] = useState('');
  const [showCreate, setShowCreate] = useState(false);
  const [newPath, setNewPath] = useState('');
  const [newContent, setNewContent] = useState('');
  const [showDelete, setShowDelete] = useState(false);

  const { data: mds = [], isLoading } = useQuery<ClaudeMDItem[]>({
    queryKey: ['claudemd'],
    queryFn: () => rpcCall('claudemd.list'),
  });

  const viewMut = useMutation({
    mutationFn: (path: string) => rpcCall('claudemd.get_content', { path }),
    onSuccess: (data: string) => setEditContent(data),
    onError: (err: Error) => toast({ title: err.message, variant: 'destructive' }),
  });

  const createMut = useMutation({
    mutationFn: () => rpcCall('claudemd.create', { path: newPath, content: newContent }),
    onSuccess: (result: string) => {
      toast({ title: result });
      setShowCreate(false);
      setNewPath('');
      setNewContent('');
      qc.invalidateQueries({ queryKey: ['claudemd'] });
    },
    onError: (err: Error) => toast({ title: err.message, variant: 'destructive' }),
  });

  const updateMut = useMutation({
    mutationFn: () => rpcCall('claudemd.update', { path: selectedPath, content: editContent }),
    onSuccess: (result: string) => {
      toast({ title: result });
      setEditing(false);
      qc.invalidateQueries({ queryKey: ['claudemd'] });
    },
    onError: (err: Error) => toast({ title: err.message, variant: 'destructive' }),
  });

  const deleteMut = useMutation({
    mutationFn: () => rpcCall('claudemd.delete', { path: selectedPath }),
    onSuccess: (result: string) => {
      toast({ title: result });
      setShowDelete(false);
      setSelectedPath(null);
      qc.invalidateQueries({ queryKey: ['claudemd'] });
    },
    onError: (err: Error) => toast({ title: err.message, variant: 'destructive' }),
  });

  function viewFile(path: string) {
    setSelectedPath(path);
    setEditing(false);
    viewMut.mutate(path);
  }

  function startEdit() {
    setEditing(true);
  }

  function cancelEdit() {
    setEditing(false);
    if (selectedPath) viewMut.mutate(selectedPath);
  }

  return (
    <div className="content">
      <div className="page-header">
        <h2>{t('claudemd.title')}</h2>
        <div className="flex gap-2">
          <Button size="sm" onClick={() => { setShowCreate(true); setNewPath(''); setNewContent(''); }}>
            + 新建
          </Button>
          <span className="refresh-hint self-center">{mds.length} files</span>
        </div>
      </div>

      {/* Create Dialog */}
      <Dialog open={showCreate} onOpenChange={setShowCreate}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>新建 CLAUDE.md</DialogTitle>
            <DialogDescription>创建新的 CLAUDE.md 配置文件</DialogDescription>
          </DialogHeader>
          <div className="flex flex-col gap-3 mt-2">
            <div>
              <label className="text-xs text-[var(--text-secondary)]">文件路径</label>
              <Input
                value={newPath}
                onChange={(e) => setNewPath(e.target.value)}
                placeholder="C:\Users\...\.claude\CLAUDE.md"
              />
              <div className="flex gap-1 mt-1 flex-wrap">
                {SUGGESTED_PATHS.map((s) => (
                  <button
                    key={s.label}
                    className="text-xs text-[var(--text-link)] hover:underline"
                    onClick={() => setNewPath(s.path)}
                  >
                    {s.label}
                  </button>
                ))}
              </div>
            </div>
            <div>
              <label className="text-xs text-[var(--text-secondary)]">内容</label>
              <Textarea
                value={newContent}
                onChange={(e) => setNewContent(e.target.value)}
                placeholder="# 项目说明&#10;..."
                className="min-h-[200px] font-mono text-xs"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreate(false)}>取消</Button>
            <Button onClick={() => createMut.mutate()} disabled={createMut.isPending || !newPath || !newContent}>
              {createMut.isPending ? '创建中...' : '创建'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Dialog */}
      <Dialog open={showDelete} onOpenChange={setShowDelete}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认删除</DialogTitle>
            <DialogDescription>
              此操作不可撤销。确定要删除以下文件吗？
              <br />
              <code className="text-xs text-[var(--danger)]">{selectedPath}</code>
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDelete(false)}>取消</Button>
            <Button variant="destructive" onClick={() => deleteMut.mutate()} disabled={deleteMut.isPending}>
              {deleteMut.isPending ? '删除中...' : '确认删除'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {isLoading ? (
        <div className="p-8 text-center text-[var(--text-secondary)]">{t('common.loading')}</div>
      ) : (
        <div className="master-detail">
          <div className="master-list">
            {mds.length === 0 ? (
              <div className="p-4 text-[var(--text-secondary)] text-sm">No CLAUDE.md files found</div>
            ) : (
              mds.map((md) => (
                <div
                  key={md.path}
                  className={cn('master-item', selectedPath === md.path && 'active')}
                  onClick={() => viewFile(md.path)}
                >
                  <div><code className="text-xs">{md.path}</code></div>
                  <div className="text-xs text-[var(--text-secondary)] mt-1">
                    {md.level} &middot; {md.size} bytes
                    {md.references?.length ? <span> &middot; {md.references.length} refs</span> : null}
                  </div>
                </div>
              ))
            )}
          </div>

          {selectedPath ? (
            <div className="detail-panel">
              <div className="flex items-center justify-between mb-3">
                <h3><code className="text-sm">{selectedPath}</code></h3>
                <div className="flex gap-2">
                  {editing ? (
                    <>
                      <Button size="sm" onClick={cancelEdit}>取消</Button>
                      <Button
                        size="sm"
                        onClick={() => updateMut.mutate()}
                        disabled={updateMut.isPending}
                      >
                        {updateMut.isPending ? '保存中...' : '保存'}
                      </Button>
                    </>
                  ) : (
                    <>
                      <Button size="sm" variant="outline" onClick={startEdit}>编辑</Button>
                      <Button size="sm" variant="destructive" onClick={() => setShowDelete(true)}>删除</Button>
                    </>
                  )}
                </div>
              </div>
              {editing ? (
                <Textarea
                  value={editContent}
                  onChange={(e) => setEditContent(e.target.value)}
                  className="min-h-[400px] font-mono text-xs"
                />
              ) : (
                <pre className="bg-[var(--bg-tertiary)] p-3 rounded-md text-xs max-h-[600px] overflow-y-auto whitespace-pre-wrap font-mono">
                  {editContent}
                </pre>
              )}
            </div>
          ) : (
            <div className="detail-panel empty-panel">Select a file to view content</div>
          )}
        </div>
      )}
    </div>
  );
}
