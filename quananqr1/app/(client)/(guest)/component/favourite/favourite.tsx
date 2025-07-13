import React, { useState, useEffect } from "react";
import {
  Card,
  CardHeader,
  CardContent,
  CardFooter,
  CardTitle
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useToast } from "@/components/ui/use-toast";
import {
  FavoriteSetListResType,
  DishListResType,
  Dish,
  FavoriteSet
} from "@/schemaValidations/dish.schema";

interface FavoriteSetProps {
  favoriteSets: FavoriteSetListResType;
  dishes: DishListResType;
  onAddToOrder: (dish: Dish) => void;
}

export function FavoriteSets({
  favoriteSets,
  dishes,
  onAddToOrder
}: FavoriteSetProps) {
  const [newSetName, setNewSetName] = useState("");
  const [selectedDishes, setSelectedDishes] = useState<number[]>([]);
  const { toast } = useToast();

  const createFavoriteSet = async () => {
    try {
      // API call to create a new favorite set
      const response = await fetch("/api/favorite-sets", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: newSetName, dishes: selectedDishes })
      });
      if (response.ok) {
        toast({ title: "Favorite set created successfully!" });
        setNewSetName("");
        setSelectedDishes([]);
        // Refresh favorite sets list
      } else {
        toast({
          title: "Failed to create favorite set",
          variant: "destructive"
        });
      }
    } catch (error) {
      console.error("Error creating favorite set:", error);
      toast({ title: "An error occurred", variant: "destructive" });
    }
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <h2 className="text-2xl font-bold mb-4">Your Favorite Sets</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
        {favoriteSets.map((set) => (
          <FavoriteSetCard
            key={set.id}
            set={set}
            dishes={dishes}
            onAddToOrder={onAddToOrder}
          />
        ))}
      </div>

      <h3 className="text-xl font-bold mb-4">Create New Favorite Set</h3>
      <div className="flex items-center gap-4 mb-4">
        <Input
          type="text"
          placeholder="Set Name"
          value={newSetName}
          onChange={(e) => setNewSetName(e.target.value)}
        />
        <Button onClick={createFavoriteSet}>Create Set</Button>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {dishes.map((dish) => (
          <div key={dish.id} className="flex items-center">
            <input
              type="checkbox"
              id={`dish-${dish.id}`}
              checked={selectedDishes.includes(dish.id)}
              onChange={() => {
                setSelectedDishes((prev) =>
                  prev.includes(dish.id)
                    ? prev.filter((id) => id !== dish.id)
                    : [...prev, dish.id]
                );
              }}
              className="mr-2"
            />
            <label htmlFor={`dish-${dish.id}`}>{dish.name}</label>
          </div>
        ))}
      </div>
    </div>
  );
}

interface FavoriteSetCardProps {
  set: FavoriteSet;
  dishes: DishListResType;
  onAddToOrder: (dish: Dish) => void;
}

const FavoriteSetCard: React.FC<FavoriteSetCardProps> = ({
  set,
  dishes,
  onAddToOrder
}) => {
  const setDishes = dishes.filter((dish) => set.dishes.includes(dish.id));

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>{set.name}</CardTitle>
      </CardHeader>
      <CardContent>
        <ul>
          {setDishes.map((dish) => (
            <li key={dish.id} className="mb-2">
              {dish.name} - ${dish.price.toFixed(2)}
            </li>
          ))}
        </ul>
      </CardContent>
      <CardFooter>
        <Button
          onClick={() => setDishes.forEach(onAddToOrder)}
          className="w-full"
        >
          Add All to Order
        </Button>
      </CardFooter>
    </Card>
  );
};
