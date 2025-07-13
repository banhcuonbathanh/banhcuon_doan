"use client";

import { DishInterface } from "@/schemaValidations/interface/type_dish";

import { DishCard } from "./disih_tem";
import React from "react";

interface DishSelectionProps {
  dishes: DishInterface[];
}

export function DishSelection({ dishes }: DishSelectionProps) {
  const [isMounted, setIsMounted] = React.useState(false);

  React.useEffect(() => {
    setIsMounted(true);
  }, []);

  if (!isMounted) {
    return null; // or a loading skeleton
  }
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
