import { OrderDetailedResponse } from '@/schemaValidations/interface/type_order';
import React from 'react';


interface OrderDetailsProps {
  order: OrderDetailedResponse;
}

export const OrderDetails: React.FC<OrderDetailsProps> = ({ order }) => (
  <div className="border-b last:border-b-0 py-4">
    <div className="grid grid-cols-2 gap-2">
      <div className="font-semibold">Table Number:</div>
      <div>{order.table_number}</div>
      <div className="font-semibold">Status:</div>
      <div className={order.takeAway ? "text-red-600 font-bold" : ""}>
        {order.takeAway ? "Take Away" : order.status}
      </div>
      <div className="font-semibold">Total Price:</div>
      <div>${order.total_price.toFixed(2)}</div>
      <div className="font-semibold">Tracking Order:</div>
      <div>{order.tracking_order}</div>
      <div className="font-semibold">Chili Number:</div>
      <div>{order.chiliNumber}</div>
      {order.topping && (
        <>
          <div className="font-semibold">Toppings:</div>
          <div>{order.topping}</div>
        </>
      )}
    </div>

    <div className="mt-4">
      <h4 className="font-semibold mb-2">Individual Dishes:</h4>
      {order.data_dish.map((dish, index) => (
        <div key={`${dish.dish_id}-${index}`} className="ml-4 mb-2">
          <div>
            {dish.name} x{dish.quantity} (${dish.price.toFixed(2)} each)
          </div>
        </div>
      ))}
    </div>

    {order.data_set.length > 0 && (
      <div className="mt-4">
        <h4 className="font-semibold mb-2">Order Sets:</h4>
        {order.data_set.map((set, index) => (
          <div key={`${set.id}-${index}`} className="ml-4 mb-2">
            <div>
              {set.name} x{set.quantity} (${set.price.toFixed(2)} each)
            </div>
            <div className="ml-4 text-gray-600">
              Includes:
              {set.dishes.map((d, i) => (
                <React.Fragment key={d.dish_id}>
                  {i > 0 && ", "}
                  <span className="inline">
                    {d.name} (x{d.quantity})
                  </span>
                </React.Fragment>
              ))}
            </div>
          </div>
        ))}
      </div>
    )}
  </div>
);

export default OrderDetails;