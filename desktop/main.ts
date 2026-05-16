import { app, BrowserWindow, ipcMain } from 'electron';
import { spawn, ChildProcess } from 'child_process';
import * as path from 'path';

let mainWindow: BrowserWindow | null = null;
let goProcess: ChildProcess | null = null;
let restartCount = 0;
const MAX_RESTARTS = 3;
let pendingResolve: ((line: string) => void) | null = null;
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
        // Response to a pending request
        if (pendingResolve) {
          pendingResolve(JSON.stringify(msg));
          pendingResolve = null;
        }
      } else if (msg.method === 'ready') {
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
  });

  proc.stdout?.on('data', handleGoStdout);

  proc.stderr?.on('data', (data: Buffer) => {
    console.error('Go stderr:', data.toString());
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
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
    },
  });

  // In development, load Vite dev server
  if (process.env.NODE_ENV === 'development') {
    mainWindow.loadURL('http://localhost:5173');
  } else {
    mainWindow.loadFile(path.join(__dirname, '../frontend/dist/index.html'));
  }

  mainWindow.on('closed', () => {
    mainWindow = null;
  });
}

// IPC handlers
ipcMain.handle('rpc-request', async (_event, data: string) => {
  return new Promise((resolve) => {
    pendingResolve = resolve;
    sendToGo(data);
  });
});

ipcMain.on('window-minimize', () => mainWindow?.minimize());
ipcMain.on('window-quit', () => app.quit());

app.whenReady().then(() => {
  goProcess = startGoBackend();
  createWindow();

  app.on('before-quit', () => {
    if (goProcess) {
      goProcess.kill();
      goProcess = null;
    }
  });
});

app.on('window-all-closed', () => {
  app.quit();
});
