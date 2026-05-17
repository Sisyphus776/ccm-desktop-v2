import { t } from '@/i18n';
import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { rpcCall } from '@/lib/rpc-client';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';

export default function Backup() {
  const { toast } = useToast();
  const [outputPath, setOutputPath] = useState('');
  const [zipPath, setZipPath] = useState('');
  const [force, setForce] = useState(false);
  const [result, setResult] = useState('');

  const backupMut = useMutation({
    mutationFn: () => rpcCall('backup.create', { outputPath: outputPath || '' }),
    onSuccess: (data: string) => setResult(data),
    onError: (err) => toast({ title: String(err), variant: 'destructive' }),
  });

  const restoreMut = useMutation({
    mutationFn: () => {
      if (!zipPath) {
        toast({ title: 'Please enter a zip file path', variant: 'destructive' });
        throw new Error('no path');
      }
      return rpcCall('backup.restore', { zipPath, force });
    },
    onSuccess: (data: string) => setResult(data),
    onError: (err) => {
      if (err.message !== 'no path') toast({ title: String(err), variant: 'destructive' });
    },
  });

  return (
    <div className="content">
      <div className="page-header">
        <h2>{t('backup.title')}</h2>
      </div>

      <div className="space-y-6">
        <div>
          <h3 className="text-sm font-semibold text-[var(--text-primary)] mb-2">Create Backup</h3>
          <div className="flex gap-2">
            <Input
              value={outputPath}
              onChange={(e) => setOutputPath(e.target.value)}
              placeholder="Leave empty to save in user directory"
            />
            <Button onClick={() => backupMut.mutate()} disabled={backupMut.isPending}>
              Create Backup
            </Button>
          </div>
        </div>

        <div>
          <h3 className="text-sm font-semibold text-[var(--text-primary)] mb-2">Restore Backup</h3>
          <div className="flex flex-col gap-2 mb-2">
            <Input
              value={zipPath}
              onChange={(e) => setZipPath(e.target.value)}
              placeholder="Enter .zip file path"
            />
          </div>
          <div className="flex items-center gap-2 mb-2">
            <Checkbox id="force" checked={force} onCheckedChange={(v) => setForce(!!v)} />
            <label htmlFor="force" className="text-xs text-[var(--text-secondary)]">Overwrite existing files</label>
          </div>
          <Button
            variant={force ? 'destructive' : 'default'}
            onClick={() => restoreMut.mutate()}
            disabled={restoreMut.isPending}
          >
            Restore Backup
          </Button>
        </div>

        {result && (
          <div className="p-3 bg-[var(--bg-tertiary)] rounded-md">
            <code className="text-xs whitespace-pre-wrap">{result}</code>
          </div>
        )}
      </div>
    </div>
  );
}
