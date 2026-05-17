import { contextBridge, ipcRenderer } from 'electron';

contextBridge.exposeInMainWorld('ccm', {
  writeStdin: (data: string): Promise<string> => ipcRenderer.invoke('rpc-request', data),

  onNotify: (callback: (method: string, params: any) => void) => {
    const handler = (_event: any, method: string, params: any) => callback(method, params);
    ipcRenderer.on('rpc-notify', handler);
    return () => ipcRenderer.removeListener('rpc-notify', handler);
  },

  minimize: () => ipcRenderer.send('window-minimize'),
  quit: () => ipcRenderer.send('window-quit'),

  onBackendReady: (callback: () => void) => {
    const handler = () => callback();
    ipcRenderer.on('backend-ready', handler);
    return () => ipcRenderer.removeListener('backend-ready', handler);
  },

  onBackendCrash: (callback: (msg: string) => void) => {
    const handler = (_event: any, msg: string) => callback(msg);
    ipcRenderer.on('backend-crash', handler);
    return () => ipcRenderer.removeListener('backend-crash', handler);
  },
});
