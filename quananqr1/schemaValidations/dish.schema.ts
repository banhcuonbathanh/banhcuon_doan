import { DishStatusValues } from "@/constants/type";
import z from "zod";

export const CreateDishBody = z.object({
  name: z.string().min(1).max(256),
  price: z.coerce.number().positive(),
  description: z.string().max(10000),
  image: z.string().url(),
  status: z.enum(DishStatusValues).optional()
});

export type CreateDishBodyType = z.TypeOf<typeof CreateDishBody>;

export const DishSchema = z.object({
  id: z.number(),
  name: z.string(),
  price: z.coerce.number(),
  description: z.string(),
  image: z.string(),
  status: z.enum(DishStatusValues),
  created_at: z.string(), // Changed from z.date() to z.string()
  updated_at: z.string(), // Changed from z.date() to z.string()
  set_id: z.number().optional()
});
export const LenientDishSchema = z
  .object({
    id: z.number(),
    name: z.string(),
    price: z.number(),
    description: z.string(),
    image: z.string(),
    status: z.string(),
    created_at: z.string().or(z.date()).optional(),
    updated_at: z.string().or(z.date()).optional()
  })
  .transform((dish) => ({
    ...dish,
    created_at: dish.created_at ? new Date(dish.created_at) : new Date(),
    updated_at: dish.updated_at ? new Date(dish.updated_at) : new Date()
  }));
export type DishListResTypeTranform = z.infer<typeof LenientDishSchema>[];
export const DishRes = z.object({
  data: DishSchema,
  message: z.string()
});

export type DishResType = z.TypeOf<typeof DishRes>;

export const DishListRes = z.array(DishSchema);

export type DishListResType = z.TypeOf<typeof DishListRes>;

export const UpdateDishBody = CreateDishBody;
export type UpdateDishBodyType = CreateDishBodyType;
export const DishParams = z.object({
  id: z.coerce.number()
});
export type DishParamsType = z.TypeOf<typeof DishParams>;
export type Dish = z.TypeOf<typeof DishSchema>;

/// set schema
export const SetSchema = z.object({
  id: z.number(),
  name: z.string(),
  description: z.string().optional(),
  dishes: z.array(DishSchema),
  userId: z.number(),
  created_at: z.string(), // Changed from z.date() to z.string()
  updated_at: z.string(), // Changed from z.date() to z.string()
  is_favourite: z.boolean(), // Changed from z.number() to z.boolean()
  like_by: z.array(z.number()).nullable(), // Changed to nullable array
  is_public: z.boolean() // Added new field
});

export const SetListRes = z.array(SetSchema);
export type SetListResType = z.TypeOf<typeof SetListRes>;

export type SetType = z.TypeOf<typeof SetSchema>;

// set favourite dish

const FavoriteSetSchema = z.object({
  id: z.number(),
  userId: z.number(),
  name: z.string(),
  dishes: z.array(z.number()), // Array of dish IDs
  createdAt: z.date(),
  updatedAt: z.date()
});
export const FavoriteSetListRes = z.array(FavoriteSetSchema);
export type FavoriteSetListResType = z.TypeOf<typeof FavoriteSetListRes>;
export type FavoriteSet = z.TypeOf<typeof FavoriteSetSchema>;
