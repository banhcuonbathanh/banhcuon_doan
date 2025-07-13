export interface DishOrderItem {
  dish_id: number;
  quantity: number;
}

export interface SetOrderItem {
  set_id: number;
  quantity: number;
}

export interface Order {
  id: number;
  guest_id: number;
  user_id: number;
  is_guest: boolean;
  table_number: number;
  order_handler_id: number;
  status: string;
  created_at: string;
  updated_at: string;
  total_price: number;
  dish_items: DishOrderItem[];
  set_items: SetOrderItem[];
  topping: string;
  tracking_order: string;
  takeAway: boolean;
  chiliNumber: number;
  table_token: string; // Added to match Go struct
  order_name: string; // Added new field
}

export interface CreateOrderRequest {
  guest_id?: number | null;
  user_id?: number | null;
  is_guest: boolean;
  table_number: number;
  order_handler_id: number;
  status: string;
  created_at: string;
  updated_at: string;
  total_price: number;
  dish_items: DishOrderItem[];
  set_items: SetOrderItem[];
  topping: string;
  tracking_order: string;
  takeAway: boolean;
  chiliNumber: number;
  table_token: string; // Fixed casing to match Go naming convention
  order_name: string; // Added new field
}

export interface UpdateOrderRequest {
  id: number;
  guest_id: number;
  user_id: number;
  table_number: number;
  order_handler_id: number;
  status: string;
  total_price: number;
  dish_items: DishOrderItem[];
  set_items: SetOrderItem[];
  is_guest: boolean;
  topping: string;
  tracking_order: string;
  takeAway: boolean; // Added missing field
  chiliNumber: number; // Added missing field
  table_token: string; // Added missing field
  order_name: string; // Added new field
}

export interface GetOrdersRequest {
  page: number;
  page_size: number;
}

export interface PayOrdersRequest {
  guest_id?: number;
  user_id?: number;
}

export interface OrderResponse {
  data: Order;
}

export interface OrderListResponse {
  data: Order[];
  pagination: PaginationInfo; // Added to match Go struct
}

export interface OrderDetailedListResponse {
  data: OrderDetailedResponse[];
  pagination: PaginationInfo; // Fixed casing to match Go naming convention
}

export interface OrderDetailedResponse {
  data_set: OrderSetDetailed[];
  data_dish: OrderDetailedDish[];
  id: number;
  guest_id: number;
  user_id: number;
  is_guest: boolean;
  table_number: number;
  order_handler_id: number;
  status: string;
  created_at: string;
  updated_at: string;
  total_price: number;
  topping: string;
  tracking_order: string;
  takeAway: boolean;
  chiliNumber: number;
  table_token: string; // Added missing field
  order_name: string; // Added new field

  deliveryData?: Record<string, number>; // Add this line
}

export interface Guest {
  id: number;
  name: string;
  table_number: number;
  created_at: string;
  updated_at: string;
}

export interface OrderSetDetailed {
  id: number;
  name: string;
  description: string;
  dishes: OrderDetailedDish[];
  userId: number;
  created_at: string;
  updated_at: string;
  is_favourite: boolean;
  like_by: number[];
  is_public: boolean;
  image: string;
  price: number;
  quantity: number;
}

export interface OrderDetailedDish {
  dish_id: number;
  quantity: number;
  name: string;
  price: number;
  description: string;
  iamge: string; // Note: This appears to be a typo in the original ("iamge")
  status: string;
}

export interface PaginationInfo {
  current_page: number;
  total_pages: number;
  total_items: number;
  page_size: number;
}

// -------
