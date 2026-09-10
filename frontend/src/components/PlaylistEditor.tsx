import React, { useState } from 'react';
import type { PlaylistItem, Media } from '../types';
import { api } from '../services/api';
import styles from './PlaylistEditor.module.css';

interface Props {
  windowId: string;
  initialPlaylist: PlaylistItem[];
  mediaLibrary: Media[];
  onClose: () => void;
}

export const PlaylistEditor: React.FC<Props> = ({ windowId, initialPlaylist, mediaLibrary, onClose }) => {
  const [items, setItems] = useState<PlaylistItem[]>([...initialPlaylist]);
  const [isSaving, setIsSaving] = useState(false);
  const [selectedMediaId, setSelectedMediaId] = useState('');
  const [duration, setDuration] = useState('10');

  const handleAdd = () => {
    if (!selectedMediaId) return;
    const media = mediaLibrary.find(m => m.id === selectedMediaId);
    if (!media) return;

    const newItem: PlaylistItem = {
      id: crypto.randomUUID(), // Valid UUID for backend parsing
      window_id: windowId,
      media_id: media.id,
      position: (items || []).length,
      duration_seconds: parseInt(duration, 10),
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      media: media
    };
    
    setItems([...items, newItem]);
  };

  const handleRemove = (index: number) => {
    const newItems = [...items];
    newItems.splice(index, 1);
    setItems(newItems);
  };

  const handleSave = async () => {
    setIsSaving(true);
    try {
      const toSave = (items || []).map((item, idx) => ({ ...item, position: idx }));
      await api.updatePlaylist(windowId, toSave);
      onClose();
    } catch (err) {
      console.error(err);
      alert('Failed to update playlist');
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className={styles.editor}>
      <h4>Edit Playlist</h4>
      
      <div className={styles.itemList}>
        {(items || []).map((item, idx) => (
          <div key={idx} className={styles.item}>
            <span className={styles.itemPosition}>{idx + 1}</span>
            <span className={styles.itemName}>{item.media?.name || item.media_id}</span>
            <span className={styles.itemDuration}>{item.duration_seconds}s</span>
            <button onClick={() => handleRemove(idx)} className={styles.removeBtn}>✕</button>
          </div>
        ))}
        {items.length === 0 && <div className={styles.empty}>Playlist is empty</div>}
      </div>

      <div className={styles.addSection}>
        <select value={selectedMediaId} onChange={e => setSelectedMediaId(e.target.value)} className={styles.select}>
          <option value="">-- Select Media --</option>
          {(mediaLibrary || []).map(m => (
            <option key={m.id} value={m.id}>{m.name} ({m.type})</option>
          ))}
        </select>
        <input 
          type="number" 
          value={duration} 
          onChange={e => setDuration(e.target.value)} 
          min="1" 
          title="Duration in seconds"
          className={styles.durationInput}
        />
        <button onClick={handleAdd} disabled={!selectedMediaId} className={styles.addBtn}>Add</button>
      </div>

      <div className={styles.actions}>
        <button onClick={onClose} className={styles.cancelBtn}>Cancel</button>
        <button onClick={handleSave} disabled={isSaving} className={styles.saveBtn}>
          {isSaving ? 'Saving...' : 'Save Playlist'}
        </button>
      </div>
    </div>
  );
};
