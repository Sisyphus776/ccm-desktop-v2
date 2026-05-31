// Wails runtime types
interface Window {
  // Wails v2 Go bindings
  go?: {
    main?: {
      App?: {
        Call(method: string, paramsJSON: string): Promise<string>;
        TranslateBatch(): Promise<void>;
      };
    };
  };

  // Wails v2 runtime
  runtime?: {
    EventsOn(eventName: string, callback: (...args: any[]) => void): () => void;
    EventsOff(eventName: string): void;
    WindowMinimise(): void;
    WindowMaximise(): void;
    WindowUnmaximise(): void;
    Quit(): void;
  };

  // Electron mode (backward compat — remove when Electron is fully dropped)
  ccm?: {
    writeStdin(data: string): Promise<string>;
    onNotify(callback: (method: string, params: any) => void): () => void;
    minimize(): void;
    maximize(): void;
    quit(): void;
    onBackendReady(callback: () => void): () => void;
    onBackendCrash(callback: (msg: string) => void): () => void;
  };
}
