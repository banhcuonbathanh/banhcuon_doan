"use client";

import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { ChevronDown, ChevronUp } from "lucide-react";
import useOrderStore from "@/zusstand/order/order_zustand";
const OrderSummary = () => {
  const {
    getFormattedSets,
    getFormattedDishes,
    getFormattedTotals,
    updateDishQuantity,
    updateSetQuantity,
    removeDishItem,
    removeSetItem,
    dishItems,
    setItems
  } = useOrderStore();

  const [showSets, setShowSets] = useState(true);
  const [showDishes, setShowDishes] = useState(true);
  const [bowlChili, setBowlChili] = useState(0);
  const [bowlNoChili, setBowlNoChili] = useState(0);

  const formattedSets = getFormattedSets();
  const formattedDishes = getFormattedDishes();
  const totals = getFormattedTotals();

  const handleDishQuantityChange = (id: number, change: number) => {
    const dish = dishItems.find((d) => d.id === id);
    if (dish) {
      const newQuantity = dish.quantity + change;
      if (newQuantity > 0) {
        updateDishQuantity(id, newQuantity);
      } else {
        removeDishItem(id);
      }
    }
  };

  const handleSetQuantityChange = (id: number, change: number) => {
    const set = setItems.find((s) => s.id === id);
    if (set) {
      const newQuantity = set.quantity + change;
      if (newQuantity > 0) {
        updateSetQuantity(id, newQuantity);
      } else {
        removeSetItem(id);
      }
    }
  };

  const handleBowlChange = (type: "chili" | "noChili", change: number) => {
    if (type === "chili") {
      const newValue = bowlChili + change;
      if (newValue >= 0) setBowlChili(newValue);
    } else {
      const newValue = bowlNoChili + change;
      if (newValue >= 0) setBowlNoChili(newValue);
    }
  };

  return (
    <div className="container mx-auto px-4 py-5 space-y-5">
      <Card>
        <CardHeader>
          <CardTitle>canh banh cuon</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-4">
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span>canh khong rau</span>
                <div className="flex items-center gap-2">
                  <Button
                    size="sm"
                    onClick={() => handleBowlChange("chili", -1)}
                  >
                    -
                  </Button>
                  <span className="mx-2">{bowlChili}</span>
                  <Button
                    size="sm"
                    onClick={() => handleBowlChange("chili", 1)}
                  >
                    +
                  </Button>
                </div>
              </div>
              <div className="flex items-center justify-between">
                <span>canh rau </span>
                <div className="flex items-center gap-2">
                  <Button
                    size="sm"
                    onClick={() => handleBowlChange("noChili", -1)}
                  >
                    -
                  </Button>
                  <span className="mx-2">{bowlNoChili}</span>
                  <Button
                    size="sm"
                    onClick={() => handleBowlChange("noChili", 1)}
                  >
                    +
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex justify-between items-center">
            <span>Order Summary</span>
            <span className="text-base">
              Total: {totals.formattedTotalPrice}
            </span>
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <div>
            <div className="flex items-center justify-between mb-3">
              <h3 className="font-semibold text-lg">Sets</h3>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setShowSets(!showSets)}
              >
                {showSets ? <ChevronUp /> : <ChevronDown />}
              </Button>
            </div>
            {showSets &&
              formattedSets.map((set) => (
                <div key={set.id} className="mb-6 border-b pb-4">
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="font-medium">
                        {set.displayString}: {set.formattedTotalPrice}
                      </span>

                      <div className="flex items-center gap-2">
                        <Button
                          size="sm"
                          onClick={() => handleSetQuantityChange(set.id, -1)}
                        >
                          -
                        </Button>
                        <span className="mx-2">
                          {setItems.find((s) => s.id === set.id)?.quantity || 0}
                        </span>
                        <Button
                          size="sm"
                          onClick={() => handleSetQuantityChange(set.id, 1)}
                        >
                          +
                        </Button>
                      </div>
                    </div>
                    <div className="ml-4 text-sm text-gray-600">
                      {set.itemsString.split(", ").map((item, index) => (
                        <div key={index}>{item}</div>
                      ))}
                    </div>
                  </div>
                </div>
              ))}
          </div>

          <div>
            <div className="flex items-center justify-between mb-3">
              <h3 className="font-semibold text-lg">Dishes</h3>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setShowDishes(!showDishes)}
              >
                {showDishes ? <ChevronUp /> : <ChevronDown />}
              </Button>
            </div>
            {showDishes &&
              formattedDishes.map((dish) => (
                <div key={dish.id} className="mb-4 border-b pb-4">
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="font-medium">
                        {dish.displayString}: {dish.formattedTotalPrice}
                      </span>

                      <div className="flex items-center gap-2">
                        <Button
                          size="sm"
                          onClick={() => handleDishQuantityChange(dish.id, -1)}
                        >
                          -
                        </Button>
                        <span className="mx-2">
                          {dishItems.find((d) => d.id === dish.id)?.quantity ||
                            0}
                        </span>
                        <Button
                          size="sm"
                          onClick={() => handleDishQuantityChange(dish.id, 1)}
                        >
                          +
                        </Button>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
          </div>
        </CardContent>
      </Card>

      {/* <div className="mt-4">
        <Button className="w-full" onClick={() => {}}>
          Add Order
        </Button>
      </div> */}
    </div>
  );
};

export default OrderSummary;
