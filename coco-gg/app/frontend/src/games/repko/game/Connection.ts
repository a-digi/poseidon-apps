import type {
  Welcome,
  StateMsg,
  ErrorMsg,
  ServerMessage,
  ClientAction,
  Hello,
} from '../types';

export interface ConnectionCallbacks {
  onWelcome?: (msg: Welcome) => void;
  onState?: (msg: StateMsg) => void;
  onError?: (msg: ErrorMsg) => void;
  onClose?: () => void;
}

export interface ConnectionOptions {
  disableAutoReconnect?: boolean;
}

const BACKOFF_SCHEDULE_MS = [500, 1000, 2000, 4000, 5000] as const;

/**
 * Connection owns one WebSocket and its reconnect lifecycle.
 *
 * Callbacks can be replaced at any time via setCallbacks() — the React layer
 * swaps the callback object instead of tearing the socket down on re-renders.
 *
 * Reconnect does NOT re-send hello: the caller decides post-reconnect behaviour.
 */
export class Connection {
  private readonly url: string;
  private callbacks: ConnectionCallbacks;
  private options: ConnectionOptions;
  private socket: WebSocket | null = null;
  private intentional = false;
  private reconnectAttempt = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private pendingMessages: object[] = [];

  constructor(url: string, callbacks: ConnectionCallbacks, options?: ConnectionOptions) {
    this.url = url;
    this.callbacks = callbacks;
    this.options = options ?? {};
  }

  setCallbacks(callbacks: ConnectionCallbacks): void {
    this.callbacks = callbacks;
  }

  connect(): void {
    this.intentional = false;
    this.openSocket();
  }

  sendHello(room: string, name: string): void {
    const hello: Hello = { type: 'hello', room, name };
    this.send(hello);
  }

  sendAction(action: ClientAction): void {
    this.send(action);
  }

  disconnect(): void {
    this.intentional = true;
    if (this.reconnectTimer !== null) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.socket !== null) {
      const s = this.socket;
      this.socket = null;
      try {
        s.close(1000);
      } catch {
        /* ignore */
      }
    }
  }

  private openSocket(): void {
    const s = new WebSocket(this.url);
    this.socket = s;

    s.onopen = () => {
      this.reconnectAttempt = 0;
      const queued = this.pendingMessages;
      this.pendingMessages = [];
      for (const msg of queued) {
        try { s.send(JSON.stringify(msg)); } catch { /* socket closed mid-send */ }
      }
    };

    s.onmessage = (event) => {
      let parsed: unknown;
      try {
        parsed = JSON.parse(event.data as string);
      } catch {
        return;
      }
      if (!isServerMessage(parsed)) return;
      this.dispatch(parsed);
    };

    s.onerror = () => {
      /* swallow — onclose follows and drives reconnect */
    };

    s.onclose = (event) => {
      if (this.socket === s) {
        this.socket = null;
      }
      this.callbacks.onClose?.();
      const clean = event.wasClean && event.code === 1000;
      if (this.intentional || clean) return;
      if (this.options.disableAutoReconnect) return;
      this.scheduleReconnect();
    };
  }

  private scheduleReconnect(): void {
    const idx = Math.min(this.reconnectAttempt, BACKOFF_SCHEDULE_MS.length - 1);
    const delay = BACKOFF_SCHEDULE_MS[idx];
    this.reconnectAttempt += 1;
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      if (this.intentional) return;
      this.openSocket();
    }, delay);
  }

  private dispatch(msg: ServerMessage): void {
    switch (msg.type) {
      case 'welcome':
        this.callbacks.onWelcome?.(msg);
        return;
      case 'state':
        this.callbacks.onState?.(msg);
        return;
      case 'error':
        this.callbacks.onError?.(msg);
        return;
    }
  }

  private send(payload: object): void {
    if (this.socket === null || this.socket.readyState === WebSocket.CONNECTING) {
      this.pendingMessages.push(payload);
      return;
    }
    if (this.socket.readyState !== WebSocket.OPEN) return;
    try {
      this.socket.send(JSON.stringify(payload));
    } catch {
      /* socket closed mid-send — onclose will handle */
    }
  }
}

function isServerMessage(value: unknown): value is ServerMessage {
  if (typeof value !== 'object' || value === null) return false;
  const t = (value as { type?: unknown }).type;
  return t === 'welcome' || t === 'state' || t === 'error';
}
