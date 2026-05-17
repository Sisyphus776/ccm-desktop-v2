"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
const electron_1 = require("electron");
const child_process_1 = require("child_process");
const path = __importStar(require("path"));
let mainWindow = null;
let goProcess = null;
let restartCount = 0;
const MAX_RESTARTS = 3;
const RPC_TIMEOUT = 30000; // 30s timeout per request
// Map of request ID → { resolve, timer } for concurrent RPC support
const pendingRequests = new Map();
let stdoutBuffer = '';
function sendToGo(line) {
    if (goProcess?.stdin?.writable) {
        goProcess.stdin.write(line + '\n');
    }
}
function handleGoStdout(data) {
    stdoutBuffer += data.toString();
    const lines = stdoutBuffer.split('\n');
    stdoutBuffer = lines.pop() || '';
    for (const line of lines) {
        if (!line.trim())
            continue;
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
            }
            else if (msg.method === 'ready') {
                restartCount = 0; // Reset on successful start
                mainWindow?.webContents.send('backend-ready');
            }
            else if (msg.method) {
                // Server notification
                mainWindow?.webContents.send('rpc-notify', msg.method, msg.params || {});
            }
        }
        catch {
            // Ignore parse errors on stdout lines
        }
    }
}
function startGoBackend() {
    const exePath = path.join(__dirname, 'ccm-backend.exe');
    const proc = (0, child_process_1.spawn)(exePath, [], {
        stdio: ['pipe', 'pipe', 'pipe'],
        windowsHide: true, // Fix: no console window flash
    });
    proc.stdout?.on('data', handleGoStdout);
    proc.stderr?.on('data', (data) => {
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
        }
        else if (restartCount >= MAX_RESTARTS) {
            mainWindow?.webContents.send('backend-crash', `Go backend crashed after ${MAX_RESTARTS} restarts`);
        }
    });
    return proc;
}
function createWindow() {
    mainWindow = new electron_1.BrowserWindow({
        width: 1200,
        height: 800,
        minWidth: 900,
        minHeight: 600,
        title: 'CCM — Claude Config Manager',
        backgroundColor: '#0c0a09',
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
    }
    else {
        mainWindow.loadFile(path.join(__dirname, '../frontend/dist/index.html'));
    }
    mainWindow.on('closed', () => {
        mainWindow = null;
    });
}
// IPC handlers — concurrent-safe RPC with Map<id, resolve> + timeout
electron_1.ipcMain.handle('rpc-request', async (_event, data) => {
    return new Promise((resolve, reject) => {
        let reqId;
        try {
            const req = JSON.parse(data);
            reqId = req.id;
        }
        catch {
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
electron_1.ipcMain.on('window-minimize', () => mainWindow?.minimize());
electron_1.ipcMain.on('window-maximize', () => {
    if (mainWindow?.isMaximized()) {
        mainWindow.unmaximize();
    }
    else {
        mainWindow?.maximize();
    }
});
electron_1.ipcMain.on('window-quit', () => electron_1.app.quit());
electron_1.app.whenReady().then(() => {
    goProcess = startGoBackend();
    createWindow();
    electron_1.app.on('before-quit', () => {
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
electron_1.app.on('window-all-closed', () => {
    electron_1.app.quit();
});
