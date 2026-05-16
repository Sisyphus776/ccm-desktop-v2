import { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import {
  LayoutDashboard, Wrench, Puzzle, Brain, Server, FileText,
  Shield, Key, Archive, Settings, Minus, Power,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';

const navItems = [
  { path: '/', icon: LayoutDashboard, label: 'Dashboard' },
  { path: '/skills', icon: Wrench, label: 'Skills' },
  { path: '/plugins', icon: Puzzle, label: 'Plugins' },
  { path: '/memory', icon: Brain, label: 'Memory' },
  { path: '/mcp', icon: Server, label: 'MCP' },
  { path: '/claudemd', icon: FileText, label: 'CLAUDE.md' },
  { path: '/portability', icon: Shield, label: 'Portability' },
  { path: '/secrets', icon: Key, label: 'Secrets' },
  { path: '/backup', icon: Archive, label: 'Backup' },
  { path: '/settings', icon: Settings, label: 'Settings' },
];

export function Sidebar() {
  const navigate = useNavigate();
  const location = useLocation();
  const [showQuit, setShowQuit] = useState(false);

  return (
    <>
      <aside className="flex flex-col items-center w-[52px] h-screen bg-[var(--bg-secondary)] border-r border-[var(--border)] py-3 shrink-0">
        {/* Nav icons */}
        <nav className="flex flex-col gap-1 flex-1">
          {navItems.map((item) => {
            const Icon = item.icon;
            const active = item.path === '/'
              ? location.pathname === '/'
              : location.pathname.startsWith(item.path);
            return (
              <button
                key={item.path}
                title={item.label}
                onClick={() => navigate(item.path)}
                className={cn(
                  'w-9 h-9 flex items-center justify-center rounded-lg transition-colors',
                  active
                    ? 'bg-[var(--active-bg)] text-[var(--text-link)]'
                    : 'text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-tertiary)]',
                )}
              >
                <Icon size={18} />
              </button>
            );
          })}
        </nav>

        {/* Bottom action buttons */}
        <div className="flex flex-col gap-1">
          <button
            title="Minimize"
            onClick={() => window.ccm.minimize()}
            className="w-9 h-9 flex items-center justify-center rounded-lg text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-tertiary)] transition-colors"
          >
            <Minus size={18} />
          </button>
          <button
            title="Quit"
            onClick={() => setShowQuit(true)}
            className="w-9 h-9 flex items-center justify-center rounded-lg text-[var(--danger)] hover:bg-[var(--danger-bg)] transition-colors"
          >
            <Power size={18} />
          </button>
        </div>
      </aside>

      {/* Quit confirmation dialog */}
      <Dialog open={showQuit} onOpenChange={setShowQuit}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Quit CCM Desktop?</DialogTitle>
            <DialogDescription>
              This will close the application. Any unsaved changes will be preserved.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowQuit(false)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={() => {
                window.ccm.quit();
                setShowQuit(false);
              }}
            >
              Quit
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
