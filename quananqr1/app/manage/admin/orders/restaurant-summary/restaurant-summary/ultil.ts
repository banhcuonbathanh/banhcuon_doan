import { OrderDetailedResponse } from "@/schemaValidations/interface/type_order";
import { AggregatedDish } from "./rs-type";


export const parseOrderName = (orderName: string): string => {
  const parts = orderName.split("-");
  return parts[0].trim();
};

export const getOrdinalSuffix = (num: number): string => {
  const j = num % 10;
  const k = num % 100;
  if (j === 1 && k !== 11) return "st";
  if (j === 2 && k !== 12) return "nd";
  if (j === 3 && k !== 13) return "rd";
  return "th";
};

export const aggregateDishes = (orders: OrderDetailedResponse[]): AggregatedDish[] => {
  const dishMap = new Map<number, AggregatedDish>();

  orders.forEach((order) => {
    // Add individual dishes
    order.data_dish.forEach((dish) => {
      const existingDish = dishMap.get(dish.dish_id);
      if (existingDish) {
        existingDish.quantity += dish.quantity;
      } else {
        dishMap.set(dish.dish_id, { ...dish, quantity: dish.quantity });
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