import { create } from "zustand";
import { toast } from "@/components/ui/use-toast";
import envConfig from "@/config";
import { CreateOrderRequest } from "@/schemaValidations/interface/type_order";

interface OrderCreationState {
  isLoading: boolean;
  createOrder: (params: {
    topping: string;
    Table_token: string;
    http: any;
    auth: {
      guest: any;
      user: any;
      isGuest: boolean;
    };
    orderStore: {
      tableNumber: number;
      getOrderSummary: () => any;
      clearOrder: () => void;
    };
    websocket: {
      disconnect: () => void;
      isConnected: boolean;
      sendMessage: (message: any) => void;
    };
    openLoginDialog: () => void;
  }) => Promise<any>; // Changed return type to Promise<any>
}

export const useOrderCreationStore = create<OrderCreationState>((set) => ({
  isLoading: false,

  createOrder: async ({
    topping,
    Table_token,
    http,
    auth: { guest, user, isGuest },
    orderStore: { tableNumber, getOrderSummary, clearOrder },
    websocket: { disconnect, isConnected, sendMessage },
    openLoginDialog
  }) => {
    if (!user && !guest) {
      openLoginDialog();
      return;
    }

    const orderSummary = getOrderSummary();

    const dish_items = orderSummary.dishes.map((dish: any) => ({
      dish_id: dish.id,
      quantity: dish.quantity
    }));

    const set_items = orderSummary.sets.map((set: any) => ({
      set_id: set.id,
      quantity: set.quantity
    }));

    const user_id = isGuest ? null : user?.id ?? null;
    const guest_id = isGuest ? guest?.id ?? null : null;
    let order_name = "";
    if (isGuest && guest) {
      order_name = guest.name;
    } else if (!isGuest && user) {
      order_name = user.name;
    }

    const orderData: CreateOrderRequest = {
      guest_id,
      user_id,
      is_guest: isGuest,
      table_number: tableNumber,
      order_handler_id: 1,
      status: "pending",
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      total_price: orderSummary.totalPrice,
      dish_items,
      set_items,
      topping: topping,
      tracking_order: "tracking_order",
      takeAway: false,
      chiliNumber: 0,
      table_token: Table_token,
      order_name
    };

    set({ isLoading: true });

    try {
      const link_order = `${envConfig.NEXT_PUBLIC_API_ENDPOINT}${envConfig.Order_External_End_Point}`;
      const response = await http.post(link_order, orderData);

      if (isConnected) {
        sendMessage({
          type: "NEW_ORDER",
          data: {
            orderId: response.data.id,
            orderData
          }
        });
      }

      toast({
        title: "Success",
        description: "Order has been created successfully"
      });

      clearOrder();

      return response.data; // Return the response data
    } catch (error) {
      console.error("Order creation failed:", error);
      toast({
        variant: "destructive",
        title: "Error",
        description:
          error instanceof Error ? error.message : "Failed to create order"
      });
      throw error; // Re-throw the error so the caller knows something went wrong
    } finally {
      set({ isLoading: false });
      disconnect();
    }
  }
}));
