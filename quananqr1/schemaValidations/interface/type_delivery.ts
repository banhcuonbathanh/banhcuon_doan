// First, let's define the delivery status values
export const DeliveryStatusValues = {
    Pending: 'Pending',
    Assigned: 'Assigned',
    PickedUp: 'Picked Up',
    InTransit: 'In Transit',
    Delivered: 'Delivered',
    Failed: 'Failed',
    Cancelled: 'Cancelled'
  } as const;
  
  // Create a type from the values
  export type DeliveryStatus = typeof DeliveryStatusValues[keyof typeof DeliveryStatusValues];
  
  export interface DishDeliveryItemInterface {
    dish_id: number;
    quantity: number;
  }
  
  export interface DeliveryInterface {
    id: number;
    guest_id: number;
    user_id: number;
    is_guest: boolean;
    table_number: number;
    order_handler_id: number;
    status: string; // Generic status
    created_at: String;
    updated_at: String;
    total_price: number;
    dish_items: DishDeliveryItemInterface[];
    order_id: number;
    bow_chili: number;
    bow_no_chili: number;
    take_away: boolean;
    chili_number: number;
    table_token: string;
    client_name: string;
    delivery_address: string;
    delivery_contact: string;
    delivery_notes: string;
    scheduled_time: String;
    delivery_fee: number;
    delivery_status: DeliveryStatus;
    estimated_delivery_time: String;
    actual_delivery_time: String;
  }
  
  export interface CreateDeliveryBody {
    guest_id: number;
    user_id: number;
    is_guest: boolean;
    table_number: number;
    order_handler_id: number;
    status: string;
    total_price: number;
    dish_items: DishDeliveryItemInterface[];
    bow_chili: number;
    bow_no_chili: number;
    take_away: boolean;
    chili_number: number;
    table_token: string;
    client_name: string;
    delivery_address: string;
    delivery_contact: string;
    delivery_notes: string;
    scheduled_time: String;
    order_id: number;
    delivery_fee: number;
    delivery_status: DeliveryStatus;
  }
  
  export type CreateDeliveryBodyType = CreateDeliveryBody;
  
  export interface UpdateDeliveryBody {
    id: number;
    status: string;
    delivery_status: DeliveryStatus;
    driver_id: number;
    estimated_delivery_time: String;
    actual_delivery_time: String;
    delivery_notes: string;
  }
  
  export type UpdateDeliveryBodyType = UpdateDeliveryBody;
  
  export interface DeliveryDetailedDishInterface {
    dish_id: number;
    quantity: number;
    name: string;
    price: number;
    description: string;
    image: string;
    status: string;
  }
  
  export interface DeliveryDetailedInterface {
    id: number;
    guest_id: number;
    user_id: number;
    table_number: number;
    order_handler_id: number;
    status: string;
    total_price: number;
    data_dish: DeliveryDetailedDishInterface[];
    is_guest: boolean;
    bow_chili: number;
    bow_no_chili: number;
    take_away: boolean;
    chili_number: number;
    table_token: string;
    client_name: string;
    delivery_status: DeliveryStatus;
    driver_id: number;
    delivery_address: string;
    estimated_delivery_time: String;
    delivery_contact: string;
    delivery_notes: string;
  }
  
  export interface PaginationInterface {
    current_page: number;
    total_pages: number;
    total_items: number;
    page_size: number;
  }
  
  export interface DeliveryDetailedListInterface {
    data: DeliveryDetailedInterface[];
    pagination: PaginationInterface;
  }
  
  export interface GetDeliveriesParams {
    page: number;
    page_size: number;
  }
  
  export type GetDeliveriesParamsType = GetDeliveriesParams;
  
  export interface DeliveryParams {
    id: number;
  }
  
  export type DeliveryParamsType = DeliveryParams;
  
  export interface DeliveryClientNameParams {
    name: string;
  }
  
  export type DeliveryClientNameParamsType = DeliveryClientNameParams;
  
  export interface DeliveryResInterface {
    data: DeliveryInterface;
    message: string;
  }
  
  export type DeliveryResTypeInterface = DeliveryResInterface;
  
  export interface DeliveryListResInterface {
    data: DeliveryInterface[];
    pagination: PaginationInterface;
  }
  
  export type DeliveryListResTypeInterface = DeliveryListResInterface;