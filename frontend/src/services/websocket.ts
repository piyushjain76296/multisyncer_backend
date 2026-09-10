type WSHandler = (type: string, payload: any) => void;

export class WSService {
  private ws: WebSocket | null = null;
  private handlers: Set<WSHandler> = new Set();
  private reconnectTimer: number | null = null;
  private serverOffsetMs: number = 0;
  private url: string;

  constructor(url: string = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws') {
    this.url = url;
  }

  connect() {
    if (this.ws) return;
    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => {
      console.log('WS Connected');
      if (this.reconnectTimer) {
        window.clearTimeout(this.reconnectTimer);
        this.reconnectTimer = null;
      }
      this.syncTime();
    };

    this.ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        if (msg.type === 'TIME_SYNC') {
          const t1 = msg.payload.t0; // client sent time
          const t2 = new Date(msg.timestamp).getTime(); // server received time
          const t3 = Date.now(); // client receive time
          // Estimate server offset:
          const rtt = t3 - t1;
          const serverNow = t2 + rtt / 2;
          this.serverOffsetMs = serverNow - t3;
        } else {
          this.handlers.forEach(h => h(msg.type, msg.payload));
        }
      } catch (e) {
        console.error('Failed to parse WS message', e);
      }
    };

    this.ws.onclose = () => {
      console.log('WS Disconnected, reconnecting in 2s...');
      this.ws = null;
      this.reconnectTimer = window.setTimeout(() => this.connect(), 2000);
      this.handlers.forEach(h => h('DISCONNECTED', null));
    };
  }

  syncTime() {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: 'TIME_SYNC', payload: { t0: Date.now() } }));
    }
  }

  getServerTime(): number {
    return Date.now() + this.serverOffsetMs;
  }

  onMessage(handler: WSHandler) {
    this.handlers.add(handler);
    return () => this.handlers.delete(handler);
  }
}

export const wsService = new WSService();
