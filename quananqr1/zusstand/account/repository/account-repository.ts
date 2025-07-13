import axios from "axios";
import { IAccountRepository } from "./interface_account_repository";
import envConfig from "@/config";

const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

class AccountRepository implements IAccountRepository {
  private baseUrl = envConfig.NEXT_PUBLIC_API_ENDPOINT;

  private createUserEndpoint = envConfig.NEXT_PUBLIC_API_Create_User;

  async fetchUsers(): Promise<any[]> {
    try {
      const response = await axios.get(this.baseUrl);
      return response.data;
    } catch (error) {
      throw new Error("Failed to fetch users");
    }
  }

  async fetchUserById(id: string): Promise<any> {
    try {
      const response = await axios.get(`${this.baseUrl}/${id}`);
      return response.data;
    } catch (error) {
      throw new Error(`Failed to fetch user with id ${id}`);
    }
  }

  async createUser(userData: { name: string; email: string }): Promise<any> {
    try {
      const response = await axios.post(
        this.baseUrl + this.createUserEndpoint,
        userData
      );
      return response.data;
    } catch (error) {
      throw new Error("Failed to create user");
    }
  }

  async updateUser(
    id: string,
    userData: { name?: string; email?: string }
  ): Promise<any> {
    try {
      const response = await axios.put(`${this.baseUrl}/${id}`, userData);
      return response.data;
    } catch (error) {
      throw new Error(`Failed to update user with id ${id}`);
    }
  }

  async deleteUser(id: string): Promise<void> {
    try {
      await axios.delete(`${this.baseUrl}/${id}`);
    } catch (error) {
      throw new Error(`Failed to delete user with id ${id}`);
    }
  }
}

// Export an instance of the class
export const accountApi = new AccountRepository();
