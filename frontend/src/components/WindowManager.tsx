import React, { useState } from 'react';
import { api } from '../services/api';
import styles from './WindowManager.module.css';

export const WindowManager: React.FC = () => {
  const [name, setName] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name) return;
    
    setIsSubmitting(true);
    try {
      await api.createWindow(name);
      setName('');
    } catch (err) {
      console.error(err);
      alert('Failed to create window');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className={styles.managerPanel}>
      <h2>Add New Window</h2>
      <form onSubmit={handleSubmit} className={styles.addForm}>
        <div className={styles.field}>
          <label>Window Name</label>
          <input type="text" value={name} onChange={e => setName(e.target.value)} required />
        </div>
        <button type="submit" disabled={isSubmitting} className={styles.addBtn}>
          {isSubmitting ? 'Adding...' : 'Add Window'}
        </button>
      </form>
    </div>
  );
};
