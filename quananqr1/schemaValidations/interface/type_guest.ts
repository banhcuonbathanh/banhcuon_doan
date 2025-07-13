// Base guest information interface
export interface GuestInfo {
  id: number;
  name: string;
  role: string;
  table_number: number;
  created_at: string; // Time.Time from Go will be serialized as string
  updated_at: string; // Time.Time from Go will be serialized as string
}

// Full guest interface with authentication details
export interface Guest {
  id: number;
  name: string;
  table_number: number;
  refresh_token: string;
  refresh_token_expires_at: string; // Time.Time from Go will be serialized as string
  created_at: string;
  updated_at: string;
}

// Login request interface
export interface GuestLoginRequest {
  name: string;
  table_number: number;
  token: string;
}

// Login response interface
export interface GuestLoginResponse {
  access_token: string;
  refresh_token: string;
  guest: GuestInfo;
  access_token_expires_at: string;
  refresh_token_expires_at: string;
  session_id: string;
}

// Logout request interface
export interface LogoutRequest {
  refresh_token: string;
}

// Refresh token request interface
export interface RefreshTokenRequest {
  refresh_token: string;
}

// Refresh token response interface
export interface RefreshTokenResponse {
  access_token: string;
  refresh_token: string;
  message: string;
}

// GRPC request interface for getting guest orders
export interface GuestGetOrdersGRPCRequest {
  guestId: number; // Note: Changed from guest_id to match your Go struct
}

// Type exports for convenience
export type GuestInfoType = GuestInfo;
export type GuestType = Guest;
export type GuestLoginRequestType = GuestLoginRequest;
export type GuestLoginResponseType = GuestLoginResponse;
export type LogoutRequestType = LogoutRequest;
export type RefreshTokenRequestType = RefreshTokenRequest;
export type RefreshTokenResponseType = RefreshTokenResponse;
export type GuestGetOrdersGRPCRequestType = GuestGetOrdersGRPCRequest;
