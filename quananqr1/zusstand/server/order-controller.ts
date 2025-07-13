import envConfig from "@/config";
import {
  GetOrdersRequest,
  OrderDetailedListResponse,
  PaginationInfo
} from "@/schemaValidations/interface/type_order";

export const get_Orders = async (
  params: GetOrdersRequest
): Promise<OrderDetailedListResponse> => {
  try {
    const baseUrl = `${envConfig.NEXT_PUBLIC_URL}${envConfig.Order_Internal_End_Point}`;
    const queryParams = new URLSearchParams({
      page: params.page.toString(),
      page_size: params.page_size.toString()
    });

    const response = await fetch(`${baseUrl}?${queryParams}`, {
      method: "GET",
      cache: "no-store"
    });

    const rawData = await response.json();
    // console.log(
    //   "quananqr1/zusstand/server/order-controller.ts rawData",
    //   rawData
    // );
    if (!response.ok) {
      throw new Error(rawData.message || "Failed to fetch orders");
    }

    interface ApiResponse {
      data: Array<{
        data_set: Array<{
          id: number;
          name: string;
          description: string;
          dishes: Array<{
            dish_id: number;
            quantity: number;
            name: string;
            price: number;
            description: string;
            image: string;
            status: string;
          }>;
          userId: number;
          created_at: string;
          updated_at: string;
          is_favourite: boolean;
          like_by: number[] | null;
          is_public: boolean;
          image: string;
          price: number;
          quantity: number;
        }> | null;
        data_dish: Array<{
          dish_id: number;
          quantity: number;
          name: string;
          price: number;
          description: string;
          image: string;
          status: string;
        }> | null;
        id: number;
        guest_id: number;
        user_id: number;
        table_number: number;
        order_handler_id: number;
        status: string;
        total_price: number;
        is_guest: boolean;
        topping: string;
        bow_no_chili: number;
        take_away: boolean;
        chili_number: number;
        table_token: string;
        order_name: string;
        tracking_order: string; // Added missing field
      }>;
      pagination: PaginationInfo;
      message: string;
    }

    const typedData = rawData as ApiResponse;

    if (!typedData.data) {
      throw new Error("Invalid response format: missing data");
    }

    const pagination: PaginationInfo = {
      current_page: typedData.pagination.current_page,
      total_pages: typedData.pagination.total_pages,
      total_items: typedData.pagination.total_items,
      page_size: typedData.pagination.page_size
    };

    const validatedData: OrderDetailedListResponse = {
      data: typedData.data.map((item) => ({
        id: item.id,
        guest_id: item.guest_id,
        user_id: item.user_id,
        is_guest: item.is_guest,
        table_number: item.table_number,
        order_handler_id: item.order_handler_id,
        status: item.status,
        created_at: "asdf", // Consider getting actual timestamps
        updated_at: "asdf", // Consider getting actual timestamps
        total_price: item.total_price,
        topping: item.topping,
        tracking_order: item.tracking_order, // Added missing field
        bow_no_chili: item.bow_no_chili,
        takeAway: item.take_away,
        chiliNumber: item.chili_number,
        table_token: item.table_token,
        order_name: item.order_name,
        // Add null check for data_set
        data_set: (item.data_set || []).map((set) => ({
          id: set.id,
          name: set.name,
          description: set.description,
          dishes: set.dishes.map((dish) => ({
            dish_id: dish.dish_id,
            quantity: dish.quantity,
            name: dish.name,
            price: dish.price,
            description: dish.description,
            iamge: dish.image, // Note: This is still using the typo from the interface
            status: dish.status
          })),
          userId: set.userId,
          created_at: set.created_at,
          updated_at: set.updated_at,
          is_favourite: set.is_favourite,
          like_by: set.like_by || [],
          is_public: set.is_public,
          image: set.image,
          price: set.price,
          quantity: set.quantity
        })),
        // Handle null data_dish by providing an empty array
        data_dish: (item.data_dish || []).map((dish) => ({
          dish_id: dish.dish_id,
          quantity: dish.quantity,
          name: dish.name,
          price: dish.price,
          description: dish.description,
          iamge: dish.image, // Note: This is still using the typo from the interface
          status: dish.status
        }))
      })),
      pagination: pagination
    };

    return validatedData;
  } catch (error) {
    console.error("Error fetching or parsing orders:", error);
    throw error;
  }
};
