"use client";

import Image from "next/image";
import React from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Plus, Minus } from "lucide-react";
import {
  SetInterface,
  SetProtoDish
} from "@/schemaValidations/interface/types_set";
import useOrderStore from "@/zusstand/order/order_zustand";
import DishList from "./set_dish";

interface SetSelectionProps {
  set: SetInterface;
}

const SetCard: React.FC<SetSelectionProps> = ({ set }) => {
  // Use the store hook to subscribe to specific state
  // const setOrderItem = useOrderStore(
  //   React.useCallback(
  //     (state) => state.setItems.find((item) => item.id === set.id),
  //     [set.id]
  //   )
  // );

  const orderStore = useOrderStore();

  // Find the set item directly from the store
  const setOrderItem = orderStore.setItems.find((item) => item.id === set.id);

  const { addSetItem, updateSetDishes, updateSetQuantity, removeSetItem } =
    orderStore;

  // Initialize dishQuantities with the default quantities from set.dishes
  const [dishQuantities, setDishQuantities] = React.useState<
    Record<number, number>
  >({});

  // Move initialization to useEffect to handle initial state
  React.useEffect(() => {
    setDishQuantities(
      set.dishes.reduce(
        (acc, dish) => ({ ...acc, [dish.dish_id]: dish.quantity }),
        {}
      )
    );
  }, [set.dishes]);

  const totalPrice = React.useMemo(() => {
    return set.dishes.reduce(
      (sum, dish) => sum + dish.price * (dishQuantities[dish.dish_id] || 0),
      0
    );
  }, [set.dishes, dishQuantities]);

  const handleIncrease = React.useCallback(() => {
    if (setOrderItem) {
      // Increase quantity of existing set item
      orderStore.updateSetQuantity(set.id, (setOrderItem.quantity || 0) + 1);
    } else {
      // Add new set item with quantity 1
      orderStore.addSetItem(set, 1);
    }
  }, [set, setOrderItem, orderStore]);

  const handleDecrease = React.useCallback(() => {
    if (setOrderItem) {
      if ((setOrderItem.quantity || 0) > 1) {
        // Decrease quantity
        orderStore.updateSetQuantity(set.id, (setOrderItem.quantity || 0) - 1);
      } else {
        // Remove item if quantity is 1
        orderStore.removeSetItem(set.id);
      }
    }
  }, [set.id, setOrderItem, orderStore]);

  const handleDishIncrease = React.useCallback(
    (dishId: number) => {
      setDishQuantities((prev) => {
        const newQuantities = {
          ...prev,
          [dishId]: (prev[dishId] || 0) + 1
        };
        if (setOrderItem) {
          const updatedDishes: SetProtoDish[] = set.dishes.map((dish) => ({
            dish_id: dish.dish_id,
            quantity: newQuantities[dish.dish_id] || 0,
            name: dish.name,
            price: dish.price,
            description: dish.description,
            image: dish.image,
            status: dish.status
          }));
          updateSetDishes(set.id, updatedDishes);
        }
        return newQuantities;
      });
    },
    [set, setOrderItem, updateSetDishes]
  );

  const handleDishDecrease = React.useCallback(
    (dishId: number) => {
      setDishQuantities((prev) => {
        const newQuantities = {
          ...prev,
          [dishId]: Math.max(0, (prev[dishId] || 0) - 1)
        };
        if (setOrderItem) {
          const updatedDishes: SetProtoDish[] = set.dishes.map((dish) => ({
            dish_id: dish.dish_id,
            quantity: newQuantities[dish.dish_id] || 0,
            name: dish.name,
            price: dish.price,
            description: dish.description,
            image: dish.image,
            status: dish.status
          }));
          updateSetDishes(set.id, updatedDishes);
        }
        return newQuantities;
      });
    },
    [set, setOrderItem, updateSetDishes]
  );

  return (
    <Card className="w-full">
      <CardContent className="p-4 flex">
        <div className="w-1/3 pr-4">
          <Image
            src={set.image || "/api/placeholder/300/400"}
            alt={set.name}
            className="w-full h-full object-cover rounded-md"
            width={200}
            height={200}
            priority
          />
        </div>
        <div className="w-2/3 flex flex-col justify-between">
          <div className="space-y-2">
            <div className="flex flex-row justify-between">
              <h2 className="text-2xl font-bold">{set.name}</h2>
              <span className="font-semibold text-lg">
                {typeof set.price === "number" ? `${set.price}k` : ""}
              </span>
            </div>
            <p className="text-sm text-gray-600">{set.description}</p>
          </div>
          <DishList
            dishes={set.dishes}
            dishQuantities={dishQuantities}
            onIncrease={handleDishIncrease}
            onDecrease={handleDishDecrease}
          />
          <div className="flex items-center justify-end mt-4">
            <div className="flex items-center space-x-4">
              <Button
                variant="outline"
                onClick={handleDecrease}
                disabled={!setOrderItem}
              >
                <Minus className="h-4 w-4" />
              </Button>
              <span className="font-semibold w-8 text-center">
                {setOrderItem?.quantity || 0}
              </span>
              <Button variant="default" onClick={handleIncrease}>
                <Plus className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default SetCard;
