export interface SetProtoDish {
  dish_id: number;
  quantity: number;
  name: string;
  price: number;
  description: string;
  image: string;
  status: string;
}

export interface SetInterface {
  id: number;
  name: string;
  description: string;
  dishes: SetProtoDish[];
  userId: number;
  created_at: string; // Time.Time from Go will be serialized as string
  updated_at: string; // Time.Time from Go will be serialized as string
  is_favourite: boolean;
  like_by: number[] | null;
  is_public: boolean;
  image: string;
  price: number;
}

export interface SetCreateBodyInterface {
  name: string;
  description?: string;
  dishes: {
    dish_id: number;
    quantity: number;
  }[];
  userId: number;
  is_public: boolean;
  image: string;
}

export interface SetListResponse {
  data: SetInterface[];
}

export type SetListResType = SetListResponse;

export type SetIntefaceType = SetInterface;

// FavoriteSet interfaces and types
// Note: We'll keep these as they are since you didn't provide new information about them

export interface FavoriteSet {
  id: number;
  userId: number;
  name: string;
  dishes: number[]; // Array of dish IDs
  createdAt: Date;
  updatedAt: Date;
}

export type FavoriteSetListRes = FavoriteSet[];
export type FavoriteSetListResType = FavoriteSetListRes;


// export interface SetCreateBodyInterface {
//   image: string;
//   name: string;
//   description?: string;
//   dishes: SetProtoDish[];
//   userId: number; // Changed from user_id to userId
//   created_at: string; // Changed from Date to string
//   updated_at: string; // Changed from Date to string
//   is_favourite: boolean; // Changed from number to boolean
//   like_by: number[] | null; // Changed to allow null
//   is_public: boolean; // Added new field
// }
