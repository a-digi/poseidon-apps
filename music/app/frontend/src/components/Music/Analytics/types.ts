export interface AnalyticsSummary {
  totalPlays: number;
  uniqueItemsPlayed: number;
  uniquePlaylistsPlayed: number;
  firstPlayAt?: number;
  lastPlayAt?: number;
}

export interface TopItem {
  itemId: string;
  title: string;
  artist: string;
  album: string;
  plays: number;
}

export interface TopPlaylist {
  playlistId: string;
  name: string;
  plays: number;
}

export interface TopArtist {
  artist: string;
  plays: number;
}

export interface TopAlbum {
  artist: string;
  album: string;
  plays: number;
}

export interface TopGenre {
  genre: string;
  plays: number;
}

export interface Bucket {
  key: string;
  plays: number;
}

export interface AnalyticsOverview {
  summary: AnalyticsSummary;
  topItems: TopItem[];
  topPlaylists: TopPlaylist[];
  topArtists: TopArtist[];
  topAlbums: TopAlbum[];
  topGenres: TopGenre[];
  byHour: Bucket[];
  byWeekday: Bucket[];
  byMonth: Bucket[];
  byYear: Bucket[];
}
