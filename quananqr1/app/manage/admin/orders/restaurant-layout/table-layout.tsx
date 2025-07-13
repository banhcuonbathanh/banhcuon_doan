"use client";

import React, { useMemo } from "react";
import { OrderDetailedResponse } from "@/schemaValidations/interface/type_order";
import OrderDetailsPage from "./order-detail-page";

interface RectangleProps {
  number: number;
  description?: string;
  secondDescription?: string;
  order?: OrderDetailedResponse | null;
  className?: string;
}

export const Rectangle: React.FC<RectangleProps> = ({
  number,
  description,
  secondDescription,
  order,
  className = ""
}) => (
  <div
    className={`
      border-4 ${
        order ? "border-green-500 bg-green-50" : "border-blue-500 bg-white"
      }
      p-4 
      relative 
      rounded-lg 
      w-32 h-48
      shadow-md 
      ${className}
    `}
  >
    {/* Top Left Number */}
    <div className="absolute top-2 left-2 text-xl font-bold text-blue-700">
      {number}
    </div>

    <div className="flex flex-col items-center justify-center h-full">
      {description && (
        <div className="text-sm text-gray-600 mt-2 text-center">
          {description}
        </div>
      )}

      {secondDescription && (
        <div className="text-sm text-gray-500 mt-2 text-center italic">
          {secondDescription}
        </div>
      )}

      {/* Order Information */}
      {order && <OrderDetailsPage order={order} />}
    </div>
  </div>
);
