import { contextBridge, ipcRenderer } from 'electron';

contextBridge.exposeInMainWorld('ccm', {
  writeStdin: (data: string): Promise<string> => ipcRenderer.invoke('rpc-request', data),
  onStdout: (callback: (line: string) => void) => {
    ipcRenderer.on('rpc-response', (_event, line) => callback(line));
  },
  onNotify: (callback: (method: string, params: any) => void) => {
    ipcRenderer.on('rpc-notify', (_event, method, params) => callback(method, params));
  },
  minimize: () => ipcRenderer.send('window-minimize'),
  quit: () => ipcRenderer.send('window-quit'),
  onBackendReady: (callback: () => void) => {
    ipcRenderer.on('backend-ready', () => callback());
  },
  onBackendCrash: (callback: (msg: string) => void) => {
    ipcRenderer.on('backend-crash', (_event, msg) => callback(msg));
  },
});
