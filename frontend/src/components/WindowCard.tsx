import React, { useEffect, useRef, useState } from 'react';
import type { Window, PlaylistItem, SyncSession, Media } from '../types';
import { usePlaybackEngine } from '../hooks/usePlaybackEngine';
import { wsService } from '../services/websocket';
import { PlaylistEditor } from './PlaylistEditor';
import styles from './WindowCard.module.css';

interface Props {
  windowData: Window;
  playlist: PlaylistItem[];
  activeSync: SyncSession | null;
  mediaLibrary: Media[];
}

export const WindowCard: React.FC<Props> = ({ windowData, playlist, activeSync, mediaLibrary }) => {
  const { currentMedia, isPlayingSync, mediaStartTimeMs } = usePlaybackEngine(playlist, activeSync);
  const videoRef = useRef<HTMLVideoElement>(null);
  const [isEditing, setIsEditing] = useState(false);

  useEffect(() => {
    if (currentMedia?.type !== 'VIDEO' || !videoRef.current) return;

    const video = videoRef.current;
    // Force play if video is selected
    video.play().catch(console.error);

    const syncVideo = () => {
      const serverNow = wsService.getServerTime();
      const expectedOffsetSeconds = (serverNow - mediaStartTimeMs) / 1000;
      
      if (Math.abs(video.currentTime - expectedOffsetSeconds) > 0.5) {
        // Only seek if we're not beyond the video's actual duration, unless duration isn't loaded yet
        if (expectedOffsetSeconds < video.duration || isNaN(video.duration)) {
          video.currentTime = expectedOffsetSeconds;
        }
      }
    };

    syncVideo();
    const intervalId = setInterval(syncVideo, 1000);

    return () => clearInterval(intervalId);
  }, [currentMedia, mediaStartTimeMs]);

  const renderMedia = (media: Media | null) => {
    if (!media) {
      return <div className={styles.blank}>NO MEDIA</div>;
    }
    
    switch (media.type) {
      case 'IMAGE':
        return <img src={media.url} alt={media.name} className={styles.mediaElement} />;
      case 'VIDEO':
        return (
          <video 
            ref={videoRef}
            src={media.url} 
            className={styles.mediaElement} 
            loop 
            muted 
            playsInline
            autoPlay 
          />
        );
      case 'BLANK':
      default:
        return <div className={styles.blank}>BLANK</div>;
    }
  };

  return (
    <div className={`${styles.card} ${isPlayingSync ? styles.syncActive : ''}`}>
      <div className={styles.header}>
        <h3>{windowData.name}</h3>
        <div className={styles.headerRight}>
          {isPlayingSync && <span className={styles.syncBadge}>SYNC ACTIVE</span>}
          <button onClick={() => setIsEditing(!isEditing)} className={styles.editBtn}>
            {isEditing ? 'Close Edit' : 'Edit Playlist'}
          </button>
        </div>
      </div>
      
      {isEditing ? (
        <PlaylistEditor 
          windowId={windowData.id} 
          initialPlaylist={playlist} 
          mediaLibrary={mediaLibrary} 
          onClose={() => setIsEditing(false)} 
        />
      ) : (
        <>
          <div className={styles.mediaContainer}>
            {renderMedia(currentMedia)}
          </div>

          <div className={styles.footer}>
            <div className={styles.status}>
              {currentMedia ? currentMedia.name : 'Waiting for media...'}
            </div>
          </div>
        </>
      )}
    </div>
  );
};
