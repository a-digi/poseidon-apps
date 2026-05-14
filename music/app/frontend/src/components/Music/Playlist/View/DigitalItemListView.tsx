import React, { useState } from 'react';
import type { DigitalItem } from '../Playlist';

interface DigitalItemListViewProps {
  items: DigitalItem[];
  playlistId: string;
  onDelete?: (playlistId: string, itemId: string) => void;
  t: any;
  showDeleteButton?: boolean;
  onItemClick?: (item: DigitalItem) => void;
  selectedItemId?: string;
}

const DigitalItemListView: React.FC<DigitalItemListViewProps> = ({
  items,
  playlistId,
  onDelete,
  t,
  showDeleteButton = true,
  onItemClick,
  selectedItemId,
}) => {
  const [itemsPerRow, setItemsPerRow] = useState(1);

  // Hilfsfunktion, um das Array in Chunks zu teilen
  function chunkArray<T>(arr: T[], size: number): T[][] {
    const result: T[][] = [];
    for (let i = 0; i < arr.length; i += size) {
      result.push(arr.slice(i, i + size));
    }
    return result;
  }

  const rows = chunkArray(items, itemsPerRow);

  return (
    <div>
      {/* Überschrift für Items pro Zeile */}
      <div className="flex items-center gap-2 mb-2">
        {[1, 2, 3].map((n) => (
          <button
            key={n}
            className={`px-2 py-1 rounded border text-xs font-semibold flex items-center justify-center ${itemsPerRow === n ? 'bg-blue-500 text-white border-blue-500' : 'bg-white text-gray-700 border-gray-300'}`}
            onClick={() => setItemsPerRow(n)}
            aria-pressed={itemsPerRow === n}
            title={`${n} ${n === 1 ? 'Item' : 'Items'} pro Zeile`}
          >
            {n === 1 && (
              <svg
                width="22"
                height="18"
                viewBox="0 0 22 18"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
              >
                <rect x="3" y="7" width="16" height="4" rx="2" fill="currentColor" />
              </svg>
            )}
            {n === 2 && (
              <svg
                width="22"
                height="18"
                viewBox="0 0 22 18"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
              >
                <rect x="3" y="4" width="7" height="4" rx="2" fill="currentColor" />
                <rect x="12" y="4" width="7" height="4" rx="2" fill="currentColor" />
              </svg>
            )}
            {n === 3 && (
              <svg
                width="22"
                height="18"
                viewBox="0 0 22 18"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
              >
                <rect x="2" y="3" width="5" height="4" rx="2" fill="currentColor" />
                <rect x="8.5" y="3" width="5" height="4" rx="2" fill="currentColor" />
                <rect x="15" y="3" width="5" height="4" rx="2" fill="currentColor" />
              </svg>
            )}
          </button>
        ))}
      </div>
      <ul className="flex flex-col gap-2">
        {rows.map((row, rowIdx) => (
          <li key={rowIdx} className="flex gap-2 list-none">
            {row.map((item) => {
              return (
                <div
                  onClick={() => onItemClick && onItemClick(item)}
                  key={item.id}
                  className={`flex flex-row items-center text-sm group gap-3 p-2 cursor-pointer rounded flex-1 ${selectedItemId === item.id ? 'bg-gray-900 text-white' : 'bg-white text-gray-900 '}`}
                  style={{ minWidth: 0 }}
                >
                  {item.picture && (
                    <img
                      src={item.picture}
                      alt={item.title}
                      className="w-20 h-20 object-cover rounded shadow border flex-shrink-0"
                    />
                  )}
                  <div className="flex flex-col flex-1 min-w-0">
                    <span className="truncate font-bold text-base" title={item.title}>
                      {item.title}
                    </span>
                    {item.artist && <span className="text-xs ">{item.artist}</span>}
                    {item.album && <span className="text-xs ">{item.album}</span>}
                    {typeof item.length === 'number' && item.length > 0 && (
                      <span className="text-xs  mt-1">
                        {t('playlist.length')}: {Math.floor(item.length / 60)}:
                        {(item.length % 60).toString().padStart(2, '0')} min
                      </span>
                    )}
                  </div>
                  {showDeleteButton && onDelete && (
                    <button
                      className="ml-2 bg-red-500 text-white rounded px-2 py-0.5 text-xs h-7 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"
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
              );
            })}
          </li>
        ))}
      </ul>
    </div>
  );
};

export default DigitalItemListView;
