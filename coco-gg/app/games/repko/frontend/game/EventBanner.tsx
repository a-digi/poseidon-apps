import { useEffect, useRef, useState } from 'react';
import type { GameEvent } from '../types';
import { eventToText } from './eventText';

interface Toast {
  id: string;
  text: string;
  expiresAt: number;
}

interface EventBannerProps {
  events: GameEvent[];
  myPlayerId: string | null;
}

const TOAST_DURATION_MS = 3000;
const MAX_VISIBLE = 3;

export function EventBanner({ events, myPlayerId }: EventBannerProps) {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const lastIndexRef = useRef(0);

  useEffect(() => {
    if (events.length <= lastIndexRef.current) {
      lastIndexRef.current = events.length;
      return;
    }
    const newToasts: Toast[] = [];
    for (let i = lastIndexRef.current; i < events.length; i++) {
      const ev = events[i];
      if (ev.actorId === myPlayerId) continue;
      newToasts.push({
        id: `${i}-${Math.random().toString(36).slice(2, 7)}`,
        text: eventToText(ev),
        expiresAt: Date.now() + TOAST_DURATION_MS,
      });
    }
    lastIndexRef.current = events.length;
    if (newToasts.length === 0) return;
    setToasts((prev) => [...prev, ...newToasts].slice(-MAX_VISIBLE));
  }, [events, myPlayerId]);

  useEffect(() => {
    if (toasts.length === 0) return;
    const id = window.setInterval(() => {
      const now = Date.now();
      setToasts((prev) => prev.filter((t) => t.expiresAt > now));
    }, 250);
    return () => window.clearInterval(id);
  }, [toasts.length]);

  if (toasts.length === 0) return null;

  return (
    <div className="pointer-events-none fixed left-1/2 top-2 z-50 flex -translate-x-1/2 flex-col items-center gap-1">
      {toasts.map((t) => (
        <div
          key={t.id}
          className="rounded-full bg-slate-900/90 px-3 py-1.5 text-xs font-semibold text-white shadow-lg"
        >
          {t.text}
        </div>
      ))}
    </div>
  );
}
