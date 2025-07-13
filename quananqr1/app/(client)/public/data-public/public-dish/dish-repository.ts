import envConfig from "@/config";
import { normalizePath } from "@/lib/utils";
import { DishListResType, CreateDishBodyType, DishResType, UpdateDishBodyType } from "./dish.schema";

export class DishRepository {
  private baseUrl: string;

  constructor() {
    this.baseUrl = envConfig.NEXT_PUBLIC_API_ENDPOINT;
  }

  private async request<T>(
    method: string,
    path: string,
    body?: any,
    options?: RequestInit
  ): Promise<T> {
    const url = `${this.baseUrl}/${normalizePath(path)}`;
    const config: RequestInit = {
      method,
      headers: {
        "Content-Type": "application/json"
      },
      ...options
    };

    if (body) {
      config.body = JSON.stringify(body);
    }

    const response = await fetch(url, config);
    return response.json();
  }

  async list(): Promise<DishListResType> {
    return this.request<DishListResType>(
      "GET",
      "dishes",
      { next: { tags: ["dishes"] } }
    );
  }

  async add(body: CreateDishBodyType): Promise<DishResType> {
    return this.request<DishResType>("POST", "dishes", body);
  }

  async getDish(id: number): Promise<DishResType> {
    return this.request<DishResType>("GET", `dishes/${id}`);
  }

  async updateDish(id: number, body: UpdateDishBodyType): Promise<DishResType> {
    return this.request<DishResType>("PUT", `dishes/${id}`, body);
  }

  async deleteDish(id: number): Promise<DishResType> {
    return this.request<DishResType>("DELETE", `dishes/${id}`);
  }
}
