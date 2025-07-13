import { RoleValues } from "@/constants/type";
import z from "zod";

export const LoginBody = z
  .object({
    email: z.string().email(),
    password: z.string().min(6).max(100)
  })
  .strict();

export type LoginBodyType = z.TypeOf<typeof LoginBody>;


const TimestampSchema = z.string();
export const LoginRes = z.object({
  access_token: z.string(),
  accessTokenExpiresAt: z.string(),
  refresh_token: z.string(),
  refreshTokenExpiresAt: z.string(),
  sessionId: z.string(),
  user: z.object({
    id: z.number(),
    name: z.string(),
    email: z.string(),
    role: RoleValues,
    address: z.string(),
    phone: z.string(),
    image: z.string().nullable(),
    created_at: TimestampSchema,
    updated_at: TimestampSchema,
    favorite_food: z.array(z.number().int()) // Added array of favorite foods
  })
});

export type LoginResType = z.TypeOf<typeof LoginRes>;

export const RefreshTokenBody = z
  .object({
    refreshToken: z.string()
  })
  .strict();

export type RefreshTokenBodyType = z.TypeOf<typeof RefreshTokenBody>;

export const RefreshTokenRes = z.object({
  data: z.object({
    accessToken: z.string(),
    refreshToken: z.string()
  }),
  message: z.string()
});

export type RefreshTokenResType = z.TypeOf<typeof RefreshTokenRes>;

export const LogoutBody = z
  .object({
    refreshToken: z.string()
  })
  .strict();

export type LogoutBodyType = z.TypeOf<typeof LogoutBody>;

export const RegisterBody = z.object({
  name: z.string().min(1, "Name is required"),
  email: z.string().email("Invalid email address"),
  password: z.string().min(8, "Password must be at least 8 characters long"),
  role: z.string(),


  phone: z.string().min(10, {
    message: "Số điện thoại phải có ít nhất 10 số.",
  }).max(15, {
    message: "Số điện thoại không được quá 15 số.",
  }).refine((val) => /^\d+$/.test(val), {
    message: "Số điện thoại chỉ được chứa các chữ số.",
  }),
  image: z.string().optional(),
  address: z.string().min(1, "Address is required"),
  created_at: z.string().datetime(),
  updated_at: z.string().datetime()
});

export type RegisterBodyType = z.infer<typeof RegisterBody>;