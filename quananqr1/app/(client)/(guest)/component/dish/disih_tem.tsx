"use client";

import React from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Plus, Minus } from "lucide-react";
import { DishInterface } from "@/schemaValidations/interface/type_dish";
import useOrderStore from "@/zusstand/order/order_zustand";

interface DishCardProps {
  dish: DishInterface;
}

export const DishCard: React.FC<DishCardProps> = ({ dish }) => {
  const { addDishItem, removeDishItem, updateDishQuantity, findDishOrderItem } =
    useOrderStore();

  const currentDish = findDishOrderItem(dish.id);
  const quantity = currentDish ? currentDish.quantity : 0;

  const handleIncrease = () => {
    if (currentDish) {
      updateDishQuantity(dish.id, currentDish.quantity + 1);
    } else {
      addDishItem(dish, 1);
    }
  };

  const handleDecrease = () => {
    if (currentDish) {
      if (currentDish.quantity > 1) {
        updateDishQuantity(dish.id, currentDish.quantity - 1);
      } else {
        removeDishItem(dish.id);
      }
    }
  };

  return (
    <Card className="w-full">
      <CardContent className="p-4 flex">
        <div className="w-1/3 pr-4">
          <img
            src={dish.image || "/api/placeholder/150/150"}
            alt={dish.name}
            className="w-full h-full object-cover rounded-md"
          />
        </div>
        <div className="w-2/3 flex flex-col justify-between">
          <div>
            <h3 className="text-lg font-semibold mb-2">{dish.name}</h3>
            <p className="text-sm mb-2">{dish.description}</p>
            <p className="font-semibold">Price: ${dish.price.toFixed(2)}</p>
          </div>
          <div className="flex items-center justify-end mt-2">
            <div className="flex items-center space-x-2">
              <Button variant="outline" size="sm" onClick={handleDecrease}>
                <Minus className="h-3 w-3" />
              </Button>
              <span className="w-8 text-center">{quantity}</span>
              <Button variant="outline" size="sm" onClick={handleIncrease}>
                <Plus className="h-3 w-3" />
              </Button>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default DishCard;
