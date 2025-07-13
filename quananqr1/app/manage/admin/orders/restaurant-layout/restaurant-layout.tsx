"use client";

import React, { useMemo } from "react";
import { OrderDetailedResponse } from "@/schemaValidations/interface/type_order";
import { Rectangle } from "./table-layout";


interface RestaurantLayoutProps {
  restaurantLayoutProps: OrderDetailedResponse[];
}


export const RestaurantLayout: React.FC<RestaurantLayoutProps> = ({
  restaurantLayoutProps
}) => {
  // Create a map of table numbers to orders
  const ordersByTableNumber = useMemo(() => {
    return restaurantLayoutProps.reduce((acc, order) => {
      if (order.table_number) {
        acc[order.table_number] = order;
      }
      return acc;
    }, {} as Record<number, OrderDetailedResponse>);
  }, [restaurantLayoutProps]);

  // Define table configurations with their descriptions
  const tableConfigurations = [
    {
      number: 11,
      description: "Window Seat",
      secondDescription: "Seats 4 people"
    },
    { number: 12, description: "Booth", secondDescription: "Cozy corner" },
    { number: 21, description: "High Top", secondDescription: "Near bar area" },
    {
      number: 22,
      description: "Round Table",
      secondDescription: "Group friendly"
    },
    {
      number: 31,
      description: "Patio Seating",
      secondDescription: "Outdoor view"
    },
    {
      number: 32,
      description: "Private Nook",
      secondDescription: "Quiet section"
    },
    {
      number: 17,
      description: "Window Table",
      secondDescription: "Natural light"
    },
    {
      number: 18,
      description: "Central Area",
      secondDescription: "Main dining"
    },
    {
      number: 19,
      description: "Intimate Table",
      secondDescription: "Romantic setting"
    },
    {
      number: 20,
      description: "Family Table",
      secondDescription: "Kid-friendly"
    },
    {
      number: 21,
      description: "Bar Adjacent",
      secondDescription: "Quick service"
    },
    {
      number: 22,
      description: "High Visibility",
      secondDescription: "Near entrance"
    },
    {
      number: 23,
      description: "Quiet Corner",
      secondDescription: "Low traffic"
    },
    {
      number: 24,
      description: "Staff Preferred",
      secondDescription: "Easy access"
    }
  ];

  // Split tables into two columns
  const firstColumnTables = tableConfigurations.slice(0, 7);
  const secondColumnTables = tableConfigurations.slice(7);

  const renderTableColumn = (tables: typeof tableConfigurations) => (
    <div className="flex flex-col border-r-2 border-gray-300 pr-8">
      {[0, 1, 2, 3, 4, 5].map((rowIndex) => (
        <div key={rowIndex} className={`flex ${rowIndex > 0 ? "mt-20" : ""}`}>
          {tables.slice(rowIndex * 2, rowIndex * 2 + 2).map((table) => (
            <Rectangle
              key={table.number}
              number={table.number}
              description={table.description}
              secondDescription={table.secondDescription}
              order={ordersByTableNumber[table.number] || null}
            />
          ))}
        </div>
      ))}
    </div>
  );

  return (
    <div className="flex bg-gray-100 p-6 space-x-8">
      {renderTableColumn(firstColumnTables)}
      {renderTableColumn(secondColumnTables)}
    </div>
  );
};

export default RestaurantLayout;
