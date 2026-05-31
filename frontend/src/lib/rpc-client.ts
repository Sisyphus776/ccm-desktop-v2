let requestId = 0;

export async function rpcCall(method: string, params: any = {}): Promise<any> {
  const id = ++requestId;
  const paramsJSON = JSON.stringify(params);

  // Wails mode: call Go backend via the bound App.Call method
  if (window.go?.main?.App?.Call) {
    const raw = await window.go.main.App.Call(method, paramsJSON);
    const response = JSON.parse(raw);
    if (response.error) {
      throw new Error(response.error.message || 'RPC error');
    }
    return response.result;
  }

  // Fallback: Electron mode (window.ccm.writeStdin)
  if (window.ccm?.writeStdin) {
    const request = JSON.stringify({ jsonrpc: '2.0', id, method, params });
    const raw = await window.ccm.writeStdin(request);
    const response = JSON.parse(raw);
    if (response.error) {
      throw new Error(response.error.message || 'RPC error');
    }
    return response.result;
  }

  throw new Error('No backend available');
}
