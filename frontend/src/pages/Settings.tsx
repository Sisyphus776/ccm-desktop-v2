import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { rpcCall } from '@/lib/rpc-client';
import type { AppSettings } from '@/lib/types';
import { cn } from '@/lib/utils';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';
import { themes, getTheme, setTheme } from '@/lib/theme';
import { t } from '@/i18n';

export default function Settings() {
  const { toast } = useToast();
  const qc = useQueryClient();
  const [currentTheme, setCurrentTheme] = useState(getTheme());

  const { data: settings, isLoading } = useQuery<AppSettings>({
    queryKey: ['settings'],
    queryFn: () => rpcCall('settings.get'),
  });

  const autoStartMut = useMutation({
    mutationFn: (enabled: boolean) => rpcCall('settings.set_autostart', { enabled }),
    onSuccess: (result: string) => {
      toast({ title: result });
      qc.invalidateQueries({ queryKey: ['settings'] });
    },
    onError: (err: Error) => {
      toast({ title: err.message, variant: 'destructive' });
    },
  });

  function switchTheme(id: string) {
    setCurrentTheme(id as typeof currentTheme);
    setTheme(id as typeof currentTheme);
  }

  return (
    <div className="content">
      <div className="page-header">
        <h2>{t('settings.title')}</h2>
      </div>

      <div className="space-y-3">
        {isLoading && <div className="p-8 text-center text-[var(--text-secondary)]">{t('common.loading')}</div>}
        {!isLoading && settings && (
          <>
            <div className="card">
              <div className="flex justify-between items-center">
                <div>
                  <h3 className="text-sm font-semibold">Claude {t('common.name')}</h3>
                  <div className="dim text-xs mt-1">{settings.claudeDir || 'Not set'}</div>
                </div>
              </div>
            </div>

            <div className="card">
              <div className="flex justify-between items-center">
                <div>
                  <h3 className="text-sm font-semibold">Home {t('common.name')}</h3>
                  <div className="dim text-xs mt-1">{settings.homeDir || 'Not set'}</div>
                </div>
              </div>
            </div>

            <div className="card">
              <div className="flex justify-between items-center">
                <div>
                  <h3 className="text-sm font-semibold">{t('common.autoStart')}</h3>
                  <div className="dim text-xs mt-1">{t('common.autoStartDesc')}</div>
                </div>
                <Button
                  size="sm"
                  variant={settings.autoStart ? 'destructive' : 'default'}
                  onClick={() => autoStartMut.mutate(!settings.autoStart)}
                >
                  {settings.autoStart ? t('common.disable') : t('common.enable')}
                </Button>
              </div>
            </div>

            <div className="card">
              <div className="flex justify-between items-center">
                <div>
                  <h3 className="text-sm font-semibold">{t('common.minimizeToTray')}</h3>
                  <div className="dim text-xs mt-1">{t('common.minimizeToTrayDesc')}</div>
                </div>
                <Button size="sm" variant="outline" onClick={() => window.ccm.minimize()}>
                  {t('common.minimize')}
                </Button>
              </div>
            </div>
          </>
        )}

        <h3 className="text-sm font-semibold text-[var(--text-primary)] mt-5">{t('common.theme')}</h3>
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3">
          {themes.map((t) => (
            <div
              key={t.id}
              className={cn(
                'border-2 rounded-lg p-3 cursor-pointer transition-all hover:border-[var(--text-link)] hover:bg-[var(--bg-tertiary)]',
                currentTheme === t.id
                  ? 'border-[var(--text-link)] bg-[var(--active-bg)]'
                  : 'border-[var(--border)]',
              )}
              onClick={() => switchTheme(t.id)}
            >
              <div className={cn(
                'h-12 rounded-md mb-2',
                t.id === 'oled' && 'bg-[#0a0a0a] border border-[#2a2a2a]',
                t.id === 'frost' && 'bg-[#0d1117] border border-[#30363d]',
                t.id === 'sepia' && 'bg-[#1e1a16] border border-[#3d3731]',
                t.id === 'monochrome' && 'bg-[#0f0f0f] border border-[#333]',
                t.id === 'neon' && 'bg-[#0a0a0f] border border-[#2a2a4a]',
              )} />
              <div className="text-sm font-semibold">{t.nameCN} {t.name}</div>
            </div>
          ))}
        </div>

        <div className="card mt-4">
          <h3 className="text-sm font-semibold">{t('common.about')}</h3>
          <div className="dim text-xs mt-1">
            CCM Desktop v2 — Electron + Go + React + shadcn/ui
          </div>
        </div>
      </div>
    </div>
  );
}
