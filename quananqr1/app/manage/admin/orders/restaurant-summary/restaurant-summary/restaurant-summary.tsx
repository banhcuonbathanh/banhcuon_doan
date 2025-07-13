import React, { useMemo } from "react";
import { logWithLevel } from "@/lib/log";
import { GroupedOrder, RestaurantSummaryProps } from "./rs-type";
import { aggregateDishes, getOrdinalSuffix, parseOrderName } from "./ultil";
import CollapsibleSection from "./colap-section";

import OrderDetails from "./order-detail";
import GroupSummary from "./group-summary";
import FirstOrderToppings from "./toppping-display";
import { DishSummary } from "../dish-summary/dishes-summary";

export const RestaurantSummary2: React.FC<RestaurantSummaryProps> = ({
  restaurantLayoutProps
}) => {
  const groupedOrders = useMemo(() => {
    const groups = new Map<string, GroupedOrder>();

    restaurantLayoutProps.forEach((order) => {
      const characteristic = parseOrderName(order.order_name);
      const groupKey = `${characteristic}-${order.table_number}`;

      if (!groups.has(groupKey)) {
        groups.set(groupKey, {
          orderName: characteristic,
          tableNumber: order.table_number,
          orders: [],
          hasTakeAway: false
        });
      }
      const group = groups.get(groupKey)!;
      group.orders.push(order);
      if (order.takeAway) {
        group.hasTakeAway = true;
      }
    });

    return Array.from(groups.values());
  }, [restaurantLayoutProps]);

  return (
    <div className="p-4">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {groupedOrders.map((group) => {
          const aggregatedDishes = aggregateDishes(group.orders);

          logWithLevel(
            { group },
            "quananqr1/app/manage/admin/orders/restaurant-summary/restaurant-summary.tsx group",
            "info",
            3
          );
          logWithLevel(
            { aggregatedDishes },
            "quananqr1/app/manage/admin/orders/restaurant-summary/restaurant-summary.tsx",
            "info",
            2
          );
          return (
            <div
              key={`${group.orderName}-${group.tableNumber}`}
              className="shadow-md rounded-lg p-4 border"
            >
              <h3 className="text-xl font-semibold mb-4">
                {group.orderName} - Bàn {group.tableNumber}
                {group.hasTakeAway && (
                  <span className="ml-2 text-red-600">(Đem đi)</span>
                )}
              </h3>

              <div className="rounded-lg shadow-sm p-4">
                <CollapsibleSection title="Canh">
                  <FirstOrderToppings orders={group.orders} />
                </CollapsibleSection>

                <CollapsibleSection title="Món Ăn">
                  {aggregatedDishes.map((dish, index) => (
                    <DishSummary key={`${dish.dish_id}-${index}`} dish={dish} orders={group.orders} />
                  ))}
                </CollapsibleSection>

                <CollapsibleSection title="Lần Gọi Đồ">
                  {group.orders.map((order, index) => (
                    <div key={order.id} className="mb-4 last:mb-0">
                      <div className="font-medium text-lg mb-2">
                        {`${index + 1}${getOrdinalSuffix(index + 1)} Order`}
                      </div>
                      <OrderDetails order={order} />
                    </div>
                  ))}
                </CollapsibleSection>

                <GroupSummary orders={group.orders} />
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default RestaurantSummary2;
