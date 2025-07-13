// import React, { useState, useMemo } from "react";
// import { ChevronDown, ChevronRight, Info } from "lucide-react";
// import {
//   OrderDetailedDish,
//   OrderDetailedResponse,
//   OrderSetDetailed
// } from "../component/new-order-column";
// import GroupToppings from "./toppping-display";
// import DishSummary from "./dish-summary/dishes-summary";
// import { logWithLevel } from "@/lib/log";

// interface RestaurantSummaryProps {
//   restaurantLayoutProps: OrderDetailedResponse[];
// }

// interface AggregatedDish extends OrderDetailedDish {}

// interface GroupedOrder {
//   orderName: string;
//   characteristic?: string;
//   tableNumber: number;
//   orders: OrderDetailedResponse[];
//   hasTakeAway: boolean;
// }

// interface OrderStore {
//   tableNumber: number;
//   getOrderSummary: () => {
//     dishes: OrderDetailedDish[];
//     totalPrice: number;
//     orderId: number;
//   };
//   clearOrder: () => void;
// }

// // Helper function to create OrderStore
// const getOrderStore = (order: OrderDetailedResponse): OrderStore => ({
//   tableNumber: order.table_number,
//   getOrderSummary: () => {
//     // Calculate total price of dishes
//     const dishTotal = order.data_dish.reduce(
//       (sum, dish) => sum + dish.price * dish.quantity,
//       0
//     );

//     // Calculate total price of sets
//     const setTotal = order.data_set.reduce(
//       (sum, set) => sum + set.price * set.quantity,
//       0
//     );

//     // Combine dishes from individual dishes and sets
//     const allDishes = [
//       ...order.data_dish,
//       ...order.data_set.flatMap((set) =>
//         set.dishes.map((dish) => ({
//           ...dish,
//           quantity: dish.quantity * set.quantity // Multiply by set quantity
//         }))
//       )
//     ];

//     // Group and sum quantities for duplicate dishes
//     const dishMap = new Map<number, OrderDetailedDish>();
//     allDishes.forEach((dish) => {
//       const existing = dishMap.get(dish.dish_id);
//       if (existing) {
//         existing.quantity += dish.quantity;
//       } else {
//         dishMap.set(dish.dish_id, { ...dish });
//       }
//     });

//     return {
//       dishes: Array.from(dishMap.values()),
//       totalPrice: dishTotal + setTotal,
//       orderId: order.id
//     };
//   },
//   clearOrder: () => {
//     // This is a view-only implementation, so clearOrder is a no-op
//     console.log("Clear order called - view only implementation");
//   }
// });

// // Helper Components
// const CollapsibleSection: React.FC<{
//   title: string;
//   children: React.ReactNode;
// }> = ({ title, children }) => {
//   const [isOpen, setIsOpen] = useState(false);

//   return (
//     <div className="mt-2">
//       <div
//         className="flex items-center cursor-pointer select-none p-2 rounded"
//         onClick={() => setIsOpen(!isOpen)}
//       >
//         <h3 className="text-md font-semibold">{title}</h3>
//         {isOpen ? (
//           <ChevronDown className="ml-2 h-4 w-4" />
//         ) : (
//           <ChevronRight className="ml-2 h-4 w-4" />
//         )}
//       </div>
//       {isOpen && (
//         <div className="p-2 border-l border-r border-b rounded-b">
//           {children}
//         </div>
//       )}
//     </div>
//   );
// };

// const OrderDetails: React.FC<{
//   order: OrderDetailedResponse;
// }> = ({ order }) => (
//   <div className="border-b last:border-b-0 py-4">
//     <div className="grid grid-cols-2 gap-2">
//       <div className="font-semibold">Table Number:</div>
//       <div>{order.table_number}</div>
//       <div className="font-semibold">Status:</div>
//       <div className={order.takeAway ? "text-red-600 font-bold" : ""}>
//         {order.takeAway ? "Take Away" : order.status}
//       </div>
//       <div className="font-semibold">Total Price:</div>
//       <div>${order.total_price.toFixed(2)}</div>
//       <div className="font-semibold">Tracking Order:</div>
//       <div>{order.tracking_order}</div>
//       <div className="font-semibold">Chili Number:</div>
//       <div>{order.chiliNumber}</div>
//       {order.topping && (
//         <>
//           <div className="font-semibold">Toppings:</div>
//           <div>{order.topping}</div>
//         </>
//       )}
//     </div>

//     <div className="mt-4">
//       <h4 className="font-semibold mb-2">Individual Dishes:</h4>
//       {order.data_dish.map((dish, index) => (
//         <div key={`${dish.dish_id}-${index}`} className="ml-4 mb-2">
//           <div>
//             {dish.name} x{dish.quantity} (${dish.price.toFixed(2)} each)
//           </div>
//         </div>
//       ))}
//     </div>

