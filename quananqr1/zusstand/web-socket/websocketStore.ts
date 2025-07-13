import { create } from "zustand";
import { WebSocketService } from "./websoket-service";
import envConfig from "@/config";
import { WebSocketMessage } from "@/schemaValidations/interface/type_websocker";

interface WebSocketState {
  socket: WebSocketService | null;
  isConnected: boolean;
  wsToken: string | null;
  wsTokenExpiry: string | null;
  connect: (params: {
    userId: string;
    isGuest: boolean;
    userToken: string;
    tableToken: string;
    role: string;
    email: string;
  }) => Promise<void>;
  disconnect: () => void;
  sendMessage: (message: WebSocketMessage) => void;
  addMessageHandler: (
    handler: (message: WebSocketMessage) => void
  ) => () => void;
  messageHandlers: Array<(message: WebSocketMessage) => void>;
  fetchWsToken: (params: {
    userId: number;
    email: string;
    role: string;
  }) => Promise<WsAuthResponse>;
}

interface WsAuthResponse {
  token: string;
  expiresAt: string;
  role: string;
  userId: number;
  email: string;
}

export const useWebSocketStore = create<WebSocketState>((set, get) => ({
  socket: null,
  isConnected: false,
  messageHandlers: [],
  wsToken: null,
  wsTokenExpiry: null,

  fetchWsToken: async ({ userId, email, role }) => {
    const serverEndpoint = envConfig.NEXT_PUBLIC_API_ENDPOINT;
    // console.log("quananqr1/zusstand/web-socket/websocketStore.ts fetchWsToken");
    try {
      const response = await fetch(`${serverEndpoint}${envConfig.wsAuth}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          userId,
          email,
          role
        })
      });

      if (!response.ok) {
        throw new Error("Failed to fetch WS token");
      }

      const data: WsAuthResponse = await response.json();

      set({
        wsToken: data.token,
        wsTokenExpiry: data.expiresAt
      });

      return data;
    } catch (error) {
      console.error("Error fetching WS token:", error);
      throw error;
    }
  },

  connect: async ({ userId, isGuest, userToken, tableToken, role, email }) => {
    console.log("quananqr1/zusstand/web-socket/websocketStore.ts");

    // Check if token is expired and fetch new one if needed
    const currentTime = new Date();
    const wsTokenExpiry = get().wsTokenExpiry;
    const tokenExpiry = wsTokenExpiry ? new Date(wsTokenExpiry) : null;

    const isTokenExpired =
      !get().wsToken ||
      !wsTokenExpiry ||
      currentTime >= new Date(wsTokenExpiry);

    if (isTokenExpired) {
      try {
        await get().fetchWsToken({
          userId: parseInt(userId),
          email: userToken,
          role
        });
      } catch (error) {
        console.error("Failed to get WS token before connection:", error);
        return;
      }
    }

    console.log(
      "quananqr1/zusstand/web-socket/websocketStore.ts 121212121 userId role usertoek tabletoken ",
      userId,
      role,
      userToken,
      tableToken
    );

    const socket = new WebSocketService(userId, role, userToken, tableToken, email);

    socket.onMessage((message: WebSocketMessage) => {
      const handlers = get().messageHandlers;
      handlers.forEach((handler) => handler(message));
    });

    socket.onConnect(() => set({ isConnected: true }));
    socket.onDisconnect(() => set({ isConnected: false }));

    socket.onMessage((message: WebSocketMessage) => {
      const handlers = get().messageHandlers;
      handlers.forEach((handler) => handler(message));
    });
    // socket.onMessage((message: WebSocketMessage) => {
    //   console.log(
    //     "quananqr1/zusstand/web-socket/websocketStore.ts message",
    //     message
    //   );
    // });
    set({ socket });
  },

  disconnect: () => {
    const { socket } = get();
    if (socket) {
      socket.disconnect();
      set({
        socket: null,
        isConnected: false,
        wsToken: null,
        wsTokenExpiry: null
      });
    }
  },

  sendMessage: (message: WebSocketMessage) => {
    const { socket } = get();
    if (socket) {
      socket.sendMessage(message);
    }
  },

  addMessageHandler: (handler) => {
    set((state) => ({
      messageHandlers: [...state.messageHandlers, handler]
    }));

    return () => {
      set((state) => ({
        messageHandlers: state.messageHandlers.filter((h) => h !== handler)
      }));
    };
  }
}));
