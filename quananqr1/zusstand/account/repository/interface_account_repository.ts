// interfaces/IUserApi.ts
export interface IAccountRepository {
  fetchUsers(): Promise<any[]>;
  fetchUserById(id: string): Promise<any>;
  createUser(userData: { name: string; email: string }): Promise<any>;
  updateUser(
    id: string,
    userData: { name?: string; email?: string }
  ): Promise<any>;
  deleteUser(id: string): Promise<void>;
}
