
import { DishRepository } from './dish-repository';
import { DishListResType, CreateDishBodyType, DishResType, UpdateDishBodyType } from './dish.schema';


const ENTITY_ERROR_STATUS = 422;
const AUTHENTICATION_ERROR_STATUS = 401;

export class DishService {
  private dishRepository: DishRepository;

  constructor(dishRepository: DishRepository) {
    this.dishRepository = dishRepository;
  }

  private async executeWithErrorHandling<T>(operation: () => Promise<T>): Promise<T> {
    try {
      return await operation();
    } catch (error: any) {
      console.error('Error details:', error);
      if (error.status === 422) {
        throw new Error('Entity Error: ' + JSON.stringify(error.data));
      } else if (error.status === 401) {
        throw new Error('Authentication error');
      } else {
        throw new Error('HTTP Error: ' + error.message);
      }
    }
  }

  async listDishes(): Promise<DishListResType> {
    return this.executeWithErrorHandling(() => this.dishRepository.list());
  }

  async addDish(body: CreateDishBodyType): Promise<DishResType> {
    return this.executeWithErrorHandling(() => this.dishRepository.add(body));
  }

  async getDish(id: number): Promise<DishResType> {
    return this.executeWithErrorHandling(() => this.dishRepository.getDish(id));
  }

  async updateDish(id: number, body: UpdateDishBodyType): Promise<DishResType> {
    return this.executeWithErrorHandling(() => this.dishRepository.updateDish(id, body));
  }

  async deleteDish(id: number): Promise<DishResType> {
    return this.executeWithErrorHandling(() => this.dishRepository.deleteDish(id));
  }
}