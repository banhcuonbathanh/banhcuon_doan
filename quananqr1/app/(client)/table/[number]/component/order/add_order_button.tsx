"use client";

import React, { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import useOrderStore from "@/zusstand/order/order_zustand";
import { useOrderCreationStore } from "./logic";
import { useApiStore } from "@/zusstand/api/api-controller";
import { useAuthStore } from "@/zusstand/new_auth/new_auth_controller";
import { useWebSocketStore } from "@/zusstand/web-socket/websocketStore";
import { WebSocketMessage } from "@/schemaValidations/interface/type_websocker";

interface OrderCreationComponentProps {
  table_token: string;

  table_number: string;
}

const OrderCreationComponent: React.FC<OrderCreationComponentProps> = ({
  table_number,
  table_token
}) => {
  const { isLoading, createOrder } = useOrderCreationStore();
  const {
    getOrderSummary,
    clearOrder,
    canhKhongRau,
    canhCoRau,
    smallBowl,
    wantChili,
    selectedFilling
  } = useOrderStore();
  const { http } = useApiStore();
  const { guest, user, isGuest, openLoginDialog, userId, isLogin } =
    useAuthStore();
  const {
    connect,
    disconnect,
    isConnected,
    sendMessage,
    wsToken,
    fetchWsToken
  } = useWebSocketStore();

  // State to track authentication check status
  const [authChecked, setAuthChecked] = useState(false);

  // If selectedFilling is an object with a name or value property:
  const getFillingString = (filling: {
    mocNhi: boolean;
    thit: boolean;
    thitMocNhi: boolean;
  }) => {
    if (filling.mocNhi) return "Mọc Nhĩ";
    if (filling.thit) return "Thịt";
    if (filling.thitMocNhi) return "Thịt Mọc Nhĩ";
    return "Không";
  };

  let topping = `canhKhongRau ${canhKhongRau} - canhCoRau ${canhCoRau} - bat be ${smallBowl} - ot tuoi ${wantChili} - nhan ${getFillingString(
    selectedFilling
  )} -`;

  // Or if selectedFilling should be a primitive value:
  // Make sure selectedFilling is set as a string/number in your state management
  const orderSummary = getOrderSummary();

  // Initialize auth state when component mounts
  useEffect(() => {
    const initializeAuth = async () => {
      // Sync auth state from cookies
      useAuthStore.getState().syncAuthState();
      setAuthChecked(true);
    };

    initializeAuth();
  }, []);

  // Effect for WebSocket initialization after auth check
  useEffect(() => {
    if (authChecked && isLogin && userId) {
      console.log("Initializing WebSocket connection for user:", userId);
      initializeWebSocket();
    }
  }, [authChecked, isLogin, userId, user?.email, guest?.name]);

  const getEmailIdentifier = () => {
    if (isGuest && guest) {
      return guest.name;
    }
    // Use persisted user data if available
    const persistedUser = useAuthStore.getState().persistedUser;
    return persistedUser?.email || user?.email;
  };

  const initializeWebSocket = async () => {
    const emailIdentifier = getEmailIdentifier();
    console.log(
      "quananqr1/app/(client)/table/[number]/component/order/add_order_button.tsx isLogin, userId, emailIdentifier",
      isLogin,
      userId,
      emailIdentifier
    );
    if (!isLogin || !userId || !emailIdentifier) {
      console.log(
        "WebSocket initialization failed: Missing required credentials"
      );
      return;
    }

    try {
      const wstoken1 = await fetchWsToken({
        userId: Number(userId),
        email: emailIdentifier,
        role: isGuest ? "Guest" : "User"
      });

      if (!wstoken1) {
        throw new Error("Failed to obtain WebSocket token");
      }

      await connect({
        userId: userId.toString(),
        isGuest,
        userToken: wstoken1.token,
        tableToken: table_token,
        role: isGuest ? "Guest" : "User",
        email: emailIdentifier
      });
    } catch (error) {
      console.error("[OrderCreation] Connection error:", error);
    }
  };

  const handleCreateOrder = async () => {
    // Ensure auth state is synced before proceeding
    useAuthStore.getState().syncAuthState();
    const currentAuthState = useAuthStore.getState();

    if (!currentAuthState.isLogin) {
      console.log("[OrderCreation] User not logged in, showing login dialog");
      openLoginDialog();
      return;
    }

    if (!isConnected) {
      console.log("Attempting to establish WebSocket connection");
      await initializeWebSocket();
    }
    console.log(
      "quananqr1/app/(client)/table/[number]/component/order/add_order_button.tsx orderSummary",
      orderSummary
    );
    if (orderSummary.totalItems === 0) {
      console.log("[OrderCreation] No items in order, aborting 111111");
      return;
    }
    if (table_number === null) {
      console.log("[OrderCreation] No items in order, aborting 22222");
      return;
    }

    console.log("[OrderCreation] Creating order with summary:", orderSummary);

    const order = await createOrder({
      topping,
      Table_token: table_token,
      http,
      auth: { guest, user, isGuest },
      orderStore: {
        tableNumber: Number(table_number),
        getOrderSummary,
        clearOrder
      },
      websocket: { disconnect, isConnected, sendMessage },
      openLoginDialog
    });
    console.log(
      "quananqr1/app/(client)/table/[number]/component/order/add_order_button.tsx done for creating order done order  121212",
      order
    );
    sendMessage1();
    console.log(
      "quananqr1/app/(client)/table/[number]/component/order/add_order_button.tsx done for creating order done 121212"
    );
  };

  const getButtonText = () => {
    if (!authChecked) {
      return "Loading...";
    }
    if (!isLogin) {
      return "Login to Order";
    }
    if (orderSummary.totalItems === 0) {
      return "Add Items to Order";
    }
    return "Create Order";
  };

  const isButtonDisabled = () => {
    if (!authChecked) return true;
    if (!isLogin) return false;
    if (orderSummary.totalItems === 0) return true;
    return isLoading;
  };

  useEffect(() => {
    return () => {
      console.log(
        "[OrderCreation] Component unmounting, cleaning up connection"
      );
      disconnect();
    };
  }, []);
  //
  const sendMessage1 = async () => {
    console.log(
      "quananqr1/app/(client)/table/[number]/component/order/add_order_button.tsx sendMessage1"
    );
    // Ensure auth state is synced before proceeding
    useAuthStore.getState().syncAuthState();
    const currentAuthState = useAuthStore.getState();

    if (!currentAuthState.isLogin) {
      console.log("[SendMessage] User not logged in, showing login dialog");
      openLoginDialog();
      return;
    }

    if (!isConnected) {
      console.log("Attempting to establish WebSocket connection");
      await initializeWebSocket();
      if (!isConnected) {
        console.log(
          "[SendMessage] Failed to establish connection, aborting message send"
        );
        return;
      }
    }

    try {
      // Get order summary from the store
      const { dishes, sets, totalPrice } = getOrderSummary();

      const messagePayload: WebSocketMessage = {
        type: "order",
        action: "create_message",
        payload: {
          fromUserId: "1",
          toUserId: "2",
          type: "order",
          action: "new_order",
          payload: {
            guest_id: null,
            user_id: 1,
            is_guest: false,
            table_number: 1,
            order_handler_id: 1,
            status: "pending",
            created_at: "2024-10-21T12:00:00Z",
            updated_at: "2024-10-21T12:00:00Z",
            total_price: 5000,
            order_name: "test",
            dish_items: [
              {
                dish_id: 1,
                quantity: 2
              },
              {
                dish_id: 2,
                quantity: 2
              },
              {
                dish_id: 3,
                quantity: 4
              }
            ],
            set_items: [
              {
                set_id: 1,
                quantity: 3
              },
              {
                set_id: 2,
                quantity: 3
              }
            ],
            bow_chili: 1,
            bow_no_chili: 2,
            take_away: true,
            chili_number: 3,
            Table_token: "MTp0YWJsZTo0ODgzMjc3NDQy.2AZhkuCtKB0"
          }
        },
        role: "User",
        roomId: "1"
      };

      sendMessage(messagePayload);
      console.log("[SendMessage] Message sent successfully", messagePayload);
    } catch (error) {
      console.error("[SendMessage] Error sending message:", error);
    }
  };
  //
  return (
    <div className="mt-4">
      <Button
        className="w-full"
        onClick={handleCreateOrder}
        disabled={isButtonDisabled()}
      >
        {getButtonText()}
      </Button>

      <Button
        className="w-full"
        onClick={sendMessage1}
        disabled={isButtonDisabled()}
      >
        {"send message"}
      </Button>
    </div>
  );
};

export default OrderCreationComponent;
