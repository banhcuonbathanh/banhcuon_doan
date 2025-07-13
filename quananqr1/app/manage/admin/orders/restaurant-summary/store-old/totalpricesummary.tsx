"use client";

import React, { useMemo, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ChevronDown, ChevronRight } from "lucide-react";
import { OrderDetailedResponse } from "../component/new-order-column";

interface TotalPriceSummaryProps {
  TotalPriceProps: OrderDetailedResponse[];
}

export const TotalPriceSummary: React.FC<TotalPriceSummaryProps> = ({
  TotalPriceProps
}) => {
  const [isOpen, setIsOpen] = useState(false);

  const calculateTotals = useMemo(() => {
    let totalDishPrice = 0;
    let totalSetPrice = 0;

    TotalPriceProps.forEach((order) => {
      // Calculate individual dishes total
      order.data_dish.forEach((dish) => {
        totalDishPrice += dish.price * dish.quantity;
      });

      // Calculate set dishes total
      order.data_set.forEach((set) => {
        totalSetPrice += set.price * set.quantity;
      });
    });

    return {
      dishTotal: totalDishPrice,
      setTotal: totalSetPrice,
      grandTotal: totalDishPrice + totalSetPrice
    };
  }, [TotalPriceProps]);

  return (
    <Card className="w-full">
      <CardHeader
        className="cursor-pointer select-none"
        onClick={() => setIsOpen(!isOpen)}
      >
        <div className="flex items-center justify-between">
          <CardTitle>
            Total Price Summary ${calculateTotals.grandTotal.toFixed(2)}
          </CardTitle>
          {isOpen ? (
            <ChevronDown className="h-4 w-4" />
          ) : (
            <ChevronRight className="h-4 w-4" />
          )}
        </div>
      </CardHeader>

      {isOpen && (
        <CardContent>
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-2">
              <div className="font-medium">Individual Dishes Total:</div>
              <div className="text-right">
                ${calculateTotals.dishTotal.toFixed(2)}
              </div>

              <div className="font-medium">Set Orders Total:</div>
              <div className="text-right">
                ${calculateTotals.setTotal.toFixed(2)}
              </div>

              <div className="font-medium border-t pt-2">Grand Total:</div>
              <div className="text-right font-bold border-t pt-2">
                ${calculateTotals.grandTotal.toFixed(2)}
              </div>
            </div>
          </div>
        </CardContent>
      )}
    </Card>
  );
};
