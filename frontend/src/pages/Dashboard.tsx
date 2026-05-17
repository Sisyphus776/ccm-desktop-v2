import { t } from '@/i18n';
import { useQuery } from '@tanstack/react-query';
import { rpcCall } from '@/lib/rpc-client';
import type { DashboardSummary, SkillItem, IssueItem } from '@/lib/types';
import { cn } from '@/lib/utils';

export default function Dashboard() {
  const { data: summary, isLoading: sLoading } = useQuery<DashboardSummary>({
    queryKey: ['dashboard'],
    queryFn: () => rpcCall('dashboard.summary'),
  });

  const { data: skills, isLoading: skLoading } = useQuery<SkillItem[]>({
    queryKey: ['skills'],
    queryFn: () => rpcCall('skills.list'),
  });

  const { data: issues, isLoading: issLoading } = useQuery<IssueItem[]>({
    queryKey: ['issues'],
    queryFn: () => rpcCall('skills.validate'),
  });

  const isLoading = sLoading || skLoading || issLoading;

  return (
    <div className="content">
      <div className="page-header">
        <h2>{t('dashboard.title')}</h2>
        <span className="refresh-hint">Auto-refresh</span>
      </div>

      {isLoading ? (
        <div className="p-8 text-center text-[var(--text-secondary)]">Loading...</div>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-5 gap-3 mb-6">
            <div className="card text-center">
              <h3 className="text-xs font-semibold text-[var(--text-secondary)] uppercase mb-1">Skills</h3>
              <div className="text-2xl font-bold text-[var(--text-link)]">{summary?.skillsCount ?? 0}</div>
            </div>
            <div className="card text-center">
              <h3 className="text-xs font-semibold text-[var(--text-secondary)] uppercase mb-1">Memory</h3>
              <div className="text-2xl font-bold text-[var(--text-link)]">{summary?.memoryCount ?? 0}</div>
            </div>
            <div className="card text-center">
              <h3 className="text-xs font-semibold text-[var(--text-secondary)] uppercase mb-1">MCP</h3>
              <div className="text-2xl font-bold text-[var(--text-link)]">{summary?.mcpServers ?? 0}</div>
            </div>
            <div className="card text-center">
              <h3 className="text-xs font-semibold text-[var(--text-secondary)] uppercase mb-1">Warnings</h3>
              <div className={cn('text-2xl font-bold', (summary?.warningCount ?? 0) > 0 ? 'text-[#d2a8ff]' : 'text-[var(--success)]')}>
                {summary?.warningCount ?? 0}
              </div>
            </div>
            <div className="card text-center">
              <h3 className="text-xs font-semibold text-[var(--text-secondary)] uppercase mb-1">Errors</h3>
              <div className={cn('text-2xl font-bold', (summary?.errorCount ?? 0) > 0 ? 'text-[var(--danger)]' : 'text-[var(--success)]')}>
                {summary?.errorCount ?? 0}
              </div>
            </div>
          </div>

          {skills && skills.length > 0 && (
            <>
              <h3 className="text-sm font-semibold text-[var(--text-primary)] mb-2">Installed Skills</h3>
              <table className="mb-6">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Invocation</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {skills.slice(0, 10).map((s) => (
                    <tr key={s.name}>
                      <td>{s.name}</td>
                      <td><code>{s.invocation || '/' + s.name}</code></td>
                      <td>
                        <span className={cn(
                          'tag',
                          s.status === 'ok' ? 'tag-ok' : s.status === 'broken' ? 'tag-err' : 'tag-warn',
                        )}>
                          {s.status === 'ok' ? 'OK' : s.status === 'broken' ? 'Broken' : 'Warning'}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </>
          )}

          {issues && issues.length > 0 && (
            <>
              <h3 className="text-sm font-semibold text-[var(--text-primary)] mb-2">Recent Issues ({issues.length})</h3>
              {issues.slice(0, 20).map((iss, i) => (
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

          {(!issues || issues.length === 0) && (
            <div className="text-[var(--text-secondary)] text-sm">No issues found</div>
          )}
        </>
      )}
    </div>
  );
}
