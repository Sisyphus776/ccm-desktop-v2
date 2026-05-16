import { useState } from 'react';
import { useQuery, useMutation } from '@tanstack/react-query';
import { rpcCall } from '@/lib/rpc-client';
import type { ClaudeMDItem } from '@/lib/types';
import { cn } from '@/lib/utils';

export default function ClaudeMd() {
  const [selectedPath, setSelectedPath] = useState<string | null>(null);
  const [content, setContent] = useState('');

  const { data: mds = [], isLoading } = useQuery<ClaudeMDItem[]>({
    queryKey: ['claudemd'],
    queryFn: () => rpcCall('claudemd.list'),
  });

  const viewMut = useMutation({
    mutationFn: (path: string) => rpcCall('claudemd.get_content', { path }),
    onSuccess: (data: string) => setContent(data),
  });

  function viewContent(path: string) {
    setSelectedPath(path);
    viewMut.mutate(path);
  }

  return (
    <div className="content">
      <div className="page-header">
        <h2>CLAUDE.md Management</h2>
        <span className="refresh-hint">{mds?.length ?? 0} files</span>
      </div>

      {isLoading ? (
        <div className="p-8 text-center text-[var(--text-secondary)]">Loading...</div>
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
                  onClick={() => viewContent(md.path)}
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
              <h3><code className="text-sm">{selectedPath}</code></h3>
              <pre className="bg-[var(--bg-tertiary)] p-3 rounded-md text-xs mt-3 max-h-[600px] overflow-y-auto whitespace-pre-wrap font-mono">
                {content}
              </pre>
            </div>
          ) : (
            <div className="detail-panel empty-panel">Select a file to view content</div>
          )}
        </div>
      )}
    </div>
  );
}
