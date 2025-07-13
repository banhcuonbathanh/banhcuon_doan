import { create } from "zustand";
import { persist } from "zustand/middleware";
import axios, {
  AxiosInstance,
  InternalAxiosRequestConfig,
  AxiosResponse
} from "axios";
import Cookies from "js-cookie";
import envConfig from "@/config";
import { RefreshTokenResType } from "@/schemaValidations/auth.schema";

interface ApiStore {
  http: AxiosInstance;
  accessToken: string | null;
  setAccessToken: (token: string | null) => void;
  refreshToken: () => Promise<void>;
  setTableToken: (token: string) => void;
  tableToken: string | null;
}

export const useApiStore = create<ApiStore>()(
  persist(
    (set) => ({
      http: axios.create({
        baseURL: envConfig.NEXT_PUBLIC_API_ENDPOINT
      }),
      accessToken: null,
      tableToken: null,
      setAccessToken: (token) => {
        set({ accessToken: token });
        console.log("Access token updated:", token);
      },
      setTableToken: (token) => {
        set({ tableToken: token });
        console.log("Table token updated:", token);
      },
      refreshToken: async () => {
        try {
          const refreshToken = Cookies.get("refreshToken");
          if (!refreshToken) {
            throw new Error("No refresh token available");
          }
          const response = await axios.post<RefreshTokenResType>(
            "/api/auth/refresh-token",
            { refreshToken },
            { baseURL: "" }
          );
          set({ accessToken: response.data.data.accessToken });
          if (response.data.data.refreshToken) {
            Cookies.set("refreshToken", response.data.data.refreshToken, {
              secure: true,
              sameSite: "strict"
            });
          }
        } catch (error) {
          set({ accessToken: null });
          Cookies.remove("refreshToken");
        }
      }
    }),
    {
      name: "api-storage",
      skipHydration: true
    }
  )
);

// Setup interceptors
let http: AxiosInstance;

if (typeof window !== "undefined") {
  const store = useApiStore.getState();
  http = store.http;

  http.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
      const { accessToken, tableToken } = useApiStore.getState();

      config.headers = config.headers || {};

      if (accessToken) {
        config.headers.Authorization = `Bearer ${accessToken}`;
        console.log("Adding access token to request:", accessToken);
      }

      if (tableToken) {
        config.headers["X-Table-Token"] = tableToken;
        console.log("Adding table token to request:", tableToken);
      }

      return config;
    },
    (error) => Promise.reject(error)
  );

  http.interceptors.response.use(
    (response: AxiosResponse) => response,
    async (error) => {
      const originalRequest = error.config;
      if (error.response?.status === 401 && !originalRequest._retry) {
        originalRequest._retry = true;
        await useApiStore.getState().refreshToken();
        return http(originalRequest);
      }
      return Promise.reject(error);
    }
  );
}
