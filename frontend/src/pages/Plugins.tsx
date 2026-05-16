import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { rpcCall } from '@/lib/rpc-client';
import type { PluginItem } from '@/lib/types';
import { cn } from '@/lib/utils';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';

export default function Plugins() {
  const { toast } = useToast();
  const qc = useQueryClient();
  const [selectedName, setSelectedName] = useState<string | null>(null);

  const { data: plugins = [], isLoading } = useQuery<PluginItem[]>({
    queryKey: ['plugins'],
    queryFn: () => rpcCall('plugins.list'),
  });

  const selected = plugins.find((p) => p.name === selectedName) || null;
  const totalSkills = plugins.reduce((a, p) => a + (p.skills?.length || 0), 0);

  function invalidate() {
    qc.invalidateQueries({ queryKey: ['plugins'] });
  }

  const disableAllMut = useMutation({
    mutationFn: () => rpcCall('plugins.disable_all'),
    onSuccess: () => { toast({ title: 'All plugins disabled' }); invalidate(); },
  });

  const enableAllMut = useMutation({
    mutationFn: () => rpcCall('plugins.enable_all'),
    onSuccess: () => { toast({ title: 'All plugins enabled' }); invalidate(); },
  });

  const togglePluginMut = useMutation({
    mutationFn: (p: PluginItem) => rpcCall('plugins.toggle_plugin', { name: p.name, installPath: p.installPath || '' }),
    onSuccess: () => { invalidate(); },
  });

  const toggleSkillMut = useMutation({
    mutationFn: (args: { pluginName: string; skillName: string; installPath: string }) =>
      rpcCall('plugins.toggle_skill', args),
    onSuccess: () => { invalidate(); },
  });

  return (
    <div className="content">
      <div className="page-header">
        <h2>Plugin Management</h2>
        <Button size="sm" variant="destructive" onClick={() => disableAllMut.mutate()}>Disable All</Button>
        <Button size="sm" onClick={() => enableAllMut.mutate()}>Enable All</Button>
        <span className="refresh-hint">{plugins.length} plugins / {totalSkills} skills</span>
      </div>

      {isLoading ? (
        <div className="p-8 text-center text-[var(--text-secondary)]">Loading...</div>
      ) : plugins.length === 0 ? (
        <div className="text-[var(--text-secondary)] text-sm">No plugins installed</div>
      ) : (
        <div className="master-detail">
          <div className="master-list">
            {plugins.map((p) => (
              <div
                key={p.name}
                className={cn('master-item', selectedName === p.name && 'active')}
                onClick={() => setSelectedName(p.name)}
              >
                <div className="master-item-row">
                  <span className="w-[22px] h-[22px] rounded-[5px] bg-[var(--active-bg)] text-[var(--text-link)] flex items-center justify-center text-xs font-bold shrink-0">
                    P
                  </span>
                  <span className="master-item-name">{p.name}</span>
                  <Button
                    size="sm"
                    variant={p.disabled ? 'outline' : 'destructive'}
                    className="ml-auto text-xs h-6 px-2"
                    onClick={(e) => { e.stopPropagation(); togglePluginMut.mutate(p); }}
                  >
                    {p.disabled ? 'Enable' : 'Disable'}
                  </Button>
                </div>
                <div className="master-item-sub">
                  <span className="tag tag-info">v{p.version}</span>
                  <span className="ml-1 text-xs text-[var(--text-secondary)]">{p.skills?.length || 0} skills</span>
                </div>
              </div>
            ))}
          </div>

          {selected ? (
            <div className="detail-panel">
              <h3>
                {selected.name}{' '}
                <span className="tag tag-info align-middle">v{selected.version}</span>
              </h3>
              <p className="text-sm text-[var(--text-secondary)] mt-1">{selected.skills?.length || 0} skills</p>

              {selected.skills?.map((s) => (
                <div key={s.name} className="bg-[var(--bg-secondary)] border border-[var(--border)] rounded-lg p-3 mt-3">
                  <div className="flex items-center justify-between mb-1">
                    <div className="flex items-center gap-2">
                      <Button
                        size="sm"
                        variant={s.disabled ? 'outline' : 'destructive'}
                        className="text-xs h-6 px-2"
                        onClick={() =>
                          toggleSkillMut.mutate({
                            pluginName: selected.name,
                            skillName: s.name,
                            installPath: selected.installPath || '',
                          })
                        }
                      >
                        {s.disabled ? 'Enable' : 'Disable'}
                      </Button>
                      <code>{s.invocation || '/' + s.name}</code>
                    </div>
                    <span className="tag tag-ok">{s.type}</span>
                  </div>
                  {s.description && <p className="text-sm mt-1">{s.description}</p>}
                  {s.descriptionCN ? (
                    <p className="text-xs text-[var(--text-secondary)] mt-1">{s.descriptionCN}</p>
                  ) : s.description ? (
                    <p className="text-xs text-[var(--text-secondary)] italic mt-1">(Original is Chinese)</p>
                  ) : null}
                </div>
              ))}
            </div>
          ) : (
            <div className="detail-panel empty-panel">Select a plugin to view details</div>
          )}
        </div>
      )}
    </div>
  );
}
