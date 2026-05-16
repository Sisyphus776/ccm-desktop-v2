import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { rpcCall } from '@/lib/rpc-client';
import type { PortabilityResult } from '@/lib/types';
import { cn } from '@/lib/utils';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export default function Portability() {
  const { toast } = useToast();
  const qc = useQueryClient();

  const [fixIdx, setFixIdx] = useState(-1);
  const [newPath, setNewPath] = useState('');

  const { data: report, isLoading } = useQuery<PortabilityResult>({
    queryKey: ['portability'],
    queryFn: () => rpcCall('portability.report'),
  });

  const fixMut = useMutation({
    mutationFn: (args: { oldPath: string; newPath: string }) =>
      rpcCall('portability.fix_path', args),
    onSuccess: (result: string) => {
      toast({ title: result });
      setFixIdx(-1);
      setNewPath('');
      qc.invalidateQueries({ queryKey: ['portability'] });
    },
  });

  function getFixPath(iss: any) {
    const parts = iss.message?.split(':');
    return parts?.[1]?.trim() || iss.message || '';
  }

  return (
    <div className="content">
      <div className="page-header">
        <h2>Portability Analysis</h2>
        <span className="refresh-hint">Cross-machine migration assessment</span>
      </div>

      {isLoading ? (
        <div className="p-8 text-center text-[var(--text-secondary)]">Loading...</div>
      ) : (
        <>
          {report && (
            <div className="grid grid-cols-3 gap-3 mb-6">
              <div className="card text-center">
                <h3 className="text-xs font-semibold text-[var(--text-secondary)] uppercase mb-1">Critical</h3>
                <div className="text-2xl font-bold text-[var(--danger)]">{report.critical}</div>
              </div>
              <div className="card text-center">
                <h3 className="text-xs font-semibold text-[var(--text-secondary)] uppercase mb-1">Warning</h3>
                <div className="text-2xl font-bold text-[#d2a8ff]">{report.warning}</div>
              </div>
              <div className="card text-center">
                <h3 className="text-xs font-semibold text-[var(--text-secondary)] uppercase mb-1">Info</h3>
                <div className="text-2xl font-bold text-[var(--text-link)]">{report.info}</div>
              </div>
            </div>
          )}

          {(!report || report.issues.length === 0) ? (
            <div className="text-[var(--text-secondary)] text-sm">No portability issues found</div>
          ) : (
            report.issues.map((iss, i) => (
              <div key={i} className={cn(
                'p-3 rounded-md mb-2 text-sm',
                iss.severity === 'critical' ? 'bg-[var(--danger-bg)] border border-[var(--danger)]'
                  : iss.severity === 'warning' ? 'bg-[var(--active-bg)] border border-[#d2a8ff]'
                  : 'bg-[var(--bg-secondary)] border border-[var(--border)]',
              )}>
                <span className="font-semibold text-xs">[{iss.severity?.toUpperCase()}]</span>{' '}
                {iss.domain}: {iss.message}
                {iss.fix && <div className="text-xs text-[var(--text-secondary)] mt-1">{iss.fix}</div>}
                <div className="mt-2 flex gap-2">
                  <Button size="sm" variant="outline" className="text-xs" onClick={() => setFixIdx(i)}>
                    Fix
                  </Button>
                </div>
                {fixIdx === i && (
                  <div className="mt-2 flex gap-2 items-center">
                    <Input
                      value={newPath}
                      onChange={(e) => setNewPath(e.target.value)}
                      placeholder="Enter new path..."
                      className="flex-1"
                    />
                    <Button
                      size="sm"
                      onClick={() => fixMut.mutate({ oldPath: getFixPath(iss), newPath })}
                    >
                      Replace
                    </Button>
                    <Button size="sm" variant="outline" onClick={() => setFixIdx(-1)}>
                      Cancel
                    </Button>
                  </div>
                )}
              </div>
            ))
          )}

          {fixMut.data && <div className="text-sm text-[var(--text-secondary)] mt-2">{fixMut.data}</div>}
          {fixMut.error && <div className="text-sm text-[var(--danger)] mt-2">{String(fixMut.error)}</div>}
        </>
      )}
    </div>
  );
}
