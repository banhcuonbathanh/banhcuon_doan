import { create } from "zustand";
import { persist } from "zustand/middleware";
import { toast } from "@/components/ui/use-toast";
import envConfig from "@/config";
import { logWithLevel } from "@/lib/log";
import { useApiStore } from "../api/api-controller";

const LOG_PATH = "quananqr1/zusstand/delivery/delivery_zustand.ts";

export const DeliveryStatusValues = {
  Pending: "Pending",
  Assigned: "Assigned",
  PickedUp: "Picked Up",
  InTransit: "In Transit",
  Delivered: "Delivered",
  Failed: "Failed",
  Cancelled: "Cancelled"
} as const;

export type DeliveryStatus =
  (typeof DeliveryStatusValues)[keyof typeof DeliveryStatusValues];

interface DishDeliveryItem {
  dish_id: number;
  quantity: number;
}

interface SingleDeliveryState {
  id: string;
  isLoading: boolean;
  guest_id: number | null;
  user_id: number;
  is_guest: boolean;
  table_number: number;
  order_handler_id: number;
  status: string;
  total_price: number;
  dish_items: DishDeliveryItem[];
  bow_chili: number;
  bow_no_chili: number;
  take_away: boolean;
  chili_number: number;
  table_token: string;
  client_name: string;
  delivery_address: string;
  delivery_contact: string;
  delivery_notes: string;
  scheduled_time: string;
  order_id: number;
  delivery_fee: number;
  delivery_status: DeliveryStatus;
  driver_id?: number;
  estimated_delivery_time?: string;
  actual_delivery_time?: string;
  created_at: string;
}

interface DeliveryState {
  deliveries: SingleDeliveryState[];
  deliveryHistory: Record<string, SingleDeliveryState>;
}

interface DeliveryActions {
  updateDeliveryInfo: (
    deliveryId: string,
    info: Partial<SingleDeliveryState>
  ) => void;
  updateStatus: (deliveryId: string, status: DeliveryStatus) => void;
  updateDriverInfo: (
    deliveryId: string,
    driverId: number,
    estimatedTime: string
  ) => void;
  completeDelivery: (deliveryId: string, actualTime: string) => void;
  addDishItem: (deliveryId: string, item: DishDeliveryItem) => void;
  removeDishItem: (deliveryId: string, dishId: number) => void;
  updateDishQuantity: (
    deliveryId: string,
    dishId: number,
    quantity: number
  ) => void;
  clearDelivery: (deliveryId: string) => void;
  clearAllDeliveries: () => void;
  getFormattedTotal: (deliveryId: string) => string;
  addToDeliveryHistory: (delivery: SingleDeliveryState) => void;
  getDeliveryHistory: () => Record<string, SingleDeliveryState>;
  getTotalDeliveredForDish: (dishId: number) => number;
  clearDeliveryHistory: () => void;
  createDelivery: (params: {
    guest: any;
    user: any;
    isGuest: boolean;
    orderStore: {
      tableNumber: number;
      getOrderSummary: () => any;
      clearOrder: () => void;
    };
    deliveryDetails: {
      deliveryAddress: string;
      deliveryContact: string;
      deliveryNotes: string;
      scheduledTime: string;
      deliveryFee: number;
    };
  }) => Promise<any>;
  getDeliveryById: (deliveryId: string) => SingleDeliveryState | undefined;
}

const INITIAL_STATE: DeliveryState = {
  deliveries: [],
  deliveryHistory: {}
};

function getPriceForDish(dishId: number): number {
  logWithLevel({ dishId }, LOG_PATH, "info", 4);
  return 1000; // Replace with actual price logic
}

function calculateTotalPrice(
  items: DishDeliveryItem[],
  deliveryFee: number
): number {
  logWithLevel({ items, deliveryFee }, LOG_PATH, "info", 4);
  const itemsTotal = items.reduce(
    (total, item) => total + item.quantity * getPriceForDish(item.dish_id),
    0
  );
  return itemsTotal + deliveryFee;
}

function formatCurrency(amount: number): string {
  return amount.toLocaleString("en-US", {
    style: "currency",
    currency: "USD"
  });
}

