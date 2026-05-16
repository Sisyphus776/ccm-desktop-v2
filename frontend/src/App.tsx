import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { HashRouter, Routes, Route } from 'react-router-dom';
import { Sidebar } from './components/Sidebar';
import { Toaster } from './components/ui/toaster';
import { lazy, Suspense, useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';

const pages = {
  Dashboard: lazy(() => import('./pages/Dashboard')),
  Skills: lazy(() => import('./pages/Skills')),
  Plugins: lazy(() => import('./pages/Plugins')),
  Memory: lazy(() => import('./pages/Memory')),
  Mcp: lazy(() => import('./pages/Mcp')),
  ClaudeMd: lazy(() => import('./pages/ClaudeMd')),
  Portability: lazy(() => import('./pages/Portability')),
  Secrets: lazy(() => import('./pages/Secrets')),
  Backup: lazy(() => import('./pages/Backup')),
  Settings: lazy(() => import('./pages/Settings')),
};

const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 30_000, retry: 1 } },
});

function AppShell() {
  const qc = useQueryClient();

  useEffect(() => {
    window.ccm.onNotify((method, params) => {
      if (method === 'config-changed') {
        const domain = params?.domain;
        if (domain === 'skills') {
          qc.invalidateQueries({ queryKey: ['skills'] });
          qc.invalidateQueries({ queryKey: ['plugins'] });
        } else if (domain === 'memory') {
          qc.invalidateQueries({ queryKey: ['memory'] });
        } else if (domain === 'settings') {
          qc.invalidateQueries({ queryKey: ['settings'] });
        } else {
          qc.invalidateQueries();
        }
      }
    });

    const paths = ['/', '/skills', '/plugins', '/memory', '/mcp', '/claudemd', '/portability', '/secrets', '/backup', '/settings'];
    const handler = (e: KeyboardEvent) => {
      if (e.ctrlKey && e.key >= '1' && e.key <= '9') {
        e.preventDefault();
        const idx = parseInt(e.key) - 1;
        if (idx < paths.length) window.location.hash = '#' + paths[idx];
      }
    };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [qc]);

  return (
    <div className="flex h-screen">
      <Sidebar />
      <main className="flex-1 overflow-y-auto bg-[var(--bg-primary)]">
        <Suspense fallback={<div className="p-8 text-[var(--text-secondary)]">Loading...</div>}>
          <Routes>
            <Route path="/" element={<pages.Dashboard />} />
            <Route path="/skills" element={<pages.Skills />} />
            <Route path="/plugins" element={<pages.Plugins />} />
            <Route path="/memory" element={<pages.Memory />} />
            <Route path="/mcp" element={<pages.Mcp />} />
            <Route path="/claudemd" element={<pages.ClaudeMd />} />
            <Route path="/portability" element={<pages.Portability />} />
            <Route path="/secrets" element={<pages.Secrets />} />
            <Route path="/backup" element={<pages.Backup />} />
            <Route path="/settings" element={<pages.Settings />} />
          </Routes>
        </Suspense>
      </main>
      <Toaster />
    </div>
  );
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <HashRouter>
        <AppShell />
      </HashRouter>
    </QueryClientProvider>
  );
}
