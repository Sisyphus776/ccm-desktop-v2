let requestId = 0;

export async function rpcCall(method: string, params: any = {}): Promise<any> {
  const id = ++requestId;
  const request = JSON.stringify({ jsonrpc: '2.0', id, method, params });
  const raw = await window.ccm.writeStdin(request);
  const response = JSON.parse(raw);
  if (response.error) {
    throw new Error(response.error.message || 'RPC error');
  }
  return response.result;
}
