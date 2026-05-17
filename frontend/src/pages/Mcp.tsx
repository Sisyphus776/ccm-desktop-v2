import { t } from '@/i18n';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { rpcCall } from '@/lib/rpc-client';
import type { MCPItem, IssueItem } from '@/lib/types';
import { cn } from '@/lib/utils';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';

export default function Mcp() {
  const { toast } = useToast();
  const qc = useQueryClient();

  const { data: servers = [], isLoading } = useQuery<MCPItem[]>({
    queryKey: ['mcp'],
    queryFn: () => rpcCall('mcp.list'),
  });

  const { data: issues = [] } = useQuery<IssueItem[]>({
    queryKey: ['mcp', 'validate'],
    queryFn: () => rpcCall('mcp.validate'),
  });

  const toggleMut = useMutation({
    mutationFn: (name: string) => rpcCall('mcp.toggle', { name }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['mcp'] });
      qc.invalidateQueries({ queryKey: ['mcp', 'validate'] });
    },
    onError: (err) => toast({ title: String(err), variant: 'destructive' }),
  });

  return (
    <div className="content">
      <div className="page-header">
        <h2>{t('mcp.title')}</h2>
        <span className="refresh-hint">{servers?.length ?? 0} services</span>
      </div>

      {isLoading ? (
        <div className="p-8 text-center text-[var(--text-secondary)]">Loading...</div>
      ) : servers.length > 0 ? (
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Command</th>
              <th>Args</th>
              <th>Status</th>
              <th style={{ width: 80 }}>Action</th>
            </tr>
          </thead>
          <tbody>
            {servers.map((s) => (
              <tr key={s.name}>
                <td className="fw-medium">{s.name}</td>
                <td><code>{s.command}</code></td>
                <td className="dim text-xs">{s.args?.join(' ') || '-'}</td>
                <td>
                  <span className={cn(
                    'tag',
                    s.status === 'ok' ? 'tag-ok' : s.status === 'missing' ? 'tag-err' : 'tag-warn',
                  )}>
                    {s.status === 'ok' ? 'OK' : s.status === 'missing' ? 'Missing' : 'Warning'}
                  </span>
                  {s.disabled && <span className="tag tag-warn ml-1">Disabled</span>}
                </td>
                <td>
                  <Button
                    size="sm"
                    variant={s.disabled ? 'default' : 'destructive'}
                    className="text-xs h-6 px-2"
                    onClick={() => toggleMut.mutate(s.name)}
                  >
                    {s.disabled ? 'Enable' : 'Disable'}
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : (
        <div className="text-[var(--text-secondary)] text-sm">No MCP services configured</div>
      )}

      {issues.length > 0 && (
        <>
          <h3 className="text-sm font-semibold text-[var(--text-primary)] mt-6 mb-2">Validation Issues ({issues.length})</h3>
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
    </div>
  );
}
