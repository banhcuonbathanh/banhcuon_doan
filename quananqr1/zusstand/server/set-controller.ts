import envConfig from "@/config";

import { SetInterface } from "@/schemaValidations/interface/types_set";

const get_Sets = async (): Promise<SetInterface[]> => {
  try {
    const baseUrl =
      envConfig.NEXT_PUBLIC_URL + envConfig.NEXT_PUBLIC_Get_set_intenal;

    const response = await fetch(baseUrl, {
      method: "GET",
      cache: "no-store"
    });

    const data = await response.json();

    const validatedData: SetInterface[] = data.data.map((set: any) => ({
      id: set.id,
      name: set.name,
      description: set.description,
      dishes: set.dishes.map((dish: any) => ({
        dish_id: dish.dish_id,
        quantity: dish.quantity,
        name: dish.name,
        price: dish.price,
        status: dish.status
      })),
      userId: set.userId,
      created_at: set.created_at,
      updated_at: set.updated_at,
      is_favourite: Boolean(set.is_favourite),
      like_by: set.like_by || [],
      is_public: Boolean(set.is_public),
      image: set.image,
      price: Number(set.price)
    }));

    return validatedData;
  } catch (error) {
    console.error("Error fetching or parsing sets:", error);
    throw error;
  }
};

export { get_Sets };
