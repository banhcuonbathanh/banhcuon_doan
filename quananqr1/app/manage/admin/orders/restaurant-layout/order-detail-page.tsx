import React from "react";
import { OrderDetailedResponse } from "@/schemaValidations/interface/type_order";

interface OrderDetailsPageProps {
  order: OrderDetailedResponse;
}

const OrderDetailsPage: React.FC<OrderDetailsPageProps> = ({ order }) => {
  return (
    <div className="w-full h-full overflow-auto p-1 text-xs bg-gray-50">
      <h1 className="text-sm font-bold mb-2 text-blue-800">Order Details</h1>

      {/* Main Order Information */}
      <div className="bg-white shadow-sm rounded-md p-2 mb-2">
        <h2 className="text-xs font-semibold mb-1 border-b pb-1">
          Order Summary
        </h2>
        <div className="grid grid-cols-2 gap-1">
          <div>
            <p className="truncate">
              <strong>Name:</strong> {order.order_name}
            </p>
            <p>
              <strong>Total:</strong> ${order.total_price.toFixed(2)}
            </p>
          </div>
        </div>
      </div>

      {/* Order Sets */}
      <div className="bg-white shadow-sm rounded-md p-2 mb-2">
        <h2 className="text-xs font-semibold mb-1 border-b pb-1">Order Sets</h2>
        {order.data_set.map((set, index) => (
          <div key={set.id} className="mb-1 pb-1 border-b last:border-b-0">
            <h3 className="font-bold text-xs mb-1">
              Set {index + 1}: {set.name}
            </h3>
            <div className="grid grid-cols-2 gap-1">
              <div>
                <p className="truncate">
                  <strong>Desc:</strong> {set.description}
                </p>
                <p>
                  <strong>Price:</strong> ${set.price.toFixed(2)}
                </p>
                <p>
                  <strong>Qty:</strong> {set.quantity}
                </p>
              </div>
            </div>

            {/* Dishes in the Set */}
            <div className="mt-1">
              <table className="w-full border-collapse text-xs">
                <thead>
                  <tr className="bg-gray-100">
                    <th className="border p-1">Dish</th>
                    <th className="border p-1">Price</th>
                    <th className="border p-1">Qty</th>
                  </tr>
                </thead>
                <tbody>
                  {set.dishes.map((dish) => (
                    <tr key={dish.dish_id} className="hover:bg-gray-50">
                      <td className="border p-1 truncate">{dish.name}</td>
                      <td className="border p-1">${dish.price.toFixed(2)}</td>
                      <td className="border p-1">{dish.quantity}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        ))}
      </div>

      {/* Additional Order Information */}
      <div className="bg-white shadow-sm rounded-md p-2">
        <h2 className="text-xs font-semibold mb-1 border-b pb-1">
          Additional Info
        </h2>
        <div className="grid grid-cols-2 gap-1">
          <div></div>
          <div>
            <p>
              <strong>Chili:</strong> {order.chiliNumber}
            </p>
            <p>
              <strong>Topping:</strong> {order.topping}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default OrderDetailsPage;
