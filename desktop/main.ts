import { app, BrowserWindow, ipcMain } from 'electron';
import { spawn, ChildProcess } from 'child_process';
import * as path from 'path';

let mainWindow: BrowserWindow | null = null;
let goProcess: ChildProcess | null = null;
let restartCount = 0;
const MAX_RESTARTS = 3;
const RPC_TIMEOUT = 30000; // 30s timeout per request

// Map of request ID → { resolve, timer } for concurrent RPC support
const pendingRequests = new Map<number, { resolve: (line: string) => void; timer: NodeJS.Timeout }>();
let stdoutBuffer = '';

function sendToGo(line: string) {
  if (goProcess?.stdin?.writable) {
    goProcess.stdin.write(line + '\n');
  }
}

function handleGoStdout(data: Buffer) {
  stdoutBuffer += data.toString();
  const lines = stdoutBuffer.split('\n');
  stdoutBuffer = lines.pop() || '';
  for (const line of lines) {
    if (!line.trim()) continue;
    try {
      const msg = JSON.parse(line);
      if (msg.id !== undefined) {
        // Response to a pending request — match by id
        const entry = pendingRequests.get(msg.id);
        if (entry) {
          clearTimeout(entry.timer);
          pendingRequests.delete(msg.id);
          entry.resolve(JSON.stringify(msg));
        }
      } else if (msg.method === 'ready') {
        restartCount = 0; // Reset on successful start
        mainWindow?.webContents.send('backend-ready');
      } else if (msg.method) {
        // Server notification
        mainWindow?.webContents.send('rpc-notify', msg.method, msg.params || {});
      }
    } catch {
      // Ignore parse errors on stdout lines
    }
  }
}

function startGoBackend(): ChildProcess {
  const exePath = path.join(__dirname, 'ccm-backend.exe');
  const proc = spawn(exePath, [], {
    stdio: ['pipe', 'pipe', 'pipe'],
    windowsHide: true, // Fix: no console window flash
  });

  proc.stdout?.on('data', handleGoStdout);

  proc.stderr?.on('data', (data: Buffer) => {
    console.error('Go stderr:', data.toString());
  });

  proc.on('error', (err) => {
    console.error('Failed to spawn Go backend:', err.message);
    mainWindow?.webContents.send('backend-crash', `Cannot start backend: ${err.message}`);
  });

  proc.on('exit', (code) => {
    if (code !== 0 && restartCount < MAX_RESTARTS) {
      restartCount++;
      console.log(`Go backend exited with code ${code}, restarting (${restartCount}/${MAX_RESTARTS})...`);
      setTimeout(() => {
        goProcess = startGoBackend();
      }, 1000);
    } else if (restartCount >= MAX_RESTARTS) {
      mainWindow?.webContents.send('backend-crash', `Go backend crashed after ${MAX_RESTARTS} restarts`);
    }
  });

  return proc;
}

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    minWidth: 900,
    minHeight: 600,
    title: 'CCM — Claude Config Manager',
    backgroundColor: '#0a0a0a',
    frame: false, // Custom title bar for theme sync
    titleBarStyle: 'hidden',
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
    },
  });

  if (process.env.NODE_ENV === 'development') {
    mainWindow.loadURL('http://localhost:5173');
  } else {
    mainWindow.loadFile(path.join(__dirname, '../frontend/dist/index.html'));
  }

  mainWindow.on('closed', () => {
    mainWindow = null;
  });
}

// IPC handlers — concurrent-safe RPC with Map<id, resolve> + timeout
ipcMain.handle('rpc-request', async (_event, data: string) => {
  return new Promise<string>((resolve, reject) => {
    let reqId: number;
    try {
      const req = JSON.parse(data);
      reqId = req.id;
    } catch {
      reject(new Error('Invalid JSON-RPC request'));
      return;
    }
    const timer = setTimeout(() => {
      pendingRequests.delete(reqId);
      reject(new Error(`RPC request ${reqId} timed out after ${RPC_TIMEOUT}ms`));
    }, RPC_TIMEOUT);
    pendingRequests.set(reqId, { resolve, timer });
    sendToGo(data);
  });
});

ipcMain.on('window-minimize', () => mainWindow?.minimize());
ipcMain.on('window-quit', () => app.quit());

app.whenReady().then(() => {
  goProcess = startGoBackend();
  createWindow();

  app.on('before-quit', () => {
    // Clear all pending requests
    for (const [id, entry] of pendingRequests) {
      clearTimeout(entry.timer);
      entry.resolve(JSON.stringify({ jsonrpc: '2.0', id, error: { code: -32603, message: 'App shutting down' } }));
    }
    pendingRequests.clear();
    if (goProcess) {
      goProcess.kill();
      goProcess = null;
    }
  });
});

app.on('window-all-closed', () => {
  app.quit();
});
