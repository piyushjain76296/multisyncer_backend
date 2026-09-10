import { useEffect, useState } from 'react';
import type { Window, Media, PlaylistItem, SyncSession } from './types';
import { api } from './services/api';
import { wsService } from './services/websocket';
import { WindowCard } from './components/WindowCard';
import { SyncPanel } from './components/SyncPanel';
import { MediaManager } from './components/MediaManager';
import { WindowManager } from './components/WindowManager';
import styles from './App.module.css';

function App() {
  const [windows, setWindows] = useState<Window[]>([]);
  const [mediaLibrary, setMediaLibrary] = useState<Media[]>([]);
  const [playlists, setPlaylists] = useState<Record<string, PlaylistItem[]>>({});
  const [activeSync, setActiveSync] = useState<SyncSession | null>(null);
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    // Initial fetch
    const loadData = async () => {
      try {
        const [wins, media, sync] = await Promise.all([
          api.getWindows(),
          api.getMedia(),
          api.getActiveSync()
        ]);
        setWindows(wins);
        setMediaLibrary(media);
        setActiveSync(sync);

        const lists: Record<string, PlaylistItem[]> = {};
        for (const w of wins) {
          lists[w.id] = await api.getPlaylist(w.id);
        }
        setPlaylists(lists);
      } catch (e) {
        console.error("Failed to load initial data", e);
      }
    };
    loadData();

    // WS Connection
    wsService.connect();
    
    // Check connection periodically since WS doesn't natively expose state easily to React
    const connInterval = setInterval(() => {
      // @ts-ignore
      setConnected(wsService.ws?.readyState === 1);
    }, 1000);

    const unsub = wsService.onMessage((type, payload) => {
      if (type === 'SYNC_SCHEDULED') {
        setActiveSync(payload);
      } else if (type === 'PLAYLIST_UPDATED') {
        setPlaylists(prev => ({
          ...prev,
          [payload.window_id]: payload.items
        }));
      } else if (type === 'WINDOW_CREATED') {
        setWindows(prev => [...prev, payload]);
        setPlaylists(prev => ({ ...prev, [payload.id]: [] }));
      } else if (type === 'MEDIA_CREATED') {
        setMediaLibrary(prev => [...prev, payload]);
      } else if (type === 'DISCONNECTED') {
        setConnected(false);
      }
    });

    return () => {
      unsub();
      clearInterval(connInterval);
    };
  }, []);

  return (
    <div className={styles.appContainer}>
      <header className={styles.topbar}>
        <h1>Multisyncer Control Panel</h1>
        <div className={styles.connectionStatus}>
          <span className={`${styles.dot} ${connected ? styles.green : styles.red}`}></span>
          {connected ? 'Connected' : 'Disconnected'}
        </div>
      </header>
      
      <main className={styles.mainContent}>
        <div className={styles.adminSection}>
          <SyncPanel mediaLibrary={mediaLibrary} />
          <MediaManager mediaLibrary={mediaLibrary} />
          <WindowManager />
        </div>
        
        <div className={styles.windowGrid}>
          {windows.map(w => (
            <WindowCard 
              key={w.id} 
              windowData={w} 
              playlist={playlists[w.id] || []} 
              activeSync={activeSync} 
              mediaLibrary={mediaLibrary}
            />
          ))}
        </div>
      </main>
    </div>
  );
}

export default App;
