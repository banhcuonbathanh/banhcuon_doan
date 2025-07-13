"use client";

import React, { useState, useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Card, CardHeader, CardContent } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import {
  OrderDetailedDish,
  OrderSetDetailed
} from "@/schemaValidations/interface/type_order";
import { TableMeta } from "./new-order-column";
import NumericKeypadInput from "./numberpad-dialog";

export const OrderTracking12 = ({
  row,
  meta
}: {
  row: any;
  meta: TableMeta;
}) => {
  const [deliveryState, setDeliveryState] = useState<Map<string, number>>(
    new Map()
  );
  const [amountPaid, setAmountPaid] = useState<string>("");
  const [change, setChange] = useState<number | null>(null);

  useEffect(() => {
    if (row.original.deliveryData) {
      setDeliveryState(new Map(Object.entries(row.original.deliveryData)));
    }
  }, [row.original.deliveryData]);

  const calculateDishTotals = () => {
    const dishTotals = new Map<string, { quantity: number; price: number }>();

    row.original.data_set?.forEach((set: OrderSetDetailed) => {
      set.dishes.forEach((dish) => {
        const totalQuantity = set.quantity * dish.quantity;
        const dishPrice = dish.price * dish.quantity;
        const current = dishTotals.get(dish.name) || { quantity: 0, price: 0 };
        dishTotals.set(dish.name, {
          quantity: current.quantity + totalQuantity,
          price: current.price + dishPrice * set.quantity
        });
      });
    });

    row.original.data_dish?.forEach((dish: OrderDetailedDish) => {
      const current = dishTotals.get(dish.name) || { quantity: 0, price: 0 };
      dishTotals.set(dish.name, {
        quantity: current.quantity + dish.quantity,
        price: current.price + dish.price * dish.quantity
      });
    });

    return dishTotals;
  };

  const calculateTotals = () => {
    const dishTotals = calculateDishTotals();
    const totalOrderValue = Array.from(dishTotals.values()).reduce(
      (sum, total) => sum + total.price,
      0
    );
    const deliveredValue = Array.from(dishTotals.entries()).reduce(
      (sum, [dishName, totals]) => {
        const delivered = deliveryState.get(dishName) || 0;
        const pricePerUnit = totals.price / totals.quantity;
        return sum + delivered * pricePerUnit;
      },
      0
    );
    const remainingValue = totalOrderValue - deliveredValue;

    return {
      totalOrderValue,
      deliveredValue,
      remainingValue
    };
  };

  const handlePaymentInput = (value: string) => {
    setAmountPaid(value);
    const numericValue = parseFloat(value) || 0;
    const { remainingValue } = calculateTotals();
    const changeAmount = numericValue - remainingValue;
    setChange(changeAmount >= 0 ? changeAmount : null);
  };

  const handleDeliveryUpdate =
    (dishName: string) => async (newValue: number) => {
      const dishTotals = calculateDishTotals();
      const totalQuantity = dishTotals.get(dishName)?.quantity || 0;
      const newDelivered = Math.min(newValue, totalQuantity);

      try {
        const response = await meta?.onDeliveryUpdate?.(
          row.original.id,
          dishName,
          newDelivered
        );

        const newState = new Map(deliveryState);
        newState.set(dishName, newDelivered);
        setDeliveryState(newState);
      } catch (error) {
        console.error("Failed to update delivery:", error);
      }
    };

  const dishTotals = calculateDishTotals();
  const { totalOrderValue, deliveredValue, remainingValue } = calculateTotals();

  return (
    <Card>
      <CardHeader className="py-3">
        <div className="text-sm font-medium text-gray-700">Order Summary</div>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Dish Items */}
        <DishDeliveryTable
          dishTotals={calculateDishTotals()}
          deliveryState={deliveryState}
          handleDeliveryUpdate={handleDeliveryUpdate}
        />

        <Separator />

        {/* Order Totals */}
        <div className="space-y-2">
          <div className="flex justify-between text-sm">
            <span className="font-medium">Total Order Value:</span>
            <span>${totalOrderValue.toFixed(2)}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="font-medium">Delivered Value:</span>
            <span className="text-green-600">${deliveredValue.toFixed(2)}</span>
          </div>
        </div>

        <Separator />

        {/* Payment Section */}
        <div className="space-y-3">
          <div className="flex justify-between items-center">
            <span className="text-sm font-medium">Amount Due:</span>
            <div className="flex items-center gap-2">
              <Input
                type="number"
                placeholder="0.00"
                value={amountPaid}
                onChange={(e) => handlePaymentInput(e.target.value)}
                className="w-24 h-8 text-right text-sm"
              />
              <span className="text-sm">$</span>
            </div>
          </div>

          {change !== null && (
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">Change:</span>
              <span
                className={`text-sm font-medium ${
                  change >= 0 ? "text-green-600" : "text-red-600"
                }`}
              >
                ${change.toFixed(2)}
              </span>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
};

interface DishTotal {
  quantity: number;
  price: number;
}

interface DishDeliveryTableProps {
  dishTotals: Map<string, DishTotal>;
  deliveryState: Map<string, number>;
  handleDeliveryUpdate: (
    dishName: string
  ) => (newValue: number) => Promise<void>;
}

export const DishDeliveryTable: React.FC<DishDeliveryTableProps> = ({
  dishTotals,
  deliveryState,
  handleDeliveryUpdate
}) => {
  return (
    <div className="w-full">
      <div className="grid grid-cols-4 gap-4 p-4  rounded-t-lg">
        <div className="font-medium text-gray-600">Dish Details</div>
        <div className="font-medium text-gray-600">Delivered</div>
        <div className="font-medium text-gray-600">Value</div>
        <div className="font-medium text-gray-600">Remaining</div>
      </div>

      <div className="space-y-2">
        {Array.from(dishTotals.entries()).map(
          ([dishName, totals]: [string, DishTotal]) => {
            const delivered = deliveryState.get(dishName) || 0;
            const pricePerUnit = totals.price / totals.quantity;
            const deliveredValue = delivered * pricePerUnit;
            const remainingValue = totals.price - deliveredValue;
            const isComplete = delivered === totals.quantity;

            return (
              <div
                key={dishName}
                className="grid grid-cols-4 gap-4 p-4 e border-b items-center"
              >
                <div className="font-medium text-sm">
                  {dishName} ({totals.quantity} x {pricePerUnit.toFixed(2)}K)
                </div>

                <div className="flex items-center">
                  <NumericKeypadInput
                    value={delivered}
                    onChange={() => {}}
                    onSubmit={handleDeliveryUpdate(dishName)}
                    max={totals.quantity}
                    className="w-16 h-8 text-center text-green-600 bg-white rounded border"
                  />
                </div>

                <div className="text-green-600">
                  {deliveredValue.toFixed(2)}
                </div>

                <div
                  className={isComplete ? "text-green-600" : "text-orange-600"}
                >
                  {remainingValue.toFixed(2)}
                </div>
              </div>
            );
          }
        )}
      </div>
    </div>
  );
};
