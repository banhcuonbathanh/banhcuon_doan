import { create } from "zustand";
import { useApiStore } from "@/zusstand/api/api-controller";

import envConfig from "@/config";
import {
  GuestCreateOrdersBodyType,
  GuestCreateOrdersResType,
  GuestGetOrdersResType,
  GuestLoginBodyType,
  GuestLoginResType
} from "./guest.schema";
import { Dish, DishListResType } from "../../schemaValidations/dish.schema";

// dishes: Dish[];
// createOrders: (orders: GuestCreateOrdersBodyType) => Promise<void>;
// isLoading: boolean;
// error: string | null;

export interface GuestStore {
  guest: GuestLoginResType["data"]["guest"] | null;
  orders: GuestCreateOrdersResType["data"];
  dishes: Dish[];
  isLoading: boolean;
  error: string | null;
  login: (body: GuestLoginBodyType) => Promise<void>;
  createOrders: (body: GuestCreateOrdersBodyType) => Promise<void>;
  getOrders: () => Promise<void>;
  getDishes: () => Promise<void>;
  logout: () => void;
  refreshToken: () => Promise<void>;
}

export const useGuestStore = create<GuestStore>((set, get) => ({
  guest: null,
  orders: [],
  dishes: [],
  isLoading: false,
  error: null,
  getDishes: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await useApiStore
        .getState()
        .http.get<DishListResType>("/api/dishes");
      set({ dishes: response.data.data, isLoading: false });
    } catch (error) {
      set({ isLoading: false, error: "Failed to fetch dishes" });
      throw error;
    }
  },
  login: async (body: GuestLoginBodyType) => {
    set({ isLoading: true, error: null });
    try {
      const response = await useApiStore
        .getState()
        .http.post<GuestLoginResType>("/api/guest/login", body);
      set({
        guest: response.data.data.guest,
        isLoading: false
      });
      localStorage.setItem("accessToken", response.data.data.accessToken);
      localStorage.setItem("refreshToken", response.data.data.refreshToken);
    } catch (error) {
      set({ isLoading: false, error: "Failed to login" });
      throw error;
    }
  },

  createOrders: async (body: GuestCreateOrdersBodyType) => {
    set({ isLoading: true, error: null });
    try {
      const response = await useApiStore
        .getState()
        .http.post<GuestCreateOrdersResType>("/api/guest/orders", body);
      set((state) => ({
        orders: [...state.orders, ...response.data.data],
        isLoading: false
      }));
    } catch (error) {
      set({ isLoading: false, error: "Failed to create orders" });
      throw error;
    }
  },

  getOrders: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await useApiStore
        .getState()
        .http.get<GuestGetOrdersResType>("/api/guest/orders");
      set({ orders: response.data.data, isLoading: false });
    } catch (error) {
      set({ isLoading: false, error: "Failed to fetch orders" });
      throw error;
    }
  },

  logout: () => {
    set({ guest: null, orders: [] });
    localStorage.removeItem("accessToken");
    localStorage.removeItem("refreshToken");
  },

  refreshToken: async () => {
    set({ isLoading: true, error: null });
    try {
      const refreshToken = localStorage.getItem("refreshToken");
      if (!refreshToken) {
        throw new Error("No refresh token found");
      }

      const response = await useApiStore.getState().http.post<{
        access_token: string;
        refresh_token: string;
        message: string;
      }>("/api/guest/refresh-token", { refresh_token: refreshToken });

      localStorage.setItem("accessToken", response.data.access_token);
      localStorage.setItem("refreshToken", response.data.refresh_token);

      set({ isLoading: false });
    } catch (error) {
      set({ isLoading: false, error: "Failed to refresh token" });
      get().logout(); // Logout the user if token refresh fails
      throw error;
    }
  }
}));

// Custom hooks for each operation
export const useGuestLoginMutation = () => {
  const { login, isLoading, error } = useGuestStore();
  return {
    mutateAsync: login,
    isPending: isLoading,
    error
  };
};

export const useCreateOrdersMutation = () => {
  const { createOrders, isLoading, error } = useGuestStore();
  return {
    mutateAsync: createOrders,
    isPending: isLoading,
    error
  };
};

export const useGuestOrdersQuery = () => {
  const { getOrders, orders, isLoading, error } = useGuestStore();
  return {
    refetch: getOrders,
    data: orders,
    isLoading,
    error
  };
};

export const useGuestLogout = () => {
  const { logout } = useGuestStore();
  return logout;
};

export const useGuestRefreshToken = () => {
  const { refreshToken, isLoading, error } = useGuestStore();
  return {
    mutateAsync: refreshToken,
    isPending: isLoading,
    error
  };
};
