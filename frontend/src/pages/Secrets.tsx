import { t } from '@/i18n';
import { useQuery } from '@tanstack/react-query';
import { rpcCall } from '@/lib/rpc-client';
import type { SecretItem } from '@/lib/types';

export default function Secrets() {
  const { data: secrets = [], isLoading } = useQuery<SecretItem[]>({
    queryKey: ['secrets'],
    queryFn: () => rpcCall('secrets.scan'),
  });

  return (
    <div className="content">
      <div className="page-header">
        <h2>{t('secrets.title')}</h2>
        <span className="refresh-hint">{secrets?.length ?? 0} findings</span>
      </div>

      {isLoading ? (
        <div className="p-8 text-center text-[var(--text-secondary)]">{t('common.loading')}</div>
      ) : secrets.length === 0 ? (
        <div className="text-[var(--text-secondary)] text-sm">No secrets detected</div>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Pattern</th>
              <th>Line</th>
              <th>Match (masked)</th>
              <th>File</th>
            </tr>
          </thead>
          <tbody>
            {secrets.map((s, i) => (
              <tr key={i}>
                <td><span className="tag tag-warn">{s.pattern}</span></td>
                <td>{s.line}</td>
                <td><code className="text-[#d2a8ff]">{s.match}</code></td>
                <td><code className="text-xs">{s.filePath}</code></td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
