import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { rpcCall } from '@/lib/rpc-client';
import type { AppSettings } from '@/lib/types';
import { cn } from '@/lib/utils';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';

const themes = [
  { id: 'oled-dark', name: 'OLED Dark', desc: 'Linear/Raycast style' },
  { id: 'glassmorphism', name: 'Glassmorphism', desc: 'Purple-blue gradient translucent' },
  { id: 'warm-minimal', name: 'Warm Minimal', desc: 'Notion style' },
  { id: 'neubrutalism', name: 'Neubrutalism', desc: 'Bold borders hard shadows' },
  { id: 'terminal-green', name: 'Terminal Green', desc: 'Matrix retro' },
];

export default function Settings() {
  const { toast } = useToast();
  const qc = useQueryClient();
  const [currentTheme, setCurrentTheme] = useState(
    localStorage.getItem('ccm-theme') || 'oled-dark',
  );

  const { data: settings } = useQuery<AppSettings>({
    queryKey: ['settings'],
    queryFn: () => rpcCall('settings.get'),
  });

  const autoStartMut = useMutation({
    mutationFn: (enabled: boolean) => rpcCall('settings.set_autostart', { enabled }),
    onSuccess: (result: string) => {
      toast({ title: result });
      qc.invalidateQueries({ queryKey: ['settings'] });
    },
  });

  function switchTheme(id: string) {
    setCurrentTheme(id);
    localStorage.setItem('ccm-theme', id);
    document.documentElement.setAttribute('data-theme', id);
  }

  return (
    <div className="content">
      <div className="page-header">
        <h2>Settings</h2>
      </div>

      <div className="space-y-3">
        {/* Directories */}
        {settings && (
          <>
            <div className="card">
              <div className="flex justify-between items-center">
                <div>
                  <h3 className="text-sm font-semibold">Claude Directory</h3>
                  <div className="dim text-xs mt-1">{settings.claudeDir || 'Not set'}</div>
                </div>
              </div>
            </div>

            <div className="card">
              <div className="flex justify-between items-center">
                <div>
                  <h3 className="text-sm font-semibold">Home Directory</h3>
                  <div className="dim text-xs mt-1">{settings.homeDir || 'Not set'}</div>
                </div>
              </div>
            </div>

            {/* Auto-start toggle */}
            <div className="card">
              <div className="flex justify-between items-center">
                <div>
                  <h3 className="text-sm font-semibold">Auto Start</h3>
                  <div className="dim text-xs mt-1">Start on system boot</div>
                </div>
                <Button
                  size="sm"
                  variant={settings.autoStart ? 'destructive' : 'default'}
                  onClick={() => autoStartMut.mutate(!settings.autoStart)}
                >
                  {settings.autoStart ? 'Disable' : 'Enable'}
                </Button>
              </div>
            </div>

            {/* Minimize to tray */}
            <div className="card">
              <div className="flex justify-between items-center">
                <div>
                  <h3 className="text-sm font-semibold">Minimize to Tray</h3>
                  <div className="dim text-xs mt-1">Hide window to system tray</div>
                </div>
                <Button size="sm" variant="outline" onClick={() => window.ccm.minimize()}>
                  Minimize
                </Button>
              </div>
            </div>
          </>
        )}

        {/* Theme selector */}
        <h3 className="text-sm font-semibold text-[var(--text-primary)] mt-5">Theme</h3>
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
                t.id === 'oled-dark' && 'bg-[#0a0e14] border border-[#1e2a3a]',
                t.id === 'glassmorphism' && 'bg-gradient-to-br from-[#0f0f23] to-[#1a1a3e] border border-white/10',
                t.id === 'warm-minimal' && 'bg-[#FFFBFA] border border-[#E6E4E1]',
                t.id === 'neubrutalism' && 'bg-[#FEFAE0] border-[3px] border-black shadow-[3px_3px_0px_black]',
                t.id === 'terminal-green' && 'bg-black border border-[#003300]',
              )} />
              <div className="text-sm font-semibold">{t.name}</div>
              <div className="text-xs text-[var(--text-secondary)] mt-1">{t.desc}</div>
            </div>
          ))}
        </div>

        {/* About */}
        <div className="card mt-4">
          <h3 className="text-sm font-semibold">About</h3>
          <div className="dim text-xs mt-1">
            CCM Desktop v2.1 -- Lightweight Claude Config Manager -- 5 switchable themes
          </div>
        </div>
      </div>
    </div>
  );
}
