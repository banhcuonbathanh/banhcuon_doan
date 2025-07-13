// import envConfig from "@/config";
// import {
//   DishListRes,
//   DishListResType,
//   DishSchema
// } from "@/zusstand/dished/domain/dish.schema";
// import { z } from "zod";




// const get_dishes = async (): Promise<DishListResType> => {
//   console.log("Fetching dishes from controller...");

//   try {
//     const baseUrl = envConfig.NEXT_PUBLIC_URL + envConfig.NEXT_PUBLIC_Get_Dished_intenal;
//     console.log("Fetching from URL:", baseUrl);

//     const response = await fetch(baseUrl, {
//       method: "GET",
//       cache: "no-store"
//     });

//     if (!response.ok) {
//       const errorData = await response.json();
//       console.log("Error response data:", errorData);
//       throw new Error(`HTTP error! status: ${response.status}, message: ${errorData.message}`);
//     }

//     const data = await response.json();
//     console.log("Fetched dishes data:", data);

//     // Create a more lenient schema for parsing
//     const LenientDishSchema = DishSchema.extend({
//       createdAt: z.string().or(z.date()).optional(),
//       updatedAt: z.string().or(z.date()).optional(),
//     }).transform((dish) => ({
//       ...dish,
//       createdAt: dish.createdAt ? new Date(dish.createdAt) : new Date(),
//       updatedAt: dish.updatedAt ? new Date(dish.updatedAt) : new Date(),
//     }));

//     // Validate the response data against the lenient schema
//     const validatedData = z.array(LenientDishSchema).parse(data.data || data);

//     return validatedData;
//   } catch (error) {
//     console.error("Error fetching or parsing dishes:", error);
//     if (error instanceof z.ZodError) {
//       console.error("Zod validation errors:", JSON.stringify(error.errors, null, 2));
//     }
//     throw error;
//   }
// };



// export { get_dishes };
