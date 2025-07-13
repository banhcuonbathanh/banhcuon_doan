import React from "react";

import { OrderClient } from "./component/order-client";
import { get_Orders } from "@/zusstand/server/order-controller";


export default async function OrdersPage() {
  console.log("quananqr1/app/manage/admin/orders/page.tsx ");
  const initialOrders = await get_Orders({
    page: 1,
    page_size: 10
  });
  // console.log(
  //   "quananqr1/app/manage/admin/orders/page.tsx initialOrders ",
  //   initialOrders
  // );
  return (
    <div>
      <OrderClient
        initialData={initialOrders.data}
        initialPagination={initialOrders.pagination}
        // deliveryData={[]}
      />
      {/* <RestaurantLayout number={0} /> */}
    </div>
  );
}
