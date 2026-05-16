export {};

declare global {
  interface Window {
    ccm: {
      writeStdin(data: string): Promise<string>;
      onNotify(callback: (method: string, params: any) => void): void;
      minimize(): void;
      quit(): void;
      onBackendReady(callback: () => void): void;
      onBackendCrash(callback: (msg: string) => void): void;
    };
  }
}
