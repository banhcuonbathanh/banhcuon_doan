
// utils/orderUtils.ts
import { logWithLevel } from "@/lib/log";
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



export const aggregateDishes = (orders: OrderDetailedResponse[]): AggregatedDish[] => {
  const dishMap = new Map<number, AggregatedDish>();
  logWithLevel(
    { dishMap },
    "quananqr1/app/manage/admin/orders/restaurant-summary/restaurant-summary.tsx",
    "info",
    1
  );
  
  orders.forEach((order) => {
    // Add individual dishes
    order.data_dish.forEach((dish) => {
      const existingDish = dishMap.get(dish.dish_id);
      if (existingDish) {
        existingDish.quantity += dish.quantity;
      } else {
        dishMap.set(dish.dish_id, {
          ...dish,
          quantity: dish.quantity
        });
      }
    });

    // Add dishes from sets
    order.data_set.forEach((set) => {
      set.dishes.forEach((setDish) => {
        const existingDish = dishMap.get(setDish.dish_id);
        if (existingDish) {
          existingDish.quantity += setDish.quantity * set.quantity;
        } else {
          dishMap.set(setDish.dish_id, {
            ...setDish,
            quantity: setDish.quantity * set.quantity
          });
        }
      });
    });
  });

  return Array.from(dishMap.values());
};

