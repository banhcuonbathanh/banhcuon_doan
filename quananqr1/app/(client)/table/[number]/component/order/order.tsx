"use client";

import React, { useEffect } from "react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import useOrderStore from "@/zusstand/order/order_zustand";
import OrderDetails from "../total-dishes-detail";
import OrderCreationComponent from "./add_order_button";
import { decodeTableToken } from "@/lib/utils";

interface OrderProps {
  number: string;
  token: string;
}

export default function OrderSummary({ number, token }: OrderProps) {
  const decoded = decodeTableToken(token);

  console.log(
    "quananqr1/app/(client)/table/[number]/component/order/order.tsx table decode",
    decoded,
    token
  );
  const {
    addTableNumber,
    addTableToken,
    getOrderSummary,
    canhKhongRau,
    canhCoRau,
    smallBowl,
    wantChili,
    selectedFilling,
    updateCanhKhongRau,
    updateCanhCoRau,
    updateSmallBowl,
    updateWantChili,
    updateSelectedFilling
  } = useOrderStore();

  useEffect(() => {
    if (token) {
      addTableToken(token);
    }
    if (number) {
      const tableNumber = addTableNumberconvert(number);
      addTableNumber(tableNumber);
    }
  }, [token, addTableToken, number, addTableNumber]);

  const handleBowlChange = (
    type: "chili" | "noChili" | "small",
    change: number
  ) => {
    switch (type) {
      case "chili":
        const newToppingValue = canhKhongRau + change;
        if (newToppingValue >= 0) updateCanhKhongRau(newToppingValue);
        break;
      case "noChili":
        const newNoChiliValue = canhCoRau + change;
        if (newNoChiliValue >= 0) updateCanhCoRau(newNoChiliValue);
        break;
      case "small":
        const newSmallBowlValue = smallBowl + change;
        if (newSmallBowlValue >= 0) updateSmallBowl(newSmallBowlValue);
        break;
    }
  };

  const orderSummary = getOrderSummary();

  return (
    <div className="container mx-auto px-4 py-5 space-y-5">
      <Card>
        <CardHeader>
          <CardTitle className="flex justify-between items-center">
            Canh Banh Cuon
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-4">
            <div className="space-y-3">
              {/* Bowl without vegetables */}
              <div className="flex items-center justify-between">
                <span>Canh không rau</span>
                <div className="flex items-center gap-2">
                  <Button
                    size="sm"
                    onClick={() => handleBowlChange("chili", -1)}
                    disabled={canhKhongRau === 0}
                  >
                    -
                  </Button>
                  <span className="mx-2 min-w-[2rem] text-center">
                    {canhKhongRau}
                  </span>
                  <Button
                    size="sm"
                    onClick={() => handleBowlChange("chili", 1)}
                  >
                    +
                  </Button>
                </div>
              </div>

              {/* Bowl with vegetables */}
              <div className="flex items-center justify-between">
                <span>Canh rau</span>
                <div className="flex items-center gap-2">
                  <Button
                    size="sm"
                    onClick={() => handleBowlChange("noChili", -1)}
                    disabled={canhCoRau === 0}
                  >
                    -
                  </Button>
                  <span className="mx-2 min-w-[2rem] text-center">
                    {canhCoRau}
                  </span>
                  <Button
                    size="sm"
                    onClick={() => handleBowlChange("noChili", 1)}
                  >
                    +
                  </Button>
                </div>
              </div>

              {/* Small bowl */}
              <div className="flex items-center justify-between">
                <span>Bát bé</span>
                <div className="flex items-center gap-2">
                  <Button
                    size="sm"
                    onClick={() => handleBowlChange("small", -1)}
                    disabled={smallBowl === 0}
                  >
                    -
                  </Button>
                  <span className="mx-2 min-w-[2rem] text-center">
                    {smallBowl}
                  </span>
                  <Button
                    size="sm"
                    onClick={() => handleBowlChange("small", 1)}
                  >
                    +
                  </Button>
                </div>
              </div>

              {/* Chili option */}
              <div className="flex items-center justify-between">
                <span>Có ớt</span>
                <div className="flex items-center gap-2">
                  <Button
                    size="sm"
                    variant={wantChili ? "default" : "outline"}
                    onClick={() => updateWantChili(!wantChili)}
                  >
                    {wantChili ? "Selected" : "Select"}
                  </Button>
                </div>
              </div>

              {/* Nhân mọc nhĩ */}
              <div className="flex items-center justify-between">
                <span>Nhân mọc nhĩ</span>
                <div className="flex items-center gap-2">
                  <Button
                    size="sm"
                    variant={selectedFilling.mocNhi ? "default" : "outline"}
                    onClick={() => updateSelectedFilling("mocNhi")}
                  >
                    {selectedFilling.mocNhi ? "Selected" : "Select"}
                  </Button>
                </div>
              </div>

              {/* Nhân thịt */}
              <div className="flex items-center justify-between">
                <span>Nhân thịt</span>
                <div className="flex items-center gap-2">
                  <Button
                    size="sm"
                    variant={selectedFilling.thit ? "default" : "outline"}
                    onClick={() => updateSelectedFilling("thit")}
                  >
                    {selectedFilling.thit ? "Selected" : "Select"}
                  </Button>
                </div>
              </div>

              {/* Nhân thịt và mọc nhĩ */}
              <div className="flex items-center justify-between">
                <span>Nhân thịt và mọc nhĩ</span>
                <div className="flex items-center gap-2">
                  <Button
                    size="sm"
                    variant={selectedFilling.thitMocNhi ? "default" : "outline"}
                    onClick={() => updateSelectedFilling("thitMocNhi")}
                  >
                    {selectedFilling.thitMocNhi ? "Selected" : "Select"}
                  </Button>
                </div>
              </div>

              {/* Total summary */}
            </div>
          </div>
        </CardContent>
      </Card>

      <OrderDetails
        dishes={orderSummary.dishes}
        sets={orderSummary.sets}
        totalPrice={orderSummary.totalPrice}
        totalItems={orderSummary.totalItems}
      />

      <OrderCreationComponent table_token={token} table_number={number} />
    </div>
  );
}

function addTableNumberconvert(value: string): number {
  let tableNumber: number;

  if (typeof value === "string") {
    if (/^\d+$/.test(value)) {
      tableNumber = parseInt(value, 10);
    } else {
      throw new Error("Invalid input: expected a string of digits.");
    }
  } else if (typeof value === "number") {
    tableNumber = value;
  } else {
    throw new Error("Invalid input: expected a string or number.");
  }

  return tableNumber;
}