//     {order.data_set.length > 0 && (
//       <div className="mt-4">
//         <h4 className="font-semibold mb-2">Order Sets:</h4>
//         {order.data_set.map((set, index) => (
//           <div key={`${set.id}-${index}`} className="ml-4 mb-2">
//             <div>
//               {set.name} x{set.quantity} (${set.price.toFixed(2)} each)
//             </div>
//             <div className="ml-4 text-gray-600">
//               Includes:
//               {set.dishes.map((d, i) => (
//                 <React.Fragment key={d.dish_id}>
//                   {i > 0 && ", "}
//                   <span className="inline">
//                     {d.name} (x{d.quantity})
//                   </span>
//                 </React.Fragment>
//               ))}
//             </div>
//           </div>
//         ))}
//       </div>
//     )}
//   </div>
// );

// const parseOrderName = (orderName: string): string => {
//   const parts = orderName.split("-");
//   return parts[0].trim();
// };

// const getOrdinalSuffix = (num: number): string => {
//   const j = num % 10;
//   const k = num % 100;
//   if (j === 1 && k !== 11) return "st";
//   if (j === 2 && k !== 12) return "nd";
//   if (j === 3 && k !== 13) return "rd";
//   return "th";
// };

// const aggregateDishes = (orders: OrderDetailedResponse[]): AggregatedDish[] => {
//   const dishMap = new Map<number, AggregatedDish>();
//   logWithLevel(
//     {
//       dishMap
//     },
//     "quananqr1/app/manage/admin/orders/restaurant-summary/restaurant-summary.tsx",
//     "info",
//     1
//   );

//   orders.forEach((order) => {
//     // Add individual dishes
//     order.data_dish.forEach((dish) => {
//       const existingDish = dishMap.get(dish.dish_id);
//       if (existingDish) {
//         existingDish.quantity += dish.quantity;
//       } else {
//         dishMap.set(dish.dish_id, {
//           ...dish,
//           quantity: dish.quantity
//         });
//       }
//     });

//     // Add dishes from sets
//     order.data_set.forEach((set) => {
//       set.dishes.forEach((setDish) => {
//         const existingDish = dishMap.get(setDish.dish_id);
//         if (existingDish) {
//           existingDish.quantity += setDish.quantity * set.quantity;
//         } else {
//           dishMap.set(setDish.dish_id, {
//             ...setDish,
//             quantity: setDish.quantity * set.quantity
//           });
//         }
//       });
//     });
//   });

//   return Array.from(dishMap.values());
// };

// const GroupSummary: React.FC<{ orders: OrderDetailedResponse[] }> = ({
//   orders
// }) => {
//   const [isDetailsVisible, setIsDetailsVisible] = useState(false);
//   const totals = useMemo(() => {
//     let dishTotal = 0;
//     let setTotal = 0;

//     orders.forEach((order) => {
//       order.data_dish.forEach((dish) => {
//         dishTotal += dish.price * dish.quantity;
//       });

//       order.data_set.forEach((set) => {
//         setTotal += set.price * set.quantity;
//       });
//     });

//     return {
//       dishTotal,
//       setTotal,
//       grandTotal: dishTotal + setTotal
//     };
//   }, [orders]);

//   return (
//     <div className="mt-4 pt-4 border-t">
//       <div
//         className="cursor-pointer select-none"
//         onClick={() => setIsDetailsVisible(!isDetailsVisible)}
//       >
//         <div className="grid grid-cols-2 gap-2">
//           <div className="font-bold text-lg">Total:</div>
//           <div className="text-right font-bold text-lg">
//             ${totals.grandTotal.toFixed(2)}
//             <ChevronDown
//               className={`inline-block ml-2 h-4 w-4 transition-transform duration-200 ${
//                 isDetailsVisible ? "transform rotate-180" : ""
//               }`}
//             />
//           </div>
//         </div>

//         {isDetailsVisible && (
//           <div className="grid grid-cols-2 gap-2 mt-2 text-sm">
//             <div className="font-medium">Individual Dishes:</div>
//             <div className="text-right">${totals.dishTotal.toFixed(2)}</div>

//             <div className="font-medium">Set Orders:</div>
//             <div className="text-right">${totals.setTotal.toFixed(2)}</div>
//           </div>
//         )}
//       </div>
//     </div>
//   );
// };

// export const RestaurantSummary: React.FC<RestaurantSummaryProps> = ({
//   restaurantLayoutProps
// }) => {
//   const groupedOrders = useMemo(() => {
//     const groups = new Map<string, GroupedOrder>();

