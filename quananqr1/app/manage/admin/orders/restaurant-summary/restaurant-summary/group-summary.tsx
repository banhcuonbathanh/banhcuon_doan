import React, { useState, useMemo } from 'react';
import { ChevronDown } from "lucide-react";
import { OrderDetailedResponse } from '@/schemaValidations/interface/type_order';


interface GroupSummaryProps {
  orders: OrderDetailedResponse[];
}

export const GroupSummary: React.FC<GroupSummaryProps> = ({ orders }) => {
  const [isDetailsVisible, setIsDetailsVisible] = useState(false);
  const totals = useMemo(() => {
    let dishTotal = 0;
    let setTotal = 0;

    orders.forEach((order) => {
      order.data_dish.forEach((dish) => {
        dishTotal += dish.price * dish.quantity;
      });

      order.data_set.forEach((set) => {
        setTotal += set.price * set.quantity;
      });
    });

    return {
      dishTotal,
      setTotal,
      grandTotal: dishTotal + setTotal
    };
  }, [orders]);

  return (
    <div className="mt-4 pt-4 border-t">
      <div
        className="cursor-pointer select-none"
        onClick={() => setIsDetailsVisible(!isDetailsVisible)}
      >
        <div className="grid grid-cols-2 gap-2">
          <div className="font-bold text-lg">Total:</div>
          <div className="text-right font-bold text-lg">
            ${totals.grandTotal.toFixed(2)}
            <ChevronDown
              className={`inline-block ml-2 h-4 w-4 transition-transform duration-200 ${
                isDetailsVisible ? "transform rotate-180" : ""
              }`}
            />
          </div>
        </div>

        {isDetailsVisible && (
          <div className="grid grid-cols-2 gap-2 mt-2 text-sm">
            <div className="font-medium">Individual Dishes:</div>
            <div className="text-right">${totals.dishTotal.toFixed(2)}</div>

            <div className="font-medium">Set Orders:</div>
            <div className="text-right">${totals.setTotal.toFixed(2)}</div>
          </div>
        )}
      </div>
    </div>
  );
};

export default GroupSummary;