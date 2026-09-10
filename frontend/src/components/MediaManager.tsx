import React, { useState } from 'react';
import type { Media } from '../types';
import { api } from '../services/api';
import styles from './MediaManager.module.css';

interface Props {
  mediaLibrary: Media[];
}

export const MediaManager: React.FC<Props> = ({ mediaLibrary }) => {
  const [name, setName] = useState('');
  const [type, setType] = useState('VIDEO');
  const [url, setUrl] = useState('');
  const [duration, setDuration] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name || !url || !duration) return;
    
    setIsSubmitting(true);
    try {
      await api.createMedia(name, type, url, parseInt(duration, 10));
      setName('');
      setUrl('');
      setDuration('');
    } catch (err) {
      console.error(err);
      alert('Failed to create media');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className={styles.managerPanel}>
      <h2>Media Manager</h2>
      
      <div className={styles.mediaList}>
        {mediaLibrary.map(m => (
          <div key={m.id} className={styles.mediaItem}>
            <span className={styles.mediaName}>{m.name}</span>
            <span className={styles.mediaType}>{m.type}</span>
            <span className={styles.mediaDuration}>{m.duration_seconds}s</span>
          </div>
        ))}
        {mediaLibrary.length === 0 && <div className={styles.empty}>No media found.</div>}
      </div>
      
      <form onSubmit={handleSubmit} className={styles.addForm}>
        <div className={styles.formRow}>
          <div className={styles.field}>
            <label>Name</label>
            <input type="text" value={name} onChange={e => setName(e.target.value)} required />
          </div>
          <div className={styles.field}>
            <label>Type</label>
            <select value={type} onChange={e => setType(e.target.value)}>
              <option value="VIDEO">Video</option>
              <option value="IMAGE">Image</option>
              <option value="BLANK">Blank</option>
            </select>
          </div>
        </div>
        <div className={styles.formRow}>
          <div className={styles.field} style={{flex: 2}}>
            <label>URL</label>
            <input type="url" value={url} onChange={e => setUrl(e.target.value)} required />
          </div>
          <div className={styles.field} style={{flex: 1}}>
            <label>Duration (s)</label>
            <input type="number" value={duration} onChange={e => setDuration(e.target.value)} required min="1" />
          </div>
        </div>
        <button type="submit" disabled={isSubmitting} className={styles.addBtn}>
          {isSubmitting ? 'Adding...' : 'Add Media'}
        </button>
      </form>
    </div>
  );
};
