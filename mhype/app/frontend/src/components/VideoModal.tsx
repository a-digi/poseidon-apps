import { useEffect } from 'react';
import { createPortal } from 'react-dom';
import { X } from 'lucide-react';

export interface VideoModalContent {
  videoId: string;
  artworkUrl?: string;
  title: string;
  artists: string[];
  chartName?: string;
  position?: number;
  country?: string;
}

interface VideoModalProps extends VideoModalContent {
  onClose: () => void;
}

export function VideoModal({ videoId, artworkUrl, title, artists, chartName, position, country, onClose }: VideoModalProps) {
  useEffect(() => {
    function handleKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose();
    }
    document.addEventListener('keydown', handleKey);
    document.body.style.overflow = 'hidden';
    return () => {
      document.removeEventListener('keydown', handleKey);
      document.body.style.overflow = '';
    };
  }, [onClose]);

  return createPortal(
    <div
      role="dialog"
      aria-modal="true"
      aria-label="Video player"
      className="fixed inset-0 z-50 flex flex-col"
    >
      <div
        className="absolute inset-0 bg-white/80 backdrop-blur-xl"
        onClick={onClose}
        aria-hidden="true"
      />

      <div className="relative z-10 flex items-stretch border-b border-slate-200">
        <div className="w-16 shrink-0 bg-gradient-to-br from-slate-200 to-slate-300">
          {artworkUrl && (
            <img src={artworkUrl} alt="" className="w-full h-full object-cover" />
          )}
        </div>

        {position != null && (
          <div className="flex items-center justify-center px-5 py-4 shrink-0 w-32">
            <span className="text-7xl font-black text-slate-900 leading-none">#{position}</span>
          </div>
        )}

        <div className="flex flex-col justify-center px-6 py-4 flex-1 min-w-0">
          <p className="text-3xl font-semibold text-slate-900 truncate leading-tight">{title}</p>
          {artists.length > 0 && (
            <p className="text-xl text-slate-500 truncate mt-1">{artists[0]}</p>
          )}
        </div>

        {(chartName || country) && (
          <div className="flex flex-col justify-center items-end px-4 py-4 shrink-0 max-w-[180px]">
            {chartName && <span className="text-sm font-medium text-black truncate text-right">{chartName}</span>}
            {country && <span className="text-xs font-medium tracking-widest uppercase text-slate-400 mt-0.5">{country}</span>}
          </div>
        )}

        <button
          type="button"
          onClick={onClose}
          aria-label="Close video"
          className="shrink-0 self-start p-3 text-slate-400 hover:text-slate-700 transition-colors"
        >
          <X className="w-5 h-5" />
        </button>
      </div>

      <div className="relative z-10 flex-1 flex items-center justify-center p-8">
        <div className="w-full max-w-3xl">
          <div className="aspect-video rounded-xl overflow-hidden shadow-2xl">
            <iframe
              src={`https://www.youtube.com/embed/${videoId}?autoplay=1`}
              title={title}
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
              allowFullScreen
              className="w-full h-full border-0"
            />
          </div>
        </div>
      </div>
    </div>,
    document.body,
  );
}
