import { create } from "zustand";
import { persist } from "zustand/middleware";

import authApiRequest from "@/apiRequests/auth";
import { AccountType } from "@/schemaValidations/user.schema";
import { LoginBodyType } from "../../../auth/domain/auth.schema";

type AccountStore = {
  account: AccountType | null;
  isAuthenticated: boolean;
  accessToken: string | null;
  refreshToken: string | null;
  login: (credentials: LoginBodyType) => Promise<void>;
  logout: () => Promise<void>;
  setUser: (account: AccountType) => void;
  clearUser: () => void;
};

export const useAccountStore = create<AccountStore>()(
  persist(
    (set, get) => ({
      account: null,
      isAuthenticated: false,
      accessToken: null,
      refreshToken: null,
      login: async (credentials: LoginBodyType) => {
        try {
          const response = await authApiRequest.login(credentials);
          const { account, accessToken, refreshToken } = response.payload.data;
          set({
            account,
            isAuthenticated: true,
            accessToken,
            refreshToken
          });
        } catch (error) {
          console.error("Login failed:", error);
          throw error;
        }
      },
      logout: async () => {
        try {
          await authApiRequest.logout();
          set({
            account: null,
            isAuthenticated: false,
            accessToken: null,
            refreshToken: null
          });
        } catch (error) {
          console.error("Logout failed:", error);
          throw error;
        }
      },
      setUser: (account: AccountType) =>
        set({ account, isAuthenticated: true }),
      clearUser: () =>
        set({
          account: null,
          isAuthenticated: false,
          accessToken: null,
          refreshToken: null
        })
    }),
    {
      name: "auth-store",
      skipHydration: true
    }
  )
);