const useDeliveryStore = create<DeliveryState & DeliveryActions>()(
  persist(
    (set, get) => ({
      ...INITIAL_STATE,

      getDeliveryById: (deliveryId) => {
        return get().deliveries.find((delivery) => delivery.id === deliveryId);
      },

      updateDeliveryInfo: (deliveryId, info) => {
        logWithLevel({ deliveryId, info }, LOG_PATH, "info", 8);
        set((state) => ({
          deliveries: state.deliveries.map((delivery) =>
            delivery.id === deliveryId ? { ...delivery, ...info } : delivery
          ),
          deliveryHistory: {
            ...state.deliveryHistory,
            [deliveryId]: { ...state.deliveryHistory[deliveryId], ...info }
          }
        }));
      },

      updateStatus: (deliveryId, status) => {
        logWithLevel({ deliveryId, status }, LOG_PATH, "info", 5);
        set((state) => ({
          deliveries: state.deliveries.map((delivery) =>
            delivery.id === deliveryId
              ? { ...delivery, delivery_status: status }
              : delivery
          ),
          deliveryHistory: {
            ...state.deliveryHistory,
            [deliveryId]: {
              ...state.deliveryHistory[deliveryId],
              delivery_status: status
            }
          }
        }));
      },

      updateDriverInfo: (deliveryId, driverId, estimatedTime) => {
        logWithLevel(
          { deliveryId, driverId, estimatedTime },
          LOG_PATH,
          "info",
          5
        );
        const updateData = {
          driver_id: driverId,
          estimated_delivery_time: estimatedTime,
          delivery_status: "Assigned" as DeliveryStatus
        };
        set((state) => ({
          deliveries: state.deliveries.map((delivery) =>
            delivery.id === deliveryId
              ? { ...delivery, ...updateData }
              : delivery
          ),
          deliveryHistory: {
            ...state.deliveryHistory,
            [deliveryId]: {
              ...state.deliveryHistory[deliveryId],
              ...updateData
            }
          }
        }));
      },

      completeDelivery: (deliveryId, actualTime) => {
        logWithLevel({ deliveryId, actualTime }, LOG_PATH, "info", 5);
        const updateData = {
          actual_delivery_time: actualTime,
          delivery_status: "Delivered" as DeliveryStatus
        };
        set((state) => ({
          deliveries: state.deliveries.map((delivery) =>
            delivery.id === deliveryId
              ? { ...delivery, ...updateData }
              : delivery
          ),
          deliveryHistory: {
            ...state.deliveryHistory,
            [deliveryId]: {
              ...state.deliveryHistory[deliveryId],
              ...updateData
            }
          }
        }));
      },

      addDishItem: (deliveryId, item) => {
        logWithLevel(
          { action: "addDishItem", deliveryId, item },
          LOG_PATH,
          "info",
          3
        );
        set((state) => {
          const delivery = state.deliveries.find((d) => d.id === deliveryId);
          if (!delivery) return state;

          const newItems = [...delivery.dish_items, item];
          const newTotal = calculateTotalPrice(newItems, delivery.delivery_fee);
          const updateData = {
            dish_items: newItems,
            total_price: newTotal
          };

          return {
            deliveries: state.deliveries.map((d) =>
              d.id === deliveryId ? { ...d, ...updateData } : d
            ),
            deliveryHistory: {
              ...state.deliveryHistory,
              [deliveryId]: {
                ...state.deliveryHistory[deliveryId],
                ...updateData
              }
            }
          };
        });
      },

      removeDishItem: (deliveryId, dishId) => {
        logWithLevel(
          { action: "removeDishItem", deliveryId, dishId },
          LOG_PATH,
          "info",
          3
        );
        set((state) => {
          const delivery = state.deliveries.find((d) => d.id === deliveryId);
          if (!delivery) return state;

          const updatedItems = delivery.dish_items.filter(
            (item) => item.dish_id !== dishId
          );
          const newTotal = calculateTotalPrice(
            updatedItems,
            delivery.delivery_fee
          );
          const updateData = {
            dish_items: updatedItems,
            total_price: newTotal
          };

          return {
            deliveries: state.deliveries.map((d) =>
              d.id === deliveryId ? { ...d, ...updateData } : d
            ),
            deliveryHistory: {
              ...state.deliveryHistory,
              [deliveryId]: {
                ...state.deliveryHistory[deliveryId],
                ...updateData
              }
            }
          };
        });
      },

      updateDishQuantity: (deliveryId, dishId, quantity) => {
        logWithLevel(
          { action: "updateDishQuantity", deliveryId, dishId, quantity },
          LOG_PATH,
          "info",
          3
        );
        set((state) => {
          const delivery = state.deliveries.find((d) => d.id === deliveryId);
          if (!delivery) return state;

          const updatedItems = delivery.dish_items.map((item) =>
            item.dish_id === dishId ? { ...item, quantity } : item
          );
          const newTotal = calculateTotalPrice(
            updatedItems,
            delivery.delivery_fee
          );
          const updateData = {
            dish_items: updatedItems,
            total_price: newTotal
          };

          return {
            deliveries: state.deliveries.map((d) =>
              d.id === deliveryId ? { ...d, ...updateData } : d
            ),
            deliveryHistory: {
              ...state.deliveryHistory,
              [deliveryId]: {
                ...state.deliveryHistory[deliveryId],
                ...updateData
              }
            }
          };
        });
      },

      addToDeliveryHistory: (delivery) => {
        logWithLevel(
          { action: "addToDeliveryHistory", delivery },
          LOG_PATH,
          "info",
          3
        );
        set((state) => ({
          deliveryHistory: {
            ...state.deliveryHistory,
            [delivery.id]: delivery
          }
        }));
      },

      getDeliveryHistory: () => {
        return get().deliveryHistory;
      },

      getTotalDeliveredForDish: (dishId) => {
        const history = Object.values(get().deliveryHistory);
        return history.reduce((total, delivery) => {
          const dishItem = delivery.dish_items.find(
            (item) => item.dish_id === dishId
          );
          return total + (dishItem?.quantity || 0);
        }, 0);
      },

      clearDelivery: (deliveryId) => {
        logWithLevel(
          { action: "clearDelivery", deliveryId },
          LOG_PATH,
          "info",
          8
        );
        set((state) => {
          const { [deliveryId]: removed, ...remainingHistory } =
            state.deliveryHistory;
          return {
            deliveries: state.deliveries.filter(
              (delivery) => delivery.id !== deliveryId
            ),
            deliveryHistory: remainingHistory
          };
        });
      },

      clearAllDeliveries: () => {
        set(INITIAL_STATE);
      },

      clearDeliveryHistory: () => {
        set((state) => ({ ...state, deliveryHistory: {} }));
      },

      getFormattedTotal: (deliveryId) => {
        const delivery = get().deliveries.find((d) => d.id === deliveryId);
        return delivery
          ? formatCurrency(delivery.total_price)
          : formatCurrency(0);
      },

      createDelivery: async ({
        guest,
        user,
        isGuest,
        orderStore,
        deliveryDetails
      }) => {
        try {
          if (!orderStore?.getOrderSummary) {
            throw new Error("Order summary function is required");
          }

          if (!orderStore?.tableNumber) {
            throw new Error("Table number is required");
          }

          const orderSummary = orderStore.getOrderSummary();
          logWithLevel({ orderSummary }, LOG_PATH, "info", 9);

          if (!orderSummary?.dishes?.length) {
            throw new Error("No dishes found in order");
          }

          if (isGuest && (!guest || !guest.id)) {
            throw new Error("Guest ID is required for guest orders");
          }

          if (!isGuest && (!user || !user.id)) {
            throw new Error("User ID is required for user orders");
          }

          const dish_items = orderSummary.dishes.map((dish: any) => ({
            dish_id: dish.id,
            quantity: dish.quantity
          }));

          const newDelivery: SingleDeliveryState = {
            id: `del_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
            isLoading: false,
            guest_id: isGuest ? guest.id : null,
            user_id: isGuest ? null : user.id,
            is_guest: isGuest,
            table_number: orderStore.tableNumber,
            order_handler_id: 1,
            status: "Pending",
            total_price: calculateTotalPrice(
              dish_items,
              deliveryDetails.deliveryFee
            ),
            dish_items,
            bow_chili: 0,
            bow_no_chili: 0,
            take_away: false,
            chili_number: 0,
            table_token: "",
            client_name: isGuest ? guest?.name : user?.name,
            delivery_address: deliveryDetails.deliveryAddress,
            delivery_contact: deliveryDetails.deliveryContact,
            delivery_notes: deliveryDetails.deliveryNotes,
            scheduled_time: deliveryDetails.scheduledTime,
            order_id: orderSummary.orderId,
            delivery_fee: deliveryDetails.deliveryFee,
            delivery_status: "Pending",
            created_at: new Date().toISOString()
          };

          set((state) => ({
            deliveries: [...state.deliveries, newDelivery],
            deliveryHistory: {
              ...state.deliveryHistory,
              [newDelivery.id]: newDelivery
            }
          }));

          const deliveryEndpoint = `${envConfig.NEXT_PUBLIC_API_ENDPOINT}${envConfig.Delivery_External_End_Point}`;
          const response = await useApiStore
            .getState()
            .http.post(deliveryEndpoint, newDelivery);

          logWithLevel(
            {
              message: "Delivery created successfully",
              deliveryData: newDelivery,
              responseData: response.data
            },
            LOG_PATH,
            "info",
            1
          );

          toast({
            title: "Success",
            description: "Delivery has been created successfully"
          });

          orderStore.clearOrder();
          return response.data;
        } catch (error) {
          logWithLevel(
            {
              message: "Failed to create delivery",
              error,
              errorMessage:
                error instanceof Error ? error.message : "Unknown error"
            },
            LOG_PATH,
            "error",
            7
          );

          toast({
            variant: "destructive",
            title: "Error",
            description:
              error instanceof Error
                ? error.message
                : "Failed to create delivery"
          });

          throw error;
        }
      }
    }),
    {
      name: "delivery-storage",
      partialize: (state) => ({
        deliveries: state.deliveries,
        deliveryHistory: state.deliveryHistory
      }),
      version: 1 // Add version for potential future migrations
    }
  )
);

export default useDeliveryStore;
