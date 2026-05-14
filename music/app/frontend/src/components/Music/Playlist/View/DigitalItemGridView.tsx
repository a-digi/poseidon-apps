import React from 'react';
import type { DigitalItem } from '../Playlist';

interface DigitalItemGridViewProps {
  items: DigitalItem[];
  playlistId: string;
  t: any;
  onDelete?: (playlistId: string, itemId: string) => void;
  showDeleteButton?: boolean;
  onItemClick?: (item: DigitalItem) => void;
  selectedItemId?: string;
}

const DigitalItemGridView: React.FC<DigitalItemGridViewProps> = ({
  items,
  playlistId,
  t,
  onDelete,
  showDeleteButton = false,
  onItemClick,
  selectedItemId,
}) => (
  <div className="grid gap-2 mt-2 grid-cols-[repeat(auto-fit,200px)] justify-left">
    {items.map((item) => {
      return (
        <div key={item.id} onClick={() => onItemClick && onItemClick(item)}>
          <div
            className={
              'rounded-xl shadow-lg overflow-hidden flex flex-col relative group border border-gray-200 hover:shadow-2xl transition-shadow duration-200 w-full cursor-pointer ' +
              (selectedItemId === item.id
                ? 'bg-blue-100'
                : 'bg-gradient-to-br from-gray-100 to-blue-100')
            }
          >
            <div className="relative w-full aspect-square bg-gray-200 flex items-center justify-center">
              {item.picture ? (
                <img src={item.picture} alt={item.title} className="w-full h-full object-cover" />
              ) : (
                <div className="w-full h-full flex items-center justify-center text-gray-400 text-4xl font-bold">
                  ?
                </div>
              )}
              {showDeleteButton && onDelete && (
                <button
                  className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity bg-red-500 text-white rounded px-2 py-0.5 text-xs shadow-lg z-10"
                  title={t('playlist.deleteMp3')}
                  onClick={(e) => {
                    e.stopPropagation();
                    onDelete(playlistId, item.id);
                  }}
                >
                  {t('playlist.delete')}
                </button>
              )}
            </div>
            <div className="flex flex-col items-center px-3 py-3 bg-white/80 backdrop-blur-sm">
              <span
                className="font-semibold text-center w-full truncate text-base text-gray-800"
                title={item.title}
              >
                {item.title}
              </span>
              {item.artist && (
                <span className="text-xs text-gray-600 mt-0.5 truncate w-full text-center">
                  {item.artist}
                </span>
              )}
              {item.album && (
                <span className="text-xs text-gray-500 mt-0.5 truncate w-full text-center">
                  {item.album}
                </span>
              )}
              {typeof item.length === 'number' && item.length > 0 && (
                <span className="text-xs text-blue-700 mt-1 font-mono">
                  {t('playlist.length')}: {Math.floor(item.length / 60)}:
                  {(item.length % 60).toString().padStart(2, '0')} min
                </span>
              )}
            </div>
          </div>
        </div>
      );
    })}
  </div>
);

export default DigitalItemGridView;

