import React from "react";
import { Plus, Minus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { SetProtoDish } from "@/schemaValidations/interface/types_set";

interface DishListProps {
  dishes: SetProtoDish[];
  dishQuantities: Record<number, number>;
  onIncrease: (dishId: number) => void;
  onDecrease: (dishId: number) => void;
}

const DishList: React.FC<DishListProps> = ({
  dishes,
  dishQuantities,
  onIncrease,
  onDecrease
}) => {
  // console.log("quananqr1/app/(guest)/component/set/set_dish.tsx ", dishes)
  return (
    <div className="mt-4">
      <h3 className="font-semibold mb-2">Dishes:</h3>
      <div className="space-y-2">
        {dishes.map((dish) => (
          <div
            key={`dish-${dish.dish_id}`}
            className="flex justify-between items-center py-2 border-b"
          >
            <span className="font-medium">{dish.name}</span>
            <div className="flex items-center gap-4">
              <span>${dish.price.toFixed(2)}</span>
              <span>Qty: {dishQuantities[dish.dish_id] || 0}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default DishList;
