type Mode = 'full' | 'minimal';

declare global {
  interface Window {
    __pluginCtx?: Record<string, PluginCtx> & { __activeId?: string };
  }
}

type PluginCtx = {
  requestMinimalMode(): void;
  releaseMinimalMode(): void;
  subscribeMode(cb: (mode: Mode) => void): () => void;
};

function isInsideIframe(): boolean {
  return window !== window.parent;
}

function getIframePluginId(): string {
  return new URLSearchParams(window.location.search).get('pluginId') ?? '';
}

function getJsPluginCtx(): PluginCtx | undefined {
  const activeId = window.__pluginCtx?.__activeId;
  if (!activeId) return undefined;
  return window.__pluginCtx?.[activeId];
}

// Register the message listener at module-load time so mode messages sent
// before any subscriber calls subscribeMode are never dropped (race condition
// between onLoad on the host side and useEffect timing on the plugin side).
let lastMode: Mode | null = null;
const modeCallbacks = new Set<(mode: Mode) => void>();

let lastTabActivePluginId: string | null = null;
let lastTabInactivePluginId: string | null = null;
const tabActiveCallbacks = new Set<(pluginId: string) => void>();
const tabInactiveCallbacks = new Set<(pluginId: string) => void>();
const tabClosedCallbacks = new Set<(pluginId: string) => void>();

window.addEventListener('message', (event: MessageEvent) => {
  if (event.data?.type === 'plugin:mode') {
    lastMode = event.data.mode as Mode;
    modeCallbacks.forEach(cb => cb(lastMode!));
  } else if (event.data?.type === 'plugin:tab:active') {
    lastTabActivePluginId = event.data.activePluginId as string;
    tabActiveCallbacks.forEach(cb => cb(lastTabActivePluginId!));
  } else if (event.data?.type === 'plugin:tab:inactive') {
    lastTabInactivePluginId = event.data.inactivePluginId as string;
    tabInactiveCallbacks.forEach(cb => cb(lastTabInactivePluginId!));
  } else if (event.data?.type === 'plugin:tab:closed') {
    const closedId = event.data.closedPluginId as string;
    tabClosedCallbacks.forEach(cb => cb(closedId));
  }
});

export function requestMinimalMode(): void {
  if (isInsideIframe()) {
    const pluginId = getIframePluginId();
    window.parent.postMessage({ type: 'plugin:minimal:request', pluginId }, '*');
  } else {
    getJsPluginCtx()?.requestMinimalMode();
  }
}

export function startDrag(offsetX: number, offsetY: number): void {
  if (isInsideIframe()) {
    const pluginId = getIframePluginId();
    window.parent.postMessage({ type: 'plugin:minimal:dragstart', pluginId, offsetX, offsetY }, '*');
  }
}

export function releaseMinimalMode(): void {
  if (isInsideIframe()) {
    const pluginId = getIframePluginId();
    window.parent.postMessage({ type: 'plugin:minimal:release', pluginId }, '*');
  } else {
    getJsPluginCtx()?.releaseMinimalMode();
  }
}

export function subscribeMode(cb: (mode: Mode) => void): () => void {
  if (isInsideIframe()) {
    modeCallbacks.add(cb);
    // Replay the last known mode immediately so late subscribers don't miss it.
    if (lastMode !== null) cb(lastMode);
    return () => modeCallbacks.delete(cb);
  } else {
    return getJsPluginCtx()?.subscribeMode(cb) ?? (() => undefined);
  }
}

export function subscribeTabActive(cb: (pluginId: string) => void): () => void {
  if (!isInsideIframe()) return () => undefined;
  tabActiveCallbacks.add(cb);
  if (lastTabActivePluginId !== null) cb(lastTabActivePluginId);
  return () => tabActiveCallbacks.delete(cb);
}

export function subscribeTabInactive(cb: (pluginId: string) => void): () => void {
  if (!isInsideIframe()) return () => undefined;
  tabInactiveCallbacks.add(cb);
  if (lastTabInactivePluginId !== null) cb(lastTabInactivePluginId);
  return () => tabInactiveCallbacks.delete(cb);
}

export function subscribeTabClosed(cb: (closedPluginId: string) => void): () => void {
  if (!isInsideIframe()) return () => undefined;
  tabClosedCallbacks.add(cb);
  return () => tabClosedCallbacks.delete(cb);
}
