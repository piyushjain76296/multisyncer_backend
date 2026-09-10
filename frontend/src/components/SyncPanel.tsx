import React, { useState } from 'react';
import type { Media } from '../types';
import { api } from '../services/api';
import styles from './SyncPanel.module.css';

interface Props {
  mediaLibrary: Media[];
}

export const SyncPanel: React.FC<Props> = ({ mediaLibrary }) => {
  const [selectedMediaId, setSelectedMediaId] = useState<string>('');
  const [duration, setDuration] = useState<number>(30);
  const [isSyncing, setIsSyncing] = useState(false);

  const handleSync = async () => {
    if (!selectedMediaId || duration <= 0) return;
    setIsSyncing(true);
    try {
      await api.createSyncSession(selectedMediaId, duration);
    } catch (e) {
      console.error('Failed to schedule sync', e);
    } finally {
      setIsSyncing(false);
    }
  };

  return (
    <div className={styles.syncPanel}>
      <h2>SYNC PLAYBACK</h2>
      <div className={styles.controls}>
        <div className={styles.field}>
          <label>Select media:</label>
          <select 
            value={selectedMediaId} 
            onChange={(e) => setSelectedMediaId(e.target.value)}
          >
            <option value="">-- Select Media --</option>
            {mediaLibrary.map(m => (
              <option key={m.id} value={m.id}>{m.name} ({m.type})</option>
            ))}
          </select>
        </div>
        
        <div className={styles.field}>
          <label>Duration (sec):</label>
          <input 
            type="number" 
            value={duration} 
            onChange={(e) => setDuration(Number(e.target.value))}
            min="1"
          />
        </div>

        <button 
          className={styles.syncBtn} 
          onClick={handleSync}
          disabled={!selectedMediaId || isSyncing}
        >
          {isSyncing ? 'SCHEDULING...' : 'SYNC ALL WINDOWS'}
        </button>
      </div>
    </div>
  );
};
