import { useState, useRef, useCallback } from 'react';
import { findYouTubeVideo, trackPlay } from '../api';
import type { VideoModalContent } from '../components/VideoModal';

export interface PlayInput {
  key: string;
  title: string;
  artists: string[];
  artworkUrl?: string;
  chartName?: string;
  crawlerId?: string;
  position?: number;
  country?: string;
}

export interface UsePlayYouTubeOptions {
  onPlayed?: () => void;
}

export interface UsePlayYouTube {
  play: (input: PlayInput) => void;
  close: () => void;
  content: VideoModalContent | null;
  loadingKey: string | null;
  error: string | null;
}

export function usePlayYouTube(options: UsePlayYouTubeOptions = {}): UsePlayYouTube {
  const [content, setContent] = useState<VideoModalContent | null>(null);
  const [loadingKey, setLoadingKey] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const reqRef = useRef(0);
  const errorTimer = useRef<number | null>(null);

  const play = useCallback((input: PlayInput) => {
    const id = ++reqRef.current;
    if (errorTimer.current !== null) {
      window.clearTimeout(errorTimer.current);
      errorTimer.current = null;
    }
    setError(null);
    setLoadingKey(input.key);
    findYouTubeVideo((input.artists[0] ?? ''), input.title)
      .then(result => {
        if (reqRef.current !== id) return;
        setContent({
          videoId: result.videoId,
          title: input.title,
          artists: input.artists,
          artworkUrl: input.artworkUrl,
          chartName: input.chartName,
          position: input.position,
          country: input.country,
        });
        trackPlay({
          title: input.title,
          artists: input.artists,
          artworkUrl: input.artworkUrl,
          chartName: input.chartName,
          crawlerId: input.crawlerId,
          position: input.position,
          country: input.country,
        }).then(() => {
          options.onPlayed?.();
        }).catch(() => undefined);
      })
      .catch(err => {
        if (reqRef.current !== id) return;
        const msg = err instanceof Error ? err.message : String(err);
        setError(msg);
        errorTimer.current = window.setTimeout(() => {
          setError(null);
          errorTimer.current = null;
        }, 5000);
      })
      .finally(() => {
        if (reqRef.current === id) setLoadingKey(null);
      });
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const close = useCallback(() => setContent(null), []);

  return { play, close, content, loadingKey, error };
}
