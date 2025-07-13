import { Button } from "@/components/ui/button";
import {
  OrderContent,
  WebSocketMessage
} from "@/schemaValidations/interface/type_websocker";
import { useWebSocketStore } from "@/zusstand/web-socket/websocketStore";
import React, { useEffect } from "react";

interface OrderData {
  data: {
    id: string;
  };
}

interface Props {
  tableNumber: string;
  response: OrderData;
}

const OrderComponent1: React.FC<Props> = ({ tableNumber, response }) => {
  // Get the WebSocket store methods
  const { connect, sendMessage, isConnected } = useWebSocketStore();

  // Connect to WebSocket when component mounts
  useEffect(() => {
    // connect();
  }, [connect]);

  const handleSendOrder = () => {
    if (isConnected) {
      const message: WebSocketMessage = {
        type: "NEW_ORDER",
        content: {
          orderID: response.data.id,
          tableNumber,
          status: "pending",
          timestamp: new Date().toISOString(),
          orderData: response.data
        } as OrderContent,
        sender: "client",
        timestamp: new Date().toISOString()
      };

      // sendMessage(message);
    } else {
      console.error("WebSocket is not connected");
    }
  };
  return (
    <div>
      <Button onClick={handleSendOrder} disabled={!isConnected}>
        Send Order
      </Button>
    </div>
  );
};

export default OrderComponent1;
