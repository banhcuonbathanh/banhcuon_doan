
import { WebSocketMessage } from "@/schemaValidations/interface/type_websocker";


// export type WebSocketMessage =
//   | {
//       type: "NEW_ORDER";
//       data: OrderPayload;
//     }
//   | {
//       type: "ORDER_STATUS_UPDATE";
//       data: {
//         orderId: number;
//         status: string;
//         timestamp: string;
//       };
//     };

export class WebSocketService {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectTimeout = 3000;
  private messageHandlers: ((message: WebSocketMessage) => void)[] = [];
  private connectHandlers: (() => void)[] = [];
  private disconnectHandlers: (() => void)[] = [];
  private userName: string;
  private role: string;
  private userToken: string;
  private tableToken: string;
  private email: string;

  constructor(
    userName: string,
    role: string,
    userToken: string,
    tableToken: string,
    email: string
  ) {
    this.email = email;
    this.userName = userName;
    this.role = role;
    this.userToken = userToken;
    this.tableToken = tableToken;
    this.connect();
  }

  public connect() {
    try {
      const wsUrl = `ws://localhost:8888/ws/${this.role.toLowerCase()}/${
        this.userName
      }?token=${this.userToken}&tableToken=${this.tableToken}&email=${
        this.email
      }`;
      console.log(
        "quananqr1/zusstand/web-socket/websoket-service.ts wsUrl",
        wsUrl
      );

      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => {
        console.log("WebSocket connected successfully");
        this.reconnectAttempts = 0;
        this.connectHandlers.forEach((handler) => handler());
      };

      this.ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data) as WebSocketMessage;
          this.messageHandlers.forEach((handler) => handler(message));
        } catch (error) {
          console.error("Error parsing WebSocket message:", error);
        }
      };

      this.ws.onclose = (event) => {
        console.log(
          `WebSocket disconnected: code=${event.code}, reason=${event.reason}`
        );
        this.disconnectHandlers.forEach((handler) => handler());
        if (event.code !== 1000) {
          this.attemptReconnect();
        }
      };

      this.ws.onerror = (error) => {
        console.error("WebSocket error:", error);
      };
    } catch (error) {
      console.error("Error creating WebSocket connection:", error);
      this.attemptReconnect();
    }
  }
  private attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const backoffTime =
        this.reconnectTimeout * Math.pow(2, this.reconnectAttempts - 1);
      console.log(
        `Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts}) in ${backoffTime}ms...`
      );
      setTimeout(() => this.connect(), backoffTime);
    } else {
      console.error(
        "Max reconnection attempts reached. Please check your connection and try again."
      );
    }
  }

  public sendMessage(message: WebSocketMessage) {
    if (!this.ws) {
      console.error("WebSocket instance not initialized");
      return;
    }

    if (this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.error(
        `WebSocket is not open (current state: ${this.ws.readyState})`
      );
    }
  }

  public onMessage(handler: (message: WebSocketMessage) => void) {
    this.messageHandlers.push(handler);
    return () => {
      this.messageHandlers = this.messageHandlers.filter((h) => h !== handler);
    };
  }

  public onConnect(handler: () => void) {
    this.connectHandlers.push(handler);
    return () => {
      this.connectHandlers = this.connectHandlers.filter((h) => h !== handler);
    };
  }

  public onDisconnect(handler: () => void) {
    this.disconnectHandlers.push(handler);
    return () => {
      this.disconnectHandlers = this.disconnectHandlers.filter(
        (h) => h !== handler
      );
    };
  }

  public disconnect() {
    if (this.ws) {
      this.ws.close(1000, "Normal closure");
      this.ws = null;
    }
  }
}
