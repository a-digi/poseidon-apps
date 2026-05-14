import { useState } from 'react';
import type { Highlights } from '../api';
import { usePlayYouTube } from '../lib/usePlayYouTube';
import { openVideoSearch } from '../lib/openVideoSearch';
import { VideoModal } from './VideoModal';
import { HighlightTile } from './HighlightTile';
import { SongListView, type SongListEntry } from './SongListView';

interface HighlightsProps {
  data: Highlights | null;
  onPlayed?: () => void;
}

type Tab = 'artists' | 'songs';

export function Highlights({ data, onPlayed }: HighlightsProps) {
  const { play, close, content, loadingKey } = usePlayYouTube({ onPlayed });
  const [activeTab, setActiveTab] = useState<Tab>('artists');

  const artistTiles = (data?.artists ?? []).slice(0, 12).map((a, i) => (
    <HighlightTile
      key={a.name}
      primaryLabel={a.name}
      position={i + 1}
      count={a.count}
      artworkUrl={a.latestArtworkUrl}
      fill
    />
  ));

  const songEntries: SongListEntry[] = (data?.songs ?? []).map((s, i) => ({
    id: `highlight-song-${s.title}`,
    position: i + 1,
    title: s.title,
    artists: s.artists,
    artworkUrl: s.artworkUrl,
    chartName: s.chartName,
    country: s.country,
  }));

  const hasArtists = artistTiles.length >= 2;
  const hasSongs = songEntries.length >= 2;
  const showTabs = hasArtists && hasSongs;

  if (!hasArtists && !hasSongs) return null;

  return (
    <section className="mx-4 mt-3 mb-5">
      <h2 className="text-lg font-semibold text-black mb-4">Highlights</h2>

      {showTabs && (
        <div className="flex gap-1 mb-4 border-b border-slate-200">
          {(['artists', 'songs'] as Tab[]).map(tab => (
            <button
              key={tab}
              type="button"
              onClick={() => setActiveTab(tab)}
              className={`px-4 py-2 text-sm font-semibold -mb-px transition-colors ${activeTab === tab
                ? 'text-black border-b-2 border-black'
                : 'text-slate-400 hover:text-slate-600'
                }`}
            >
              {tab === 'artists' ? 'Top Artists' : 'Top Songs'}
            </button>
          ))}
        </div>
      )}

      {(!showTabs || activeTab === 'artists') && hasArtists && (
        <div className="grid grid-cols-4 gap-3">{artistTiles}</div>
      )}

      {(!showTabs || activeTab === 'songs') && hasSongs && (
        <SongListView
          entries={songEntries}
          paletteId="highlights"
          loadingId={loadingKey}
          onPlay={entry => play({
            key: entry.id,
            title: entry.title,
            artists: entry.artists,
            artworkUrl: entry.artworkUrl,
            chartName: entry.chartName,
            position: entry.position,
            country: entry.country,
          })}
          onSearch={openVideoSearch}
        />
      )}

      {content && <VideoModal {...content} onClose={close} />}
    </section>
  );
}
