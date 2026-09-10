import type { Window, Media, PlaylistItem, SyncSession } from '../types';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

export const api = {
  getWindows: async (): Promise<Window[]> => {
    const res = await fetch(`${API_URL}/windows`);
    return res.json();
  },
  createWindow: async (name: string): Promise<Window> => {
    const res = await fetch(`${API_URL}/windows`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name }),
    });
    return res.json();
  },
  getPlaylist: async (windowId: string): Promise<PlaylistItem[]> => {
    const res = await fetch(`${API_URL}/windows/${windowId}/playlist`);
    return res.json();
  },
  updatePlaylist: async (windowId: string, items: PlaylistItem[]): Promise<PlaylistItem[]> => {
    const res = await fetch(`${API_URL}/windows/${windowId}/playlist`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(items),
    });
    if (!res.ok) {
      throw new Error(await res.text());
    }
    return res.json();
  },
  getMedia: async (): Promise<Media[]> => {
    const res = await fetch(`${API_URL}/media`);
    return res.json();
  },
  createMedia: async (name: string, type: string, url: string, duration_seconds: number): Promise<Media> => {
    const res = await fetch(`${API_URL}/media`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, type, url, duration_seconds }),
    });
    return res.json();
  },
  createSyncSession: async (mediaId: string, durationSeconds: number): Promise<SyncSession> => {
    const res = await fetch(`${API_URL}/sync`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ media_id: mediaId, duration_seconds: durationSeconds }),
    });
    return res.json();
  },
  getActiveSync: async (): Promise<SyncSession | null> => {
    const res = await fetch(`${API_URL}/sync/active`);
    if (res.status === 204 || res.headers.get('content-length') === '0') return null;
    const data = await res.json();
    return data && data.id ? data : null;
  }
};
