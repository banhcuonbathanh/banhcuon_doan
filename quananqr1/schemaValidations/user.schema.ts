import z from "zod";
import { RoleValues } from "@/constants/type";

// Define a schema for Google's protobuf Timestamp (assuming it's in string format for simplicity)
const TimestampSchema = z.string();

export const UserSchema = z.object({
  id: z.number().int(),
  name: z.string(),
  email: z.string(),
  password: z.string(),
  role: RoleValues,
  phone: z.string(),
  image: z.string().nullable(),
  address: z.string().nullable(),
  created_at: TimestampSchema,
  updated_at: TimestampSchema,
  favorite_food: z.array(z.number().int()) // Added array of favorite foods
});

export type User = z.infer<typeof UserSchema>;

export const UserListRes = z.object({
  data: z.array(UserSchema),
  message: z.string()
});

export type UserListResType = z.infer<typeof UserListRes>;

export const UserRes = z
  .object({
    data: UserSchema,
    message: z.string()
  })
  .strict();

export type UserResType = z.infer<typeof UserRes>;

export const CreateEmployeeUserBody = z
  .object({
    name: z.string().trim().min(2).max(256),
    email: z.string().email(),
    avatar: z.string().url().optional(),
    password: z.string().min(6).max(100),
    confirmPassword: z.string().min(6).max(100)
  })
  .strict()
  .superRefine(({ confirmPassword, password }, ctx) => {
    if (confirmPassword !== password) {
      ctx.addIssue({
        code: "custom",
        message: "Mật khẩu không khớp",
        path: ["confirmPassword"]
      });
    }
  });

export type CreateEmployeeUserBodyType = z.TypeOf<
  typeof CreateEmployeeUserBody
>;

export const UpdateEmployeeUserBody = z
  .object({
    name: z.string().trim().min(2).max(256),
    email: z.string().email(),
    avatar: z.string().url().optional(),
    changePassword: z.boolean().optional(),
    password: z.string().min(6).max(100).optional(),
    confirmPassword: z.string().min(6).max(100).optional()
  })
  .strict()
  .superRefine(({ confirmPassword, password, changePassword }, ctx) => {
    if (changePassword) {
      if (!password || !confirmPassword) {
        ctx.addIssue({
          code: "custom",
          message: "Hãy nhập mật khẩu mới và xác nhận mật khẩu mới",
          path: ["changePassword"]
        });
      } else if (confirmPassword !== password) {
        ctx.addIssue({
          code: "custom",
          message: "Mật khẩu không khớp",
          path: ["confirmPassword"]
        });
      }
    }
  });

export type UpdateEmployeeUserBodyType = z.TypeOf<
  typeof UpdateEmployeeUserBody
>;

export const UpdateMeBody = z
  .object({
    name: z.string().trim().min(2).max(256),
    avatar: z.string().url().optional()
  })
  .strict();

export type UpdateMeBodyType = z.TypeOf<typeof UpdateMeBody>;

export const ChangePasswordBody = z
  .object({
    oldPassword: z.string().min(6).max(100),
    password: z.string().min(6).max(100),
    confirmPassword: z.string().min(6).max(100)
  })
  .strict()
  .superRefine(({ confirmPassword, password }, ctx) => {
    if (confirmPassword !== password) {
      ctx.addIssue({
        code: "custom",
        message: "Mật khẩu mới không khớp",
        path: ["confirmPassword"]
      });
    }
  });

export type ChangePasswordBodyType = z.TypeOf<typeof ChangePasswordBody>;

export const UserIdParam = z.object({
  id: z.coerce.number()
});

export type UserIdParamType = z.TypeOf<typeof UserIdParam>;

export const GetListGuestsRes = z.object({
  data: z.array(
    z.object({
      id: z.number(),
      name: z.string(),
      tableNumber: z.number().nullable(),
      createdAt: z.date(),
      updatedAt: z.date()
    })
  ),
  message: z.string()
});

export type GetListGuestsResType = z.TypeOf<typeof GetListGuestsRes>;

export const GetGuestListQueryParams = z.object({
  fromDate: z.coerce.date().optional(),
  toDate: z.coerce.date().optional()
});

export type GetGuestListQueryParamsType = z.TypeOf<
  typeof GetGuestListQueryParams
>;

export const CreateGuestBody = z
  .object({
    name: z.string().trim().min(2).max(256),
    tableNumber: z.number()
  })
  .strict();

export type CreateGuestBodyType = z.TypeOf<typeof CreateGuestBody>;

export const CreateGuestRes = z.object({
  message: z.string(),
  data: z.object({
    id: z.number(),
    name: z.string(),
    role: RoleValues,
    tableNumber: z.number().nullable(),
    createdAt: z.date(),
    updatedAt: z.date()
  })
});

export type CreateGuestResType = z.TypeOf<typeof CreateGuestRes>;
