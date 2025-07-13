// config.ts
import { z } from "zod";

const configSchema = z.object({
  NEXT_SERVER_API_ENDPOINT: z.string().default("http://localhost:8888/"),
  NEXT_PUBLIC_API_ENDPOINT: z.string().default("http://localhost:8888/"),
  NEXT_PUBLIC_URL: z.string().default("http://nextjs_app:3000/"),
  NEXT_PUBLIC_API_Create_User: z.string().default("users"),
  NEXT_PUBLIC_API_Get_Account_Email: z.string().default("users/email/"),
  NEXT_PUBLIC_API_Login: z.string().default("users/login"),
  NEXT_PUBLIC_API_Logout: z.string().default("users/logout"),
  NEXT_PUBLIC_Image_Upload: z.string().default("images/upload"),
  NEXT_PUBLIC_Add_Dished: z.string().default("dishes"),
  NEXT_PUBLIC_Add_Guest_login: z.string().default("qr/guest"),
  NEXT_PUBLIC_Get_Dished_intenal: z.string().default("api/guest"),
  NEXT_PUBLIC_Upload: z.string().default("uploads/"),
  NEXT_PUBLIC_Folder1_BE: z.string().default("quananqr/"),
  NEXT_PUBLIC_Table_List: z.string().default("table/list"),
  NEXT_PUBLIC_intern_table_end_point: z.string().default("api/table"),
  NEXT_PUBLIC_Table_End_Point: z.string().default("table"),
  NEXT_PUBLIC_Set_End_Point: z.string().default("sets"),
  NEXT_PUBLIC_Get_set_intenal: z.string().default("api/sets"),
  Order_Internal_End_Point: z.string().default("api/order"),
  Order_External_End_Point: z.string().default("orders"),
  NEXT_PUBLIC_API_Guest_Login: z.string().default("qr/guest/login"),
  NEXT_PUBLIC_API_Guest_Logout: z.string().default("qr/guest/logout"),
  wslink: z.string().default("ws://localhost:8888/ws/"),
  wsAuth: z.string().default("ws/api/ws-auth"),

  Delivery_External_End_Point: z.string().default("delivery")
});

const configProject = configSchema.safeParse({
  Delivery_External_End_Point:
    process.env.Delivery_External_End_Point || "delivery",

  wslink: process.env.NEXT_PUBLIC_WS_LINK || "ws://localhost:8888/ws/",
  wsAuth: process.env.NEXT_PUBLIC_WS_AUTH || "ws/api/ws-auth",
  Order_Internal_End_Point: process.env.ORDER_INTERNAL_END_POINT || "api/order",
  Order_External_End_Point: process.env.ORDER_EXTERNAL_END_POINT || "orders",
  NEXT_PUBLIC_API_Guest_Login:
    process.env.NEXT_PUBLIC_API_Guest_Login || "qr/guest/login",
  NEXT_PUBLIC_API_Guest_Logout:
    process.env.NEXT_PUBLIC_API_Guest_Logout || "qr/guest/logout",
  NEXT_PUBLIC_Get_set_intenal:
    process.env.NEXT_PUBLIC_Get_set_intenal || "api/sets",
  NEXT_SERVER_API_ENDPOINT:
    process.env.NEXT_SERVER_API_ENDPOINT || "http://go_app_ai:8888/",
  NEXT_PUBLIC_Set_End_Point: process.env.NEXT_PUBLIC_Set_End_Point || "sets",
  NEXT_PUBLIC_Table_End_Point:
    process.env.NEXT_PUBLIC_Table_End_Point || "table",
  NEXT_PUBLIC_intern_table_end_point:
    process.env.NEXT_PUBLIC_intern_table_end_point || "api/table",
  NEXT_PUBLIC_Table_List: process.env.NEXT_PUBLIC_Table_List || "table/list",
  NEXT_PUBLIC_Folder1_BE: process.env.NEXT_PUBLIC_Folder1_BE || "quananqr/",
  NEXT_PUBLIC_Upload: process.env.NEXT_PUBLIC_Upload || "uploads/",
  NEXT_PUBLIC_API_ENDPOINT:
    process.env.NEXT_PUBLIC_API_ENDPOINT || "http://localhost:8888/",
  NEXT_PUBLIC_URL: process.env.NEXT_PUBLIC_URL || "http://nextjs_app:3000/",
  NEXT_PUBLIC_API_Create_User:
    process.env.NEXT_PUBLIC_API_Create_User || "users",
  NEXT_PUBLIC_API_Get_Account_Email:
    process.env.NEXT_PUBLIC_API_Get_Account_Email || "users/email/",
  NEXT_PUBLIC_API_Login: process.env.NEXT_PUBLIC_API_Login || "users/login",
  NEXT_PUBLIC_API_Logout: process.env.NEXT_PUBLIC_API_Logout || "users/logout",
  NEXT_PUBLIC_Image_Upload:
    process.env.NEXT_PUBLIC_Image_Upload || "images/upload",
  NEXT_PUBLIC_Add_Dished: process.env.NEXT_PUBLIC_Add_Dished || "dishes",
  NEXT_PUBLIC_Add_Guest_login:
    process.env.NEXT_PUBLIC_Add_Guest_login || "qr/guest",
  NEXT_PUBLIC_Get_Dished_intenal:
    process.env.NEXT_PUBLIC_Get_Dished_intenal || "api/guest"
});

if (!configProject.success) {
  console.error("Environment validation failed:", configProject.error.errors);
  throw new Error("Invalid environment configuration");
}

const envConfig = configProject.data;

export default envConfig;

// - NEXT_PUBLIC_URL=${NEXT_PUBLIC_URL:-http://nextjs_app:3000/}

// - NEXT_PUBLIC_URL=${NEXT_PUBLIC_URL:-http://localhost:3000/}

export const APP_CONSTANTS = {
  Intervel_revalidata_Page_Order: 6000000,
  SESSION_TIMEOUT: 30 * 60 * 1000, // 30 minutes in milliseconds
  DEFAULT_PAGE_SIZE: 10
} as const;
