export interface Playlist {
  id: string;
  name: string;
  createdAt: number;
  items?: DigitalItem[];
}

export interface DigitalItem {
  id: string;
  title: string;
  url: string;
  playlistId?: string;
  artist?: string;
  album?: string;
  genre?: string;
  year?: number;
  track?: number;
  length?: number;
  picture?: string;
  mimeType?: string;
}