//     restaurantLayoutProps.forEach((order) => {
//       const characteristic = parseOrderName(order.order_name);
//       const groupKey = `${characteristic}-${order.table_number}`;

//       if (!groups.has(groupKey)) {
//         groups.set(groupKey, {
//           orderName: characteristic,
//           tableNumber: order.table_number,
//           orders: [],
//           hasTakeAway: false
//         });
//       }
//       const group = groups.get(groupKey)!;
//       group.orders.push(order);
//       // Update hasTakeAway if any order in the group is takeaway
//       if (order.takeAway) {
//         group.hasTakeAway = true;
//       }
//     });

//     return Array.from(groups.values());
//   }, [restaurantLayoutProps]);

//   return (
//     <div className="p-4">
//       <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
//         {groupedOrders.map((group) => {
//           const aggregatedDishes = aggregateDishes(group.orders);
//           logWithLevel(
//             {
//               aggregatedDishes
//             },
//             "quananqr1/app/manage/admin/orders/restaurant-summary/restaurant-summary.tsx",
//             "info",
//             2
//           );
//           return (
//             <div
//               key={`${group.orderName}-${group.tableNumber}`}
//               className="shadow-md rounded-lg p-4 border"
//             >
//               <h3 className="text-xl font-semibold mb-4">
//                 {group.orderName} - Bàn {group.tableNumber}
//                 {group.hasTakeAway && (
//                   <span className="ml-2 text-red-600">(Đem đi)</span>
//                 )}
//               </h3>

//               <div className="rounded-lg shadow-sm p-4">
//                 <CollapsibleSection title="Canh">
//                   <GroupToppings orders={group.orders} />
//                 </CollapsibleSection>

//                 <CollapsibleSection title="Món Ăn">
//                   {aggregatedDishes.map((dish, index) => {
//                     const order = group.orders[0]; // Get the first order as reference

//                     return (
//                       <DishSummary
//                         key={`${dish.dish_id}-${index}`}
//                         dish={{
//                           id: dish.dish_id,
//                           name: dish.name,
//                           price: dish.price,
//                           description: dish.description,
//                           image_url: dish.iamge // Note: Using the typo'd field name from interface
//                         }}
//                         guest={
//                           order.is_guest
//                             ? {
//                                 id: order.guest_id,
//                                 name: order.order_name // Using order_name as guest name
//                               }
//                             : null
//                         }
//                         user={
//                           !order.is_guest
//                             ? {
//                                 id: order.user_id,
//                                 name: order.order_name // Using order_name as user name
//                               }
//                             : null
//                         }
//                         isGuest={order.is_guest}
//                         orderStore={{
//                           tableNumber: order.table_number,
//                           getOrderSummary: () => {
//                             const dishes = group.orders.flatMap((order) => [
//                               ...order.data_dish.map((d) => ({
//                                 id: d.dish_id,
//                                 quantity: d.quantity
//                               })),
//                               ...order.data_set.flatMap((set) =>
//                                 set.dishes.map((d) => ({
//                                   id: d.dish_id,
//                                   quantity: d.quantity * set.quantity
//                                 }))
//                               )
//                             ]);

//                             // Combine quantities for same dish IDs
//                             const combinedDishes = dishes.reduce(
//                               (acc, curr) => {
//                                 const existing = acc.find(
//                                   (d) => d.id === curr.id
//                                 );
//                                 if (existing) {
//                                   existing.quantity += curr.quantity;
//                                 } else {
//                                   acc.push({ ...curr });
//                                 }
//                                 return acc;
//                               },
//                               [] as Array<{ id: number; quantity: number }>
//                             );

//                             return {
//                               dishes: combinedDishes,
//                               orderId: order.id,
//                               totalPrice: order.total_price
//                             };
//                           },
//                           clearOrder: () => {
//                             console.log(
//                               "Clear order called - view only implementation"
//                             );
//                           }
//                         }}
//                       />
//                     );
//                   })}
//                 </CollapsibleSection>
//                 <CollapsibleSection title="Lần Gọi Đồ">
//                   {group.orders.map((order, index) => (
//                     <div key={order.id} className="mb-4 last:mb-0">
//                       <div className="font-medium text-lg mb-2">
//                         {`${index + 1}${getOrdinalSuffix(index + 1)} Order`}
//                       </div>
//                       <OrderDetails order={order} />
//                     </div>
//                   ))}
//                 </CollapsibleSection>
//                 <GroupSummary orders={group.orders} />
//               </div>
//             </div>
//           );
//         })}
//       </div>
//     </div>
//   );
// };

// export default RestaurantSummary;
