import axios from "axios";

import envConfig from "@/config";

import { IAuthRepository } from "../../../auth/repository/interface_auth_repository";
import {
  LoginBodyType,
  LoginResType,
  LogoutBodyType,
  RefreshTokenBodyType,
  RefreshTokenResType,
  RegisterBodyType
} from "../../../auth/domain/auth.schema";
import {
  GuestLoginBodyType,
  GuestLoginResType
} from "@/schemaValidations/guest.schema";

class AuthRepository implements IAuthRepository {
  register(body: RegisterBodyType): Promise<RegisterBodyType> {
    throw new Error("Method not implemented.");
  }
  guestLogin(body: GuestLoginBodyType): Promise<GuestLoginResType> {
    throw new Error("Method not implemented.");
  }
  private baseUrl = envConfig.NEXT_PUBLIC_API_ENDPOINT;
  private createUserEndpoint = envConfig.NEXT_PUBLIC_API_Create_User;

  private refreshTokenRequest: Promise<{
    status: number;
    payload: RefreshTokenResType;
  }> | null = null;

  async sLogin(body: LoginBodyType): Promise<LoginResType> {
    const response = await axios.post<LoginResType>(
      this.baseUrl + this.createUserEndpoint,
      body
    );
    return response.data;
  }
  async sLogout(body: LogoutBodyType & { accessToken: string }): Promise<void> {
    await axios.post(
      "/auth/logout",
      { refreshToken: body.refreshToken },
      {
        headers: {
          Authorization: `Bearer ${body.accessToken}`
        }
      }
    );
  }

  async sRefreshToken(
    body: RefreshTokenBodyType
  ): Promise<RefreshTokenResType> {
    const response = await axios.post<RefreshTokenResType>(
      this.baseUrl + this.createUserEndpoint,
      body
    );
    return response.data;
  }

  async login(body: LoginBodyType): Promise<LoginResType> {
    const response = await axios.post<LoginResType>(
      this.baseUrl + this.createUserEndpoint,
      body
    );
    return response.data;
  }

  async logout(): Promise<void> {
    await axios.post(this.baseUrl + this.createUserEndpoint);
  }

  async refreshToken(): Promise<{
    status: number;
    payload: RefreshTokenResType;
  }> {
    if (this.refreshTokenRequest) {
      return this.refreshTokenRequest;
    }

    this.refreshTokenRequest = axios
      .post<RefreshTokenResType>(this.baseUrl + this.createUserEndpoint, null)
      .then((response) => ({
        status: response.status,
        payload: response.data
      }));

    const result = await this.refreshTokenRequest;
    this.refreshTokenRequest = null;
    return result;
  }
}

// Export an instance of the class
export const authApi = new AuthRepository();
