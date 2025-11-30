import { create } from 'zustand';

interface SignalingState {
  socket: WebSocket | null;
  isConnected: boolean;
  messages: string[]; // Log temporal para ver si funciona
  
  connect: () => void;
  sendMessage: (msg: any) => void;
  disconnect: () => void;
}

export const useSignalingStore = create<SignalingState>((set, get) => ({
  socket: null,
  isConnected: false,
  messages: [],

  connect: () => {
    const url = process.env.NEXT_PUBLIC_WS_URL;
    if (!url) {
        console.error("WS URL no definida");
        return;
    }

    if (get().socket) return; // Ya conectado

    const ws = new WebSocket(url);

    ws.onopen = () => {
      console.log('Conectado al Servidor de Señalizacion');
      set({ isConnected: true });
    };

    ws.onmessage = (event) => {
      console.log('Mensaje recibido:', event.data);
      // Guardamos el mensaje en el estado para mostrarlo en UI (debugging)
      set((state) => ({ messages: [...state.messages, event.data] }));
    };

    ws.onclose = () => {
      console.log('Desconectado');
      set({ isConnected: false, socket: null });
    };

    set({ socket: ws });
  },

  sendMessage: (msg: any) => {
    const { socket } = get();
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify(msg));
    } else {
      console.warn('No se puede enviar: Socket desconectado');
    }
  },

  disconnect: () => {
    const { socket } = get();
    if (socket) {
      socket.close();
      set({ socket: null, isConnected: false });
    }
  },
}));