import type { Input, ServerMessage, Welcome, Snapshot, Left, ErrorMsg } from '../types';

export interface ConnectionCallbacks {
  onWelcome?: (msg: Welcome) => void;
  onSnapshot?: (msg: Snapshot) => void;
  onLeft?: (msg: Left) => void;
  onError?: (msg: ErrorMsg) => void;
  onClose?: () => void;
}

const BACKOFF_SCHEDULE_MS = [500, 1000, 2000, 4000, 5000] as const;

/**
 * Connection owns one WebSocket and its reconnect lifecycle.
 *
 * Callbacks can be replaced at any time via setCallbacks(). This is the
 * intentional integration seam between React (which re-renders) and the
 * long-lived socket: instead of tearing down the socket every time a
 * React effect re-runs, the React layer just swaps the callback object.
 *
 * Reconnect is intentionally minimal: it does NOT re-send hello. The owning
 * caller decides what to do after a reconnect (re-hello, surface an error,
 * etc.) — the connection only reopens the transport.
 */
export class Connection {
  private readonly url: string;
  private callbacks: ConnectionCallbacks;
  private socket: WebSocket | null = null;
  private intentional = false;
  private reconnectAttempt = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;

  constructor(url: string, callbacks: ConnectionCallbacks) {
    this.url = url;
    this.callbacks = callbacks;
  }

  setCallbacks(callbacks: ConnectionCallbacks): void {
    this.callbacks = callbacks;
  }

  connect(): void {
    this.intentional = false;
    this.openSocket();
  }

  sendHello(room: string, name: string): void {
    this.send({ type: 'hello', room, name });
  }

  sendInput(input: Omit<Input, 'type'>): void {
    this.send({ type: 'input', ...input });
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
      case 'snapshot':
        this.callbacks.onSnapshot?.(msg);
        return;
      case 'left':
        this.callbacks.onLeft?.(msg);
        return;
      case 'error':
        this.callbacks.onError?.(msg);
        return;
    }
  }

  private send(payload: object): void {
    if (this.socket === null || this.socket.readyState !== WebSocket.OPEN) return;
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
  return t === 'welcome' || t === 'snapshot' || t === 'left' || t === 'error';
}
