import { DishStatusValues } from "@/constants/type";

export interface CreateDishBody {
  name: string;
  price: number;
  description: string;
  image: string;
  status?: typeof DishStatusValues[number];
}

export type CreateDishBodyType = CreateDishBody;



export interface DishInterface {
  id: number;
  name: string;
  price: number;
  description: string;
  image: string;
  status: typeof DishStatusValues[number];
  created_at: String;  // Note: capital 'S'
  updated_at: String;  // Note: capital 'S'
  set_id?: number;
}


export interface DishResInterface {
  data: DishInterface;
  message: string;
}

export type DishResTypeInterface = DishResInterface;

export type DishListResInterface = DishInterface[];

export type DishListResTypeInterface = DishListResInterface;

export type UpdateDishBody = CreateDishBody;
export type UpdateDishBodyType = CreateDishBodyType;

export interface DishParams {
  id: number;
}

export type DishParamsType = DishParams;

// New interfaces and types for Set and FavoriteSet

