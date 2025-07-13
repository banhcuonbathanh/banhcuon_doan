import { authApi } from "../repository/account-repository";
import { IAuthRepository } from "../repository/interface_account_repository";

// Define DTOs (Data Transfer Objects)
export interface UserDTO {
  id: string;
  name: string;
  email: string;
}

export interface CreateUserDTO {
  name: string;
  email: string;
}

export interface UpdateUserDTO {
  name?: string;
  email?: string;
}

// Define custom error class
export class AuthError extends Error {
  constructor(message: string, public statusCode?: number) {
    super(message);
    this.name = "AuthError";
  }
}

export class AuthService {
  private authRepository: IAuthRepository;

  constructor(authRepository: IAuthRepository = authApi) {
    this.authRepository = authRepository;
  }

  async getAllUsers(): Promise<UserDTO[]> {
    try {
      const users = await this.authRepository.fetchUsers();
      return users.map(this.mapToUserDTO);
    } catch (error) {
      throw new AuthError("Failed to fetch users", 500);
    }
  }

  async getUserById(id: string): Promise<UserDTO> {
    try {
      const user = await this.authRepository.fetchUserById(id);
      return this.mapToUserDTO(user);
    } catch (error) {
      throw new AuthError(`Failed to fetch user with id ${id}`, 404);
    }
  }

  async createUser(userData: CreateUserDTO): Promise<UserDTO> {
    try {
      const createdUser = await this.authRepository.createUser(userData);
      return this.mapToUserDTO(createdUser);
    } catch (error) {
      throw new AuthError("Failed to create user", 400);
    }
  }

  async updateUser(id: string, userData: UpdateUserDTO): Promise<UserDTO> {
    try {
      const updatedUser = await this.authRepository.updateUser(id, userData);
      return this.mapToUserDTO(updatedUser);
    } catch (error) {
      throw new AuthError(`Failed to update user with id ${id}`, 400);
    }
  }

  async deleteUser(id: string): Promise<void> {
    try {
      await this.authRepository.deleteUser(id);
    } catch (error) {
      throw new AuthError(`Failed to delete user with id ${id}`, 400);
    }
  }

  private mapToUserDTO(user: any): UserDTO {
    return {
      id: user.id,
      name: user.name,
      email: user.email
    };
  }
}

// Export an instance of the service
export const authService = new AuthService();
