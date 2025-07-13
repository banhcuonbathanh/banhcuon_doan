

import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Plus, Minus } from "lucide-react";
import { DishInterface } from "@/schemaValidations/interface/type_dish";

import { DishCard } from "./disih_tem";

interface DishSelectionProps {
  dishes: DishInterface[];
}

export function DishSelection({ dishes }: DishSelectionProps) {
  return (
    <div className="container mx-auto px-4 py-8">
    
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {dishes.map((dish: DishInterface) => (
          <DishCard key={dish.id} dish={dish} />
        ))}
      </div>
    </div>
  );
}
