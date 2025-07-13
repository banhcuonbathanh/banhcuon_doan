import envConfig from "@/config";
import { DishInterface } from "@/schemaValidations/interface/type_dish";

import { z } from "zod";

const get_dishes = async (): Promise<DishInterface[]> => {
  try {
    const baseUrl =
      envConfig.NEXT_PUBLIC_URL + envConfig.NEXT_PUBLIC_Get_Dished_intenal;
    console.log(
      "quananqr1/zusstand/server/dish-controller.ts baseUrl",
      baseUrl
    );

    const response = await fetch(baseUrl, {
      method: "GET",
      cache: "no-store"
    });

    if (!response.ok) {
      const errorData = await response.json();
      console.log("Error response data:", errorData);
      throw new Error(
        `HTTP error! status: ${response.status}, message: ${errorData.message}`
      );
    }

    const data = await response.json();

    // Validate the response data against the DishSchema

    return data.data;
  } catch (error) {
    console.error("Error fetching or parsing dishes:", error);
    if (error instanceof z.ZodError) {
      console.error(
        "Zod validation errors:",
        JSON.stringify(error.errors, null, 2)
      );
    }
    throw error;
  }
};

export { get_dishes };
