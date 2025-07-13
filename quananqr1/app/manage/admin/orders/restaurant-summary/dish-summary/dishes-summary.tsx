"use client";

import React, { useState } from "react";
import { Dialog, DialogContent } from "@/components/ui/dialog";
import { toast } from "@/components/ui/use-toast";

import NumericKeypad from "./num-pad";
import { logWithLevel } from "@/lib/log";
import {
  OrderDetailedResponse,
  Guest
} from "@/schemaValidations/interface/type_order";
import { AggregatedDish } from "./aggregateDishes";
import useDeliveryStore from "@/zusstand/delivery/delivery_zustand";

const LOG_PATH =
  "quananqr1/app/manage/admin/orders/restaurant-summary/dishes-summary.tsx";

interface DishSummaryProps {
  dish: AggregatedDish;
  orders: OrderDetailedResponse[];
}

export const DishSummary: React.FC<DishSummaryProps> = ({ dish, orders }) => {
  const [showDetails, setShowDetails] = useState(false);
  const [showKeypad, setShowKeypad] = useState(false);
  const [deliveryNumber, setDeliveryNumber] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const createDelivery = useDeliveryStore((state) => state.createDelivery);

  // Log component initialization
  React.useEffect(() => {
    logWithLevel(
      {
        action: "component_mounted",
        dishName: dish.name,
        quantity: dish.quantity,
        ordersCount: orders.length
      },
      LOG_PATH,
      "debug",
      1
    );
  }, [dish.name, dish.quantity, orders.length]);

  const handleDeliveryClick = () => {
    logWithLevel(
      {
        action: "delivery_click",
        dishName: dish.name,
        currentDeliveryNumber: deliveryNumber,
        maxQuantity: dish.quantity,
        timestamp: new Date().toISOString()
      },
      LOG_PATH,
      "info",
      1
    );
    setShowKeypad(true);
  };

  const handleKeypadSubmit = async () => {
    logWithLevel(
      {
        action: "keypad_submit_attempt",
        dishName: dish.name,
        deliveryNumber,
        maxQuantity: dish.quantity,
        timestamp: new Date().toISOString()
      },
      LOG_PATH,
      "info",
      2
    );

    if (deliveryNumber > dish.quantity) {
      toast({
        variant: "destructive",
        title: "Error",
        description: "Delivery quantity cannot exceed available quantity"
      });
      logWithLevel(
        {
          error: "Invalid delivery number",
          deliveryNumber,
          maxQuantity: dish.quantity,
          dishName: dish.name,
          timestamp: new Date().toISOString()
        },
        LOG_PATH,
        "error",
        3
      );
      return;
    }

    setIsSubmitting(true);

    try {
      // Get the first order as reference for delivery details
      const referenceOrder = orders[0];
      if (!referenceOrder) {
        logWithLevel(
          {
            error: "No reference order found",
            ordersLength: orders.length,
            dishName: dish.name,
            timestamp: new Date().toISOString()
          },
          LOG_PATH,
          "error",
          6
        );
        throw new Error("No order reference found");
      }

      // Log reference order details
      logWithLevel(
        {
          action: "reference_order_selected",
          orderId: referenceOrder.id,
          isGuest: referenceOrder.is_guest,
          tableNumber: referenceOrder.table_number,
          timestamp: new Date().toISOString()
        },
        LOG_PATH,
        "info",
        6
      );

      // Create a guest object if the order is from a guest
      const guest: Guest | null = referenceOrder.is_guest
        ? {
            id: referenceOrder.guest_id,
            name: referenceOrder.order_name,
            table_number: referenceOrder.table_number,
            created_at: referenceOrder.created_at,
            updated_at: referenceOrder.updated_at
          }
        : null;

      // Log guest order processing
      if (referenceOrder.is_guest) {
        logWithLevel(
          {
            action: "guest_order_processing",
            guestId: referenceOrder.guest_id,
            orderName: referenceOrder.order_name,
            tableNumber: referenceOrder.table_number,
            timestamp: new Date().toISOString()
          },
          LOG_PATH,
          "info",
          7
        );
      }

      // Create order store with data from the reference order
      const orderStore = {
        tableNumber: referenceOrder.table_number,
        getOrderSummary: () => ({
          dishes: [
            {
              id: dish.dish_id,
              quantity: deliveryNumber,
              name: dish.name,
              price: dish.price
            }
          ],
          orderId: referenceOrder.id
        }),
        clearOrder: () => {}
      };

      // Log order store creation
      logWithLevel(
        {
          action: "order_store_created",
          tableNumber: orderStore.tableNumber,
          dishDetails: {
            id: dish.dish_id,
            name: dish.name,
            quantity: deliveryNumber
          },
          timestamp: new Date().toISOString()
        },
        LOG_PATH,
        "info",
        8
      );

      // Create delivery details from order data
      const deliveryDetails = {
        deliveryAddress: "",
        deliveryContact: "",
        deliveryNotes: `Delivery for ${dish.name} - Quantity: ${deliveryNumber} - Order #${referenceOrder.id}`,
        scheduledTime: new Date().toISOString(),
        deliveryFee: 0
      };

      // Log delivery details preparation
      logWithLevel(
        {
          action: "delivery_details_prepared",
          details: deliveryDetails,
          timestamp: new Date().toISOString()
        },
        LOG_PATH,
        "info",
        8
      );

      const response = await createDelivery({
        guest,
        user: referenceOrder.is_guest
          ? null
          : {
              id: referenceOrder.user_id,
              name: referenceOrder.order_name
            },
        isGuest: referenceOrder.is_guest,
        orderStore,
        deliveryDetails
      });

      logWithLevel(
        {
          action: "delivery_submitted",
          dishName: dish.name,
          deliveryNumber,
          success: true,
          orderId: referenceOrder.id,
          response: response
        },
        LOG_PATH,
        "info",
        4
      );

      toast({
        title: "Success",
        description: `Created delivery for ${deliveryNumber} ${dish.name}`
      });

      setShowKeypad(false);
      setDeliveryNumber(0);
    } catch (error) {
      logWithLevel(
        {
          action: "delivery_submission_failed",
          dishName: dish.name,
          deliveryNumber,
          error: error instanceof Error ? error.message : "Unknown error",
          timestamp: new Date().toISOString()
        },
        LOG_PATH,
        "error",
        5
      );

      toast({
        variant: "destructive",
        title: "Error",
        description:
          error instanceof Error ? error.message : "Failed to create delivery"
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="p-2 mb-2">
      <div className="flex items-center justify-between">
        <div
          className="flex-1 cursor-pointer"
          onClick={() => {
            setShowDetails(!showDetails);
            // Log details toggle
            logWithLevel(
              {
                action: "details_toggle",
                dishName: dish.name,
                showDetails: !showDetails,
                timestamp: new Date().toISOString()
              },
              LOG_PATH,
              "debug",
              1
            );
          }}
        >
          <span className="font-bold">
            {dish.name} :{dish.quantity} -
          </span>
          <span
            className="text-blue-600 cursor-pointer hover:text-blue-800"
            onClick={(e) => {
              e.stopPropagation();
              handleDeliveryClick();
            }}
          >
            {deliveryNumber > 0 ? `delivery (${deliveryNumber})` : "delivery"}
          </span>
        </div>
      </div>

      {showDetails && (
        <div className="mt-2 pl-4 text-gray-600">
          <div className="grid grid-cols-2 gap-1">
            <div className="font-medium">Price per Unit:</div>
            <div>${dish.price.toFixed(2)}</div>
            <div className="font-medium">Total Price:</div>
            <div>${(dish.price * dish.quantity).toFixed(2)}</div>
          </div>
        </div>
      )}

      <Dialog
        open={showKeypad}
        onOpenChange={(open) => {
          setShowKeypad(open);
          // Log keypad dialog state change
          logWithLevel(
            {
              action: "keypad_dialog_toggle",
              dishName: dish.name,
              showKeypad: open,
              timestamp: new Date().toISOString()
            },
            LOG_PATH,
            "debug",
            1
          );
        }}
      >
        <DialogContent className="sm:max-w-md">
          <div className="py-4">
            <h2 className="text-lg font-semibold mb-4 text-center">
              Enter Delivery Number for {dish.name}
            </h2>
            <NumericKeypad
              value={deliveryNumber}
              onChange={(value) => {
                setDeliveryNumber(value);
                // Log delivery number change
                logWithLevel(
                  {
                    action: "delivery_number_change",
                    dishName: dish.name,
                    newValue: value,
                    timestamp: new Date().toISOString()
                  },
                  LOG_PATH,
                  "debug",
                  1
                );
              }}
              onSubmit={handleKeypadSubmit}
              min={0}
              max={dish.quantity}
              className="w-full"
              disabled={isSubmitting}
            />
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default DishSummary;
