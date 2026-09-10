export interface Window {
  id: string;
  name: string;
  created_at: string;
  updated_at: string;
}

export type MediaType = 'IMAGE' | 'VIDEO' | 'BLANK';

export interface Media {
  id: string;
  name: string;
  type: MediaType;
  url: string;
  duration_seconds: number;
  created_at: string;
  updated_at: string;
}

export interface PlaylistItem {
  id: string;
  window_id: string;
  media_id: string;
  position: number;
  duration_seconds: number;
  created_at: string;
  updated_at: string;
  media?: Media;
}

export type SyncStatus = 'SCHEDULED' | 'ACTIVE' | 'COMPLETED';

export interface SyncSession {
  id: string;
  media_id: string;
  start_at: string;
  end_at: string;
  status: SyncStatus;
  created_at: string;
  media?: Media;
}

export interface PlaybackState {
  currentMedia: Media | null;
  isPlayingSync: boolean;
  timeRemainingMs: number;
  mediaStartTimeMs: number;
}
