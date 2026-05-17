import { t } from '@/i18n';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { rpcCall } from '@/lib/rpc-client';
import type { MCPItem, IssueItem } from '@/lib/types';
import { cn } from '@/lib/utils';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';

const statusLabels: Record<string, string> = {
  ok: '正常',
  warning: '警告',
  missing: '缺失',
};

export default function Mcp() {
  const { toast } = useToast();
  const qc = useQueryClient();

  const { data: servers, isLoading } = useQuery<MCPItem[]>({
    queryKey: ['mcp'],
    queryFn: () => rpcCall('mcp.list'),
  });
  const safeServers = servers || [];

  const { data: issues } = useQuery<IssueItem[]>({
    queryKey: ['mcp', 'validate'],
    queryFn: () => rpcCall('mcp.validate'),
  });
  const safeIssues = issues || [];

  const toggleMut = useMutation({
    mutationFn: (name: string) => rpcCall('mcp.toggle', { name }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['mcp'] }); qc.invalidateQueries({ queryKey: ['mcp', 'validate'] }); },
    onError: (err) => toast({ title: String(err), variant: 'destructive' }),
  });

  return (
    <div className="content">
      <div className="page-header">
        <h2>{t('mcp.title')}</h2>
        <span className="refresh-hint">{safeServers.length} 个服务</span>
      </div>

      {isLoading ? (
        <div className="p-8 text-center text-[var(--text-secondary)]">{t('common.loading')}</div>
      ) : safeServers.length > 0 ? (
        <table>
          <thead>
            <tr>
              <th>{t('common.name')}</th>
              <th>命令</th>
              <th>参数</th>
              <th>{t('common.status')}</th>
              <th style={{ width: 80 }}>{t('common.actions')}</th>
            </tr>
          </thead>
          <tbody>
            {safeServers.map((s) => (
              <tr key={s.name}>
                <td className="fw-medium">{s.name}</td>
                <td><code>{s.command}</code></td>
                <td className="dim text-xs">{s.args?.join(' ') || '-'}</td>
                <td>
                  <span className={cn('tag', s.status === 'ok' ? 'tag-ok' : s.status === 'missing' ? 'tag-err' : 'tag-warn')}>
                    {statusLabels[s.status] || s.status}
                  </span>
                  {s.disabled && <span className="tag tag-warn ml-1">已禁用</span>}
                </td>
                <td>
                  <Button size="sm" variant={s.disabled ? 'default' : 'destructive'} className="text-xs h-6 px-2" onClick={() => toggleMut.mutate(s.name)}>
                    {s.disabled ? t('common.enable') : t('common.disable')}
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : (
        <div className="text-[var(--text-secondary)] text-sm">未配置 MCP 服务</div>
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
              {iss.fix && <div className="text-xs text-[var(--text-secondary)] mt-1">{iss.fix}</div>}
            </div>
          ))}
        </>
      )}
    </div>
  );
}
