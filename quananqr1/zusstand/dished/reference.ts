// import { create } from 'zustand';
// import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
// import { DishResType, DishListResType, CreateDishBodyType, UpdateDishBodyType, DishParamsType } from './types';
// import { DishStatusValues } from '@/constants/type';

// interface DishStore {
//   dish: DishResType['data'] | null;
//   dishes: DishResType['data'][];
//   isLoading: boolean;
//   error: string | null;
//   getDish: (id: number) => Promise<void>;
//   updateDish: (body: UpdateDishBodyType & DishParamsType) => Promise<void>;
//   getDishes: () => Promise<void>;
//   addDish: (body: CreateDishBodyType) => Promise<void>;
//   deleteDish: (id: number) => Promise<void>;
// }

// const dishApiRequest = {
//   list: () => http.get<DishListResType>('dishes', { next: { tags: ['dishes'] } }),
//   add: (body: CreateDishBodyType) => http.post<DishResType>('dishes', body),
//   getDish: (id: number) => http.get<DishResType>(`dishes/${id}`),
//   updateDish: (id: number, body: UpdateDishBodyType) => http.put<DishResType>(`dishes/${id}`, body),
//   deleteDish: (id: number) => http.delete<DishResType>(`dishes/${id}`)
// };

// export const useDishStore = create<DishStore>((set) => ({
//   dish: null,
//   dishes: [],
//   isLoading: false,
//   error: null,
//   getDish: async (id: number) => {
//     set({ isLoading: true, error: null });
//     try {
//       const response = await dishApiRequest.getDish(id);
//       set({ dish: response.data.data, isLoading: false });
//     } catch (error) {
//       set({ isLoading: false, error: 'Failed to fetch dish' });
//       throw error;
//     }
//   },
//   updateDish: async (body: UpdateDishBodyType & DishParamsType) => {
//     set({ isLoading: true, error: null });
//     try {
//       const response = await dishApiRequest.updateDish(body.id, body);
//       set({ dish: response.data.data, isLoading: false });
//     } catch (error) {
//       set({ isLoading: false, error: 'Failed to update dish' });
//       throw error;
//     }
//   },
//   getDishes: async () => {
//     set({ isLoading: true, error: null });
//     try {
//       const response = await dishApiRequest.list();
//       set({ dishes: response.data.data, isLoading: false });
//     } catch (error) {
//       set({ isLoading: false, error: 'Failed to fetch dishes' });
//       throw error;
//     }
//   },
//   addDish: async (body: CreateDishBodyType) => {
//     set({ isLoading: true, error: null });
//     try {
//       const response = await dishApiRequest.add(body);
//       set((state) => ({
//         dishes: [...state.dishes, response.data.data],
//         isLoading: false
//       }));
//     } catch (error) {
//       set({ isLoading: false, error: 'Failed to add dish' });
//       throw error;
//     }
//   },
//   deleteDish: async (id: number) => {
//     set({ isLoading: true, error: null });
//     try {
//       await dishApiRequest.deleteDish(id);
//       set((state) => ({
//         dishes: state.dishes.filter(dish => dish.id !== id),
//         isLoading: false
//       }));
//     } catch (error) {
//       set({ isLoading: false, error: 'Failed to delete dish' });
//       throw error;
//     }
//   },
// }));

// export const useAddDishMutation = () => {
//   const { addDish, isLoading, error } = useDishStore();
//   return {
//     mutateAsync: addDish,
//     isPending: isLoading,
//     error,
//   };
// };

// export const useDeleteDishMutation = () => {
//   const queryClient = useQueryClient();
//   return useMutation({
//     mutationFn: dishApiRequest.deleteDish,
//     onSuccess: () => {
//       queryClient.invalidateQueries({
//         queryKey: ['dishes']
//       });
//     }
//   });
// };

// export const useDishListQuery = () => {
//   return useQuery({
//     queryKey: ['dishes'],
//     queryFn: dishApiRequest.list
//   });
// };

// export const useGetDishQuery = ({
//   id,
//   enabled
// }: {
//   id: number
//   enabled: boolean
// }) => {
//   return useQuery({
//     queryKey: ['dishes', id],
//     queryFn: () => dishApiRequest.getDish(id),
//     enabled
//   });
// };

// export const useUpdateDishMutation = () => {
//   const queryClient = useQueryClient();
//   return useMutation({
//     mutationFn: ({ id, ...body }: UpdateDishBodyType & { id: number }) =>
//       dishApiRequest.updateDish(id, body),
//     onSuccess: () => {
//       queryClient.invalidateQueries({
//         queryKey: ['dishes'],
//         exact: true
//       });
//     }
//   });
// };