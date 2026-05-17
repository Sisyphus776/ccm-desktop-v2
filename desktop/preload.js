"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const electron_1 = require("electron");
electron_1.contextBridge.exposeInMainWorld('ccm', {
    writeStdin: (data) => electron_1.ipcRenderer.invoke('rpc-request', data),
    onNotify: (callback) => {
        const handler = (_event, method, params) => callback(method, params);
        electron_1.ipcRenderer.on('rpc-notify', handler);
        return () => electron_1.ipcRenderer.removeListener('rpc-notify', handler);
    },
    minimize: () => electron_1.ipcRenderer.send('window-minimize'),
    maximize: () => electron_1.ipcRenderer.send('window-maximize'),
    quit: () => electron_1.ipcRenderer.send('window-quit'),
    onBackendReady: (callback) => {
        const handler = () => callback();
        electron_1.ipcRenderer.on('backend-ready', handler);
        return () => electron_1.ipcRenderer.removeListener('backend-ready', handler);
    },
    onBackendCrash: (callback) => {
        const handler = (_event, msg) => callback(msg);
        electron_1.ipcRenderer.on('backend-crash', handler);
        return () => electron_1.ipcRenderer.removeListener('backend-crash', handler);
    },
});
