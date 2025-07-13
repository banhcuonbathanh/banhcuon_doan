import { create } from "zustand";

import { useApiStore } from "@/zusstand/api/api-controller";

import envConfig from "@/config";
import { z } from "zod";

import {
  CreateDishBodyType,
  DishParamsType,
  DishResType,
  UpdateDishBodyType
} from "@/schemaValidations/dish.schema";
import { DishInterface } from "@/schemaValidations/interface/type_dish";
import { SetCreateBodyInterface } from "@/schemaValidations/interface/types_set";

interface DishStore {
  set_Dishes: string[];
  dish: DishInterface | null;
  dishes: DishInterface[];
  isLoading: boolean;
  error: string | null;
  getDish: (id: number) => Promise<void>;
  updateDish: (body: UpdateDishBodyType & DishParamsType) => Promise<void>;
  getDishes: () => Promise<DishInterface[]>;
  addDish: (body: CreateDishBodyType) => Promise<void>;
  deleteDish: (id: number) => Promise<void>;
  addSet: (body: SetCreateBodyInterface) => Promise<void>; // New function
}

export const useDishStore = create<DishStore>((set) => ({
  set_Dishes: [],
  dish: null,
  dishes: [],
  isLoading: false,
  error: null,
  getDish: async (id: number) => {
    set({ isLoading: true, error: null });
    try {
      const response = await useApiStore
        .getState()
        .http.get<DishResType>(`/api/dishes/${id}`);
      set({ dish: response.data.data, isLoading: false });
    } catch (error) {
      set({ isLoading: false, error: "Failed to fetch dish" });
      throw error;
    }
  },
  updateDish: async (body: UpdateDishBodyType & DishParamsType) => {
    set({ isLoading: true, error: null });
    try {
      const response = await useApiStore
        .getState()
        .http.put<DishResType>(`/api/dishes/${body.id}`, body);
      set({ dish: response.data.data, isLoading: false });
    } catch (error) {
      set({ isLoading: false, error: "Failed to update dish" });
      throw error;
    }
  },
  getDishes: async (): Promise<DishInterface[]> => {
    const link =
      envConfig.NEXT_PUBLIC_API_ENDPOINT + envConfig.NEXT_PUBLIC_Add_Dished;
    set({ isLoading: true, error: null });

    try {
      const response = await useApiStore
        .getState()
        .http.get<{ data: DishInterface[]; message: string }>(link);

      console.log(
        "quananqr1/zusstand/dished/controller/dished-controller.ts response",
        response.data
      );
      console.log("111111111111111 link", link);

      const dishes = response.data.data;
      set({ dishes, isLoading: false });
      return dishes;
    } catch (error) {
      console.error("Error fetching or parsing dishes:", error);
      if (error instanceof z.ZodError) {
        console.error(
          "Zod validation errors:",
          JSON.stringify(error.errors, null, 2)
        );
      }
      set({ isLoading: false, error: "Failed to fetch dishes" });
      throw error;
    }
  },
  addDish: async (body: CreateDishBodyType) => {
    const link =
      envConfig.NEXT_PUBLIC_API_ENDPOINT + envConfig.NEXT_PUBLIC_Add_Dished;
    set({ isLoading: true, error: null });
    try {
      const response = await useApiStore
        .getState()
        .http.post<DishResType>(link, body);
      set((state) => ({
        dishes: [...state.dishes, response.data.data],
        isLoading: false
      }));
    } catch (error) {
      set({ isLoading: false, error: "Failed to add dish" });
      throw error;
    }
  },
  deleteDish: async (id: number) => {
    set({ isLoading: true, error: null });
    try {
      await useApiStore
        .getState()
        .http.delete<DishResType>(`/api/dishes/${id}`);
      set((state) => ({
        dishes: state.dishes.filter((dish) => dish.id !== id),
        isLoading: false
      }));
    } catch (error) {
      set({ isLoading: false, error: "Failed to delete dish" });
      throw error;
    }
  },

  addSet: async (body: SetCreateBodyInterface) => {
    console.log("quananqr1/zusstand/dished/dished-controller.ts start");
    const link = `${envConfig.NEXT_PUBLIC_API_ENDPOINT}sets`; // Assuming this is the correct endpoint

    console.log("quananqr1/zusstand/dished/dished-controller.ts link", link);
    set({ isLoading: true, error: null });
    try {
      const response = await useApiStore
        .getState()
        .http.post<{ data: SetCreateBodyInterface; message: string }>(
          link,
          body
        );
      console.log(
        "quananqr1/zusstand/dished/dished-controller.ts response",
        response.data.data
      );
      set((state) => ({
        set_Dishes: [...state.set_Dishes, response.data.data.name], // Assuming set_Dishes stores set names
        isLoading: false
      }));
    } catch (error) {
      set({ isLoading: false, error: "Failed to add set" });
      throw error;
    }
  }
}));

// Custom hooks for each operation
export const useAddDishMutation = () => {
  const { addDish, isLoading, error } = useDishStore();
  return {
    mutateAsync: addDish,
    isPending: isLoading,
    error
  };
};

export const useDeleteDishMutation = () => {
  const { deleteDish, isLoading, error } = useDishStore();
  return {
    mutateAsync: deleteDish,
    isPending: isLoading,
    error
  };
};

export const useDishListQuery = async () => {
  const { getDishes, dishes, isLoading, error } = useDishStore();

  console.log(
    "quananqr1/zusstand/dished/controller/dished-controller.ts useDishListQuery getDishes",
    await getDishes()
  );
  return {
    refetch: getDishes,
    data: dishes,
    isLoading,
    error
  };
};

export const useGetDishQuery = () => {
  const { getDish, dish, isLoading, error } = useDishStore();
  return {
    refetch: getDish,
    data: dish,
    isLoading,
    error
  };
};

export const useUpdateDishMutation = () => {
  const { updateDish, isLoading, error } = useDishStore();
  return {
    mutateAsync: updateDish,
    isPending: isLoading,
    error
  };
};

// const { mutateAsync: addDish, isPending: isAdding, error: addError } = useAddDishMutation();
// const { mutateAsync: deleteDish, isPending: isDeleting, error: deleteError } = useDeleteDishMutation();
// const { refetch: fetchDishes, data: dishes, isLoading: isLoadingDishes, error: dishesError } = useDishListQuery();
// const { refetch: fetchDish, data: dish, isLoading: isLoadingDish, error: dishError } = useGetDishQuery();
// const { mutateAsync: updateDish, isPending: isUpdating, error: updateError } = useUpdateDishMutation();
