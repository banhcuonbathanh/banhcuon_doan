import { DishService } from "./dish-application";
import { DishRepository } from "./dish-repository";

import {
  DishListResType,
  DishResType,
  CreateDishBodyType,
  UpdateDishBodyType
} from "./dish.schema";

export class DishController {
  private dishService: DishService;

  constructor() {
    const dishRepository = new DishRepository(); // You might need to adjust this based on your actual repository implementation
    this.dishService = new DishService(dishRepository);
  }

  async listDishes(): Promise<DishListResType["data"]> {
    try {
      const result = await this.dishService.listDishes();
      return result.data;
    } catch (error) {
      console.error("Error fetching dishes:", error);
      throw error;
    }
  }

  async addDish(body: CreateDishBodyType): Promise<DishResType["data"]> {
    try {
      const result = await this.dishService.addDish(body);
      return result.data;
    } catch (error) {
      console.error("Error adding dish:", error);
      throw error;
    }
  }

  async getDish(id: number): Promise<DishResType["data"]> {
    try {
      const result = await this.dishService.getDish(id);
      return result.data;
    } catch (error) {
      console.error(`Error fetching dish with id ${id}:`, error);
      throw error;
    }
  }

  async updateDish(
    id: number,
    body: UpdateDishBodyType
  ): Promise<DishResType["data"]> {
    try {
      const result = await this.dishService.updateDish(id, body);
      return result.data;
    } catch (error) {
      console.error(`Error updating dish with id ${id}:`, error);
      throw error;
    }
  }

  async deleteDish(id: number): Promise<void> {
    try {
      await this.dishService.deleteDish(id);
    } catch (error) {
      console.error(`Error deleting dish with id ${id}:`, error);
      throw error;
    }
  }
}

export default new DishController();
