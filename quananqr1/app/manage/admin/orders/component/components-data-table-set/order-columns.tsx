"use client";

import React, { useEffect, useState } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from "@/components/ui/select";
import {
  OrderDetailedDish,
  OrderDetailedResponse,
  OrderSetDetailed
} from "@/schemaValidations/interface/type_order";

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

export const columns: ColumnDef<OrderDetailedResponse, any>[] = [
  {
    accessorKey: "order_info",
    header: "Order Information",
    cell: ({ row, table }) => {
      const orderName = row.original.order_name.split("_")[0];
      const tableNumber = row.original.table_number;
      const [selectedStatus, setSelectedStatus] = useState<OrderStatus>(
        row.original.status as OrderStatus
      );
      const withChili = row.original.bow_chili;
      const noChili = row.original.bow_no_chili;
      const total = withChili + noChili;
      const isTakeAway = row.original.takeAway;
      const chiliNumber = row.original.chiliNumber;

      const statusStyles: Record<OrderStatus, string> = {
        ORDERING: "bg-blue-100 text-blue-800",
        SERVING: "bg-yellow-100 text-yellow-800",
        WAITING: "bg-orange-100 text-orange-800",
        DONE: "bg-green-100 text-green-800"
      };

      const meta = table.options.meta as TableMeta;

      return (
        <div className="space-y-3">
          {/* Name Section */}
          <div className="flex flex-col">
            <span className="text-sm font-medium text-gray-600">Name</span>
            <span className="font-medium mt-1">{orderName}</span>
          </div>

          {/* Table Section */}
          <div className="flex flex-col">
            <span className="text-sm font-medium text-gray-600">Table</span>
            <div
              className={`mt-1 ${
                isTakeAway
                  ? "bg-orange-600 text-white rounded-md px-2 py-1 inline-block"
                  : ""
              }`}
            >
              {tableNumber}
            </div>
          </div>

          {/* Status Section */}
          <div className="flex flex-col">
            <span className="text-sm font-medium text-gray-600">Status</span>
            <div className="mt-1">
              <Select
                value={selectedStatus}
                onValueChange={(newStatus: OrderStatus) => {
                  setSelectedStatus(newStatus);
                  meta?.onStatusChange?.(row.original.id, newStatus);
                }}
              >
                <SelectTrigger
                  className={`w-[120px] h-8 ${statusStyles[selectedStatus]}`}
                >
                  <SelectValue>{selectedStatus}</SelectValue>
                </SelectTrigger>
                <SelectContent>
                  {ORDER_STATUSES.map((orderStatus) => (
                    <SelectItem
                      key={orderStatus}
                      value={orderStatus}
                      className={statusStyles[orderStatus]}
                    >
                      {orderStatus}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
        </div>
      );
    }
  },

  {
    accessorKey: "combined_orders",
    header: "Order Details",
    cell: ({ row }) => {
      const sets = row.original.data_set as OrderSetDetailed[];
      const dishes = row.original.data_dish as OrderDetailedDish[];

      return (
        <div className="space-y-4">
          {/* Sets Section */}
          {sets && sets.length > 0 && (
            <div>
              <div className="text-sm font-semibold text-gray-700 mb-2">
                Sets
              </div>
              <div className="space-y-2">
                {sets.map((set) => (
                  <div
                    key={set.id}
                    className="border-b border-gray-100 pb-2 last:border-0"
                  >
                    <div className="text-sm font-medium">
                      {set.quantity}x {set.name} (${set.price})
                    </div>
                    <div className="pl-4 text-sm text-gray-600">
                      {set.dishes.map((dish, index) => (
                        <div key={`${dish.dish_id}-${index}`}>
                          {set.quantity} x {dish.quantity} ={" "}
                          {set.quantity * dish.quantity} {dish.name}
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Individual Dishes Section */}
          {dishes && dishes.length > 0 && (
            <div>
              <div className="text-sm font-semibold text-gray-700 mb-2">
                Individual Dishes
              </div>
              <div className="space-y-1">
                {dishes.map((dish, index) => (
                  <div key={`${dish.dish_id}-${index}`} className="text-sm">
                    {dish.quantity}x {dish.name} (${dish.price})
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      );
    }
  },

  {
    accessorKey: "order_tracking",
    header: "Order Tracking",
    cell: ({ row, table }) => {
      const meta = table.options.meta as TableMeta;
      const [deliveryState, setDeliveryState] = useState<Map<string, number>>(
        new Map()
      );

      useEffect(() => {
        if (row.original.deliveryData) {
          setDeliveryState(new Map(Object.entries(row.original.deliveryData)));
        }
      }, [row.original.deliveryData]);

      // Calculate totals from sets and individual dishes
      const calculateDishTotals = () => {
        const dishTotals = new Map<string, number>();

        // Calculate totals from sets
        row.original.data_set?.forEach((set) => {
          set.dishes.forEach((dish) => {
            const totalQuantity = set.quantity * dish.quantity;
            const currentTotal = dishTotals.get(dish.name) || 0;
            dishTotals.set(dish.name, currentTotal + totalQuantity);
          });
        });

        // Add individual dishes to totals
        row.original.data_dish?.forEach((dish) => {
          const currentTotal = dishTotals.get(dish.name) || 0;
          dishTotals.set(dish.name, currentTotal + dish.quantity);
        });

        return dishTotals;
      };

      const handleDeliveryUpdate =
        (dishName: string) => (e: React.ChangeEvent<HTMLInputElement>) => {
          const dishTotals = calculateDishTotals();
          const totalQuantity = dishTotals.get(dishName) || 0;
          const newDelivered = Math.min(
            parseInt(e.target.value) || 0,
            totalQuantity
          );

          const newState = new Map(deliveryState);
          newState.set(dishName, newDelivered);
          setDeliveryState(newState);

          if (meta?.onDeliveryUpdate) {
            meta.onDeliveryUpdate(row.original.id, dishName, newDelivered);
          }
        };

      const dishTotals = calculateDishTotals();

      return (
        <div className="space-y-2">
          {/* Header */}
          <div className="grid grid-cols-4 gap-2 px-2 py-1  rounded-t text-sm font-medium">
            <div className="col-span-1">Dish</div>
            <div className="text-center">Total</div>
            <div className="text-center text-green-600">Delivered</div>
            <div className="text-center text-orange-600">Remaining</div>
          </div>

          {/* Dish Rows */}
          {Array.from(dishTotals.entries()).map(([dishName, totalQuantity]) => {
            const delivered = deliveryState.get(dishName) || 0;
            const remaining = totalQuantity - delivered;
            const isComplete = remaining === 0;

            return (
              <div
                key={dishName}
                className={`grid grid-cols-4 gap-2 items-center py-1 border-b border-gray-100 last:border-0 ${
                  isComplete ? "bg-green-50" : ""
                }`}
              >
                {/* Dish Name */}
                <div className="col-span-1 text-sm font-medium">{dishName}</div>

                {/* Total */}
                <div className="text-center font-medium">{totalQuantity}</div>

                {/* Delivered */}
                <div className="flex justify-center">
                  <Input
                    type="number"
                    value={delivered}
                    onChange={handleDeliveryUpdate(dishName)}
                    className="w-16 h-7 text-center text-green-600 font-medium"
                    min="0"
                    max={totalQuantity}
                  />
                </div>

                {/* Remaining */}
                <div className="text-center">
                  <span
                    className={`font-medium ${
                      isComplete ? "text-green-600" : "text-orange-600"
                    }`}
                  >
                    {remaining}
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      );
    }
  },
  {
    accessorKey: "bow_details",
    header: "Bowl Details",
    cell: ({ row }) => {
      const withChili = row.original.bow_chili;
      const noChili = row.original.bow_no_chili;
      const total = withChili + noChili;
      const isTakeAway = row.original.takeAway;
      const chiliNumber = row.original.chiliNumber;

      return total > 0 || (isTakeAway && chiliNumber > 0) ? (
        <div className="space-y-1 text-sm">
          {withChili > 0 && <div>With Chili: {withChili}</div>}
          {noChili > 0 && <div>No Chili: {noChili}</div>}
          {isTakeAway && chiliNumber > 0 && (
            <div className="font-medium">Takeaway Chili: {chiliNumber}</div>
          )}
        </div>
      ) : null;
    }
  },

  //--------

  {
    accessorKey: "payment_info",
    header: "Payment Information",
    cell: ({ row, table }) => {
      const [selectedPayment, setSelectedPayment] =
        useState<PaymentMethod>("CASH");
      const totalPrice = row.original.total_price as number;
      const [amountPaid, setAmountPaid] = useState<string>("");
      const [change, setChange] = useState<number | null>(null);

      const paymentStyles: Record<PaymentMethod, string> = {
        CASH: "bg-emerald-50 text-emerald-700",
        TRANSFER: "bg-indigo-50 text-indigo-700"
      };

      const meta = table.options.meta as TableMeta;

      const handlePaymentInput = (value: string) => {
        setAmountPaid(value);
        const numericValue = parseFloat(value) || 0;
        const changeAmount = numericValue - totalPrice;
        setChange(changeAmount >= 0 ? changeAmount : null);
      };

      return (
        <div className="space-y-3">
          {/* Payment Method Section */}
          <div className="flex flex-col">
            <span className="text-sm font-medium text-gray-600">
              Payment Method
            </span>
            <div className="mt-1">
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

          {/* Total Amount Section */}
          <div className="flex flex-col">
            <span className="text-sm font-medium text-gray-600">
              Total Amount
            </span>
            <span className="font-medium mt-1">${totalPrice}</span>
          </div>

          {/* Payment Details Section */}
          <div className="flex flex-col">
            <span className="text-sm font-medium text-gray-600">
              Amount Paid
            </span>
            <div className="mt-1 space-y-2">
              <div className="flex items-center gap-2">
                <Input
                  type="number"
                  placeholder="Amount paid"
                  value={amountPaid}
                  onChange={(e) => handlePaymentInput(e.target.value)}
                  className="w-24 h-8 text-right"
                />
                <span className="text-sm">$</span>
              </div>
              {change !== null && (
                <div className="flex flex-col">
                  <div
                    className={`text-sm ${
                      change >= 0 ? "text-green-600" : "text-red-600"
                    }`}
                  >
                    Change: ${change.toFixed(2)}
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      );
    }
  }
];
