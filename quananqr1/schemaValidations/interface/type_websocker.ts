export interface WebSocketMessage {
  type: string;
  action: any;
  payload: {};
  role: string;
  roomId?: string;
}

export interface OrderContent {
  orderID: string;
  tableNumber: string;
  status: string;
  timestamp: string;
  orderData: any;
}

export interface WebSocketMessage21 {
  type: string;
  content: {
    orderID: number;
    tableNumber: string;
    status: string;
    timestamp: string;
    orderData?: {
      id: string;
      item: string;
      quantity: number;
    };
  };
  sender: string;
  timestamp: string;
}
