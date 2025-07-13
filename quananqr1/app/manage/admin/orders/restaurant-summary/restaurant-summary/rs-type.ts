import { OrderDetailedDish, OrderDetailedResponse } from "@/schemaValidations/interface/type_order";


export interface RestaurantSummaryProps {
  restaurantLayoutProps: OrderDetailedResponse[];
}

export interface AggregatedDish extends OrderDetailedDish {}

export interface GroupedOrder {
  orderName: string;
  characteristic?: string;
  tableNumber: number;
  orders: OrderDetailedResponse[];
  hasTakeAway: boolean;
}

export interface OrderStore {
  tableNumber: number;
  getOrderSummary: () => {
    dishes: OrderDetailedDish[];
    totalPrice: number;
    orderId: number;
  };
  clearOrder: () => void;
}