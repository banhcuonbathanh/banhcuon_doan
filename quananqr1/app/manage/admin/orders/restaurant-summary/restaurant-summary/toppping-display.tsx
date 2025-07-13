import { OrderDetailedResponse } from "@/schemaValidations/interface/type_order";
import React from "react";


interface ToppingDisplayProps {
  orders: OrderDetailedResponse[];
}

const FirstOrderToppings: React.FC<ToppingDisplayProps> = ({ orders }) => {
  const parseToppings = (
    toppingString: string
  ): Array<{ key: string; value: string }> => {
    return toppingString
      .split("-")
      .map((item) => item.trim())
      .filter((item) => item.length > 0)
      .map((item) => {
        // Handle special case for "ot tuoi"
        if (item.includes("ot tuoi")) {
          const isTrue = item.includes("true");
          return {
            key: "ot tuoi",
            value: isTrue ? "co" : "khong"
          };
        }

        // Remove any existing colons from the input
        item = item.replace(/:/g, "").trim();

        // Handle numeric values (like "canhKhongRau 2" or "bat be 2")
        if (item.match(/.*\s\d+$/)) {
          const lastSpace = item.lastIndexOf(" ");
          return {
            key: item.substring(0, lastSpace).trim(),
            value: item.substring(lastSpace + 1).trim()
          };
        }

        // Handle text values (like "nhan Thịt Mọc Nhĩ")
        const firstSpace = item.indexOf(" ");
        if (firstSpace !== -1) {
          return {
            key: item.substring(0, firstSpace).trim(),
            value: item.substring(firstSpace + 1).trim()
          };
        }

        return {
          key: item,
          value: ""
        };
      });
  };

  if (orders.length === 0) {
    return (
      <div className="rounded-lg shadow-sm p-4">
        <div className="text-gray-500 text-center py-4">
          No orders available
        </div>
      </div>
    );
  }

  const firstOrder = orders[0];
  const toppings = firstOrder.topping ? parseToppings(firstOrder.topping) : [];

  return (
    <div className="rounded-lg shadow-sm p-4">
      <div className="space-y-4">
        <div className="space-y-2">
          {toppings.map((topping, index) => (
            <div key={index} className="flex items-center p-2 rounded-md">
              <span className="font-medium">
                {`${topping.key}: ${topping.value}`}
              </span>
            </div>
          ))}
        </div>

        {toppings.length === 0 && (
          <div className="text-gray-500 text-center py-4">
            No toppings found in this order
          </div>
        )}
      </div>
    </div>
  );
};

export default FirstOrderToppings;
