"use client";

import React, { useState } from "react";
import { ColumnDef } from "@tanstack/react-table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from "@/components/ui/select";

const ORDER_STATUSES = ["ORDERING", "SERVING", "WAITING", "DONE"] as const;
type OrderStatus = (typeof ORDER_STATUSES)[number];

const PAYMENT_METHODS = ["CASH", "TRANSFER"] as const;
type PaymentMethod = (typeof PAYMENT_METHODS)[number];

interface TableMeta {
  onStatusChange?: (orderId: number, newStatus: string) => void;
  onPaymentMethodChange?: (orderId: number, newMethod: string) => void;
  onDeliveryUpdate?: (
    orderId: number,
    dishName: string,
    deliveredQuantity: number
  ) => void;
}

const OrderHeader = ({ row, meta }: { row: any; meta: TableMeta }) => {
  const orderName = row.original.order_name.split("_")[0];
  const tableNumber = row.original.table_number;
  const [selectedStatus, setSelectedStatus] = useState<OrderStatus>(
    row.original.status as OrderStatus
  );
  const [selectedPayment, setSelectedPayment] = useState<PaymentMethod>("CASH");
  const isTakeAway = row.original.takeAway;
  const [isCalculationsVisible, setIsCalculationsVisible] = useState(false);

  // Bowl details
  const withChili = row.original.bow_chili || 0;
  const noChili = row.original.bow_no_chili || 0;
  const totalBowls = withChili + noChili;
  const chiliNumber = row.original.chiliNumber || 0;

  const paymentStyles: Record<PaymentMethod, string> = {
    CASH: "bg-emerald-50 text-emerald-700",
    TRANSFER: "bg-indigo-50 text-indigo-700"
  };

  return (
    <div className="flex flex-col rounded-t-lg border-b">
      {/* First Row - Main Order Details */}
      <div className="flex items-center gap-6 p-4 pb-2">
        {/* Name Section */}
        <div className="flex flex-row min-w-[100px]">
          <span className="text-sm font-medium text-gray-600">Name</span>
          <span className="ml-2 text-sm font-medium ">{orderName}</span>
        </div>

        {/* Table Section */}
        <div className="flex flex-row min-w-[100px]">
          <span className="text-sm font-medium ">Table</span>
          <div
            className={`ml-2 text-sm font-medium ${
              isTakeAway
                ? "bg-orange-600 text-white rounded-md px-2 py-1"
                : "text-gray-600"
            }`}
          >
            {tableNumber}
          </div>
        </div>

        {/* Status Section */}
        <div className="flex flex-row ">
          <span className="text-sm font-medium text-gray-600">Status</span>
        </div>

        {/* Payment Method Section */}
        <div className="flex flex-row min-w-[150px]">
          <span className="text-sm font-medium text-gray-600">Payment</span>
          <div className="ml-2">
            <Select
              value={selectedPayment}
              onValueChange={(newMethod: PaymentMethod) => {
                setSelectedPayment(newMethod);
                meta?.onPaymentMethodChange?.(row.original.id, newMethod);
              }}
            >
              <SelectTrigger
                className={`w-[120px] h-8 ${paymentStyles[selectedPayment]}`}
              >
                <SelectValue>
                  {selectedPayment === "CASH" ? "Cash" : "Transfer"}
                </SelectValue>
              </SelectTrigger>
              <SelectContent>
                {PAYMENT_METHODS.map((method) => (
                  <SelectItem
                    key={method}
                    value={method}
                    className={paymentStyles[method]}
                  >
                    {method === "CASH" ? "Cash" : "Transfer"}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
      </div>

      {/* Second Row - Bowl Details */}
      {(totalBowls > 0 || (isTakeAway && chiliNumber > 0)) && (
        <div className="px-4 py-2 flex items-center">
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-1">
              <span className="text-sm font-medium text-gray-600">Bowls:</span>
            </div>
            <div className="flex gap-3">
              {withChili > 0 && (
                <span className="text-sm px-3 py-1 bg-red-50 text-red-600 rounded-md flex items-center gap-1">
                  <span className="text-lg">🥬</span>
                  <span className="font-medium">{withChili}</span>
                </span>
              )}
              {noChili > 0 && (
                <span className="text-sm px-3 py-1 bg-green-50 text-green-600 rounded-md flex items-center gap-1">
                  <span className="text-lg">⛔</span>
                  <span className="font-medium">{noChili}</span>
                </span>
              )}
              {isTakeAway && chiliNumber > 0 && (
                <span className="text-sm px-3 py-1 bg-orange-50 text-orange-600 rounded-md flex items-center gap-1">
                  <span className="text-lg">🏃</span>
                  <span className="font-medium">{chiliNumber}</span>
                </span>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Third Row - Order Details */}
      <div className="px-4 py-2 rounded-b-lg">
        <div className="flex flex-wrap gap-x-8 gap-y-2">
          {/* Sets Section */}
          {row.original.data_set && row.original.data_set.length > 0 && (
            <div className="flex-1 min-w-[300px]">
              <div className="flex items-center gap-2">
                <span className="text-sm font-semibold text-gray-700">
                  Sets:{" "}
                </span>
                <button
                  onClick={() =>
                    setIsCalculationsVisible(!isCalculationsVisible)
                  }
                  className="text-gray-500 hover:text-gray-700"
                >
                  {isCalculationsVisible ? (
                    <ChevronUp className="w-4 h-4" />
                  ) : (
                    <ChevronDown className="w-4 h-4" />
                  )}
                </button>
              </div>
              <div className="mt-1 space-y-2">
                {row.original.data_set.map((set: OrderSetDetailed) => (
                  <div key={set.id}>
                    {/* Always visible set information */}
                    <div className="text-sm text-gray-700">
                      {set.quantity}x {set.name} (${set.price})
                    </div>

                    {/* Togglable detailed calculations */}
                    {isCalculationsVisible && (
                      <div className="pl-4 text-sm space-y-1">
                        {set.dishes.map((dish, index) => (
                          <div
                            key={`${dish.dish_id}-${index}`}
                            className="flex items-center gap-2"
                          >
                            <span className="text-gray-500">{dish.name}</span>
                            <span className="bg-gray-600 text-white px-2 py-1 rounded text-xs">
                              {set.quantity} x {dish.quantity} ={" "}
                              {set.quantity * dish.quantity}
                            </span>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Individual Dishes Section */}
          {row.original.data_dish && row.original.data_dish.length > 0 && (
            <div className="flex-1 min-w-[300px]">
              <span className="text-sm font-semibold text-gray-700">
                Individual Dishes:{" "}
              </span>
              <div className="mt-1 space-y-1">
                {row.original.data_dish.map(
                  (dish: OrderDetailedDish, index: number) => (
                    <div
                      key={`${dish.dish_id}-${index}`}
                      className="text-sm text-gray-700"
                    >
                      {dish.quantity}x {dish.name} (${dish.price})
                    </div>
                  )
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

// Order Tracking Component
// ... (previous code remains the same until OrderTracking component)

// ... (rest of the code remains the same)

// Main Column Definition
export const columns: ColumnDef<OrderDetailedResponse, any>[] = [
  {
    id: "order_details",
    cell: ({ row, table }) => {
      const meta = table.options.meta as TableMeta;

      return (
        <div className="flex flex-col w-full">
          <OrderHeader row={row} meta={meta} />

          <div className="space-y-4">
            <h3 className="font-semibold">Order Tracking</h3>
            <OrderTracking12 row={row} meta={meta} />
          </div>
        </div>
      );
    },
    meta: {
      skipHeaderRender: true
    }
  }
];

// Example usage of the table (you can include this if needed)
const OrderTable = ({ data }: { data: OrderDetailedResponse[] }) => {
  const [rowSelection, setRowSelection] = useState({});

  const handleStatusChange = (orderId: number, newStatus: string) => {
    console.log(`Order ${orderId} status changed to ${newStatus}`);
    // Implement your status change logic here
  };

  const handlePaymentMethodChange = (orderId: number, newMethod: string) => {
    console.log(`Order ${orderId} payment method changed to ${newMethod}`);
    // Implement your payment method change logic here
  };

  const handleDeliveryUpdate = (
    orderId: number,
    dishName: string,
    deliveredQuantity: number
  ) => {
    console.log(
      `Order ${orderId} dish ${dishName} delivery updated to ${deliveredQuantity}`
    );
    // Implement your delivery update logic here
  };

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    onRowSelectionChange: setRowSelection,
    state: {
      rowSelection
    },
    meta: {
      onStatusChange: handleStatusChange,
      onPaymentMethodChange: handlePaymentMethodChange,
      onDeliveryUpdate: handleDeliveryUpdate
    }
  });

  return (
    <div className="rounded-md border">
      <Table>
        <TableBody>
          {table.getRowModel().rows.map((row) => (
            <TableRow key={row.id}>
              {row.getVisibleCells().map((cell) => (
                <TableCell key={cell.id}>
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
};

// Required imports for the table component
import { Table, TableBody, TableCell, TableRow } from "@/components/ui/table";
import {
  flexRender,
  getCoreRowModel,
  useReactTable
} from "@tanstack/react-table";
import {
  OrderDetailedDish,
  OrderDetailedResponse,
  OrderSetDetailed
} from "@/schemaValidations/interface/type_order";
import { ChevronUp, ChevronDown } from "lucide-react";

import {OrderTracking12}  from "./order-tracking";

// Additional types that might be needed
interface OrderTableProps {
  data: OrderDetailedResponse[];
}

// Export the components
export { OrderTable };
export type { OrderTableProps, OrderDetailedResponse, TableMeta };

// You might also want to include these utility types/interfaces if needed elsewhere
export type { OrderStatus, PaymentMethod, OrderSetDetailed, OrderDetailedDish };
