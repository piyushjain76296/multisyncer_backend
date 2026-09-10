import { useState, useEffect } from 'react';
import type { PlaylistItem, SyncSession, PlaybackState } from '../types';
import { wsService } from '../services/websocket';

const CYCLE_DURATION_MS = 5 * 60 * 60 * 1000; // 5 hours

export function usePlaybackEngine(
  playlist: PlaylistItem[],
  activeSync: SyncSession | null
): PlaybackState {
  const [state, setState] = useState<PlaybackState>({
    currentMedia: null,
    isPlayingSync: false,
    timeRemainingMs: 0,
    mediaStartTimeMs: 0,
  });

  useEffect(() => {
    let animationFrameId: number;

    const calculateState = () => {
      const serverNow = wsService.getServerTime();

      // Check Sync Session
      if (activeSync) {
        const startAt = new Date(activeSync.start_at).getTime();
        const endAt = new Date(activeSync.end_at).getTime();
        
        if (serverNow >= startAt && serverNow < endAt) {
          setState({
            currentMedia: activeSync.media || null,
            isPlayingSync: true,
            timeRemainingMs: endAt - serverNow,
            mediaStartTimeMs: startAt,
          });
          animationFrameId = requestAnimationFrame(calculateState);
          return;
        }
      }

      // Normal Playback
      if (!playlist || playlist.length === 0) {
        setState({ currentMedia: null, isPlayingSync: false, timeRemainingMs: 0, mediaStartTimeMs: 0 });
        animationFrameId = requestAnimationFrame(calculateState);
        return;
      }

      const totalPlaylistMs = playlist.reduce((sum, item) => sum + (item.duration_seconds * 1000), 0);
      if (totalPlaylistMs === 0) {
        setState({ currentMedia: null, isPlayingSync: false, timeRemainingMs: 0, mediaStartTimeMs: 0 });
        animationFrameId = requestAnimationFrame(calculateState);
        return;
      }

      const timeInCycle = serverNow % CYCLE_DURATION_MS;
      const timeInPlaylist = timeInCycle % totalPlaylistMs;

      let accumulatedMs = 0;
      let activeItem: PlaylistItem | null = null;
      let timeRemainingMs = 0;
      let mediaStartTimeMs = 0;

      for (const item of playlist) {
        const itemDurationMs = item.duration_seconds * 1000;
        if (timeInPlaylist >= accumulatedMs && timeInPlaylist < accumulatedMs + itemDurationMs) {
          activeItem = item;
          timeRemainingMs = (accumulatedMs + itemDurationMs) - timeInPlaylist;
          mediaStartTimeMs = serverNow - timeInPlaylist + accumulatedMs;
          break;
        }
        accumulatedMs += itemDurationMs;
      }

      // Edge case if something goes wrong with float math
      if (!activeItem) {
        activeItem = playlist[0];
        timeRemainingMs = activeItem.duration_seconds * 1000;
        mediaStartTimeMs = serverNow;
      }

      setState(prev => {
        if (
          prev.currentMedia?.id === activeItem?.media?.id &&
          prev.isPlayingSync === false
        ) {
          return prev; // Do not trigger re-render if the media hasn't changed
        }
        return {
          currentMedia: activeItem?.media || null,
          isPlayingSync: false,
          timeRemainingMs,
          mediaStartTimeMs,
        };
      });

      animationFrameId = requestAnimationFrame(calculateState);
    };

    animationFrameId = requestAnimationFrame(calculateState);

    return () => cancelAnimationFrame(animationFrameId);
  }, [playlist, activeSync]);

  return state;
}
