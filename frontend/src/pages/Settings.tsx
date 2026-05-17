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
                t.id === 'lacquer' && 'bg-[#0c0a09] border border-[#2a2520]',
                t.id === 'alabaster' && 'bg-gradient-to-br from-[#faf8f5] to-[#f3f0eb] border border-[#d9d3c8]',
                t.id === 'slate' && 'bg-gradient-to-br from-[#0b0e14] to-[#181d28] border border-[#232b3b]',
                t.id === 'photon' && 'bg-gradient-to-br from-[#f8fafd] to-[#f1f5f9] border border-[#dde4ef]',
                t.id === 'obsidian' && 'bg-gradient-to-br from-[#06050c] to-[#14101e] border border-[#221d30]',
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
