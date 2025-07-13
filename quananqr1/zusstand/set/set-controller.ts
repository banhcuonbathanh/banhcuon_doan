import { create } from "zustand";
import { useApiStore } from "@/zusstand/api/api-controller";
import envConfig from "@/config";
import { z } from "zod";
import { SetType, SetListResType, SetListRes } from "@/schemaValidations/dish.schema";



interface SetStore {
  set: SetType | null;
  sets: SetType[];
  isLoading: boolean;
  error: string | null;
  getSet: (id: number) => Promise<void>;
  updateSet: (body: Partial<SetType> & { id: number }) => Promise<void>;
  getSets: () => Promise<SetType[]>;
  addSet: (body: Omit<SetType, "id" | "createdAt" | "updatedAt">) => Promise<void>;
  deleteSet: (id: number) => Promise<void>;
}

export const useSetStore = create<SetStore>((set) => ({
  set: null,
  sets: [],
  isLoading: false,
  error: null,
  getSet: async (id: number) => {
    set({ isLoading: true, error: null });
    try {
      const response = await useApiStore
        .getState()
        .http.get<{ data: SetType }>(`/api/sets/${id}`);
      set({ set: response.data.data, isLoading: false });
    } catch (error) {
      set({ isLoading: false, error: "Failed to fetch set" });
      throw error;
    }
  },
  updateSet: async (body: Partial<SetType> & { id: number }) => {
    set({ isLoading: true, error: null });
    try {
      const response = await useApiStore
        .getState()
        .http.put<{ data: SetType }>(`/api/sets/${body.id}`, body);
      set({ set: response.data.data, isLoading: false });
    } catch (error) {
      set({ isLoading: false, error: "Failed to update set" });
      throw error;
    }
  },
  getSets: async () => {
    const link = `${envConfig.NEXT_PUBLIC_API_ENDPOINT}${envConfig.NEXT_PUBLIC_Set_End_Point}`;
    set({ isLoading: true, error: null });

    try {
      const response = await useApiStore
        .getState()
        .http.get<SetListResType>(link);

      const validatedData = SetListRes.parse(response.data);

      set({ sets: validatedData, isLoading: false });
      return validatedData;
    } catch (error) {
      console.error("Error fetching or parsing sets:", error);
      if (error instanceof z.ZodError) {
        console.error(
          "Zod validation errors:",
          JSON.stringify(error.errors, null, 2)
        );
      }
      set({ isLoading: false, error: "Failed to fetch sets" });
      throw error;
    }
  },
  addSet: async (body: Omit<SetType, "id" | "createdAt" | "updatedAt">) => {
    const link = `${envConfig.NEXT_PUBLIC_API_ENDPOINT}${envConfig.NEXT_PUBLIC_Set_End_Point}`;
    set({ isLoading: true, error: null });
    try {
      const response = await useApiStore
        .getState()
        .http.post<{ data: SetType }>(link, body);
      set((state) => ({
        sets: [...state.sets, response.data.data],
        isLoading: false
      }));
    } catch (error) {
      set({ isLoading: false, error: "Failed to add set" });
      throw error;
    }
  },
  deleteSet: async (id: number) => {
    set({ isLoading: true, error: null });
    try {
      await useApiStore
        .getState()
        .http.delete<{ data: SetType }>(`/api/sets/${id}`);
      set((state) => ({
        sets: state.sets.filter((set) => set.id !== id),
        isLoading: false
      }));
    } catch (error) {
      set({ isLoading: false, error: "Failed to delete set" });
      throw error;
    }
  }
}));

// Custom hooks for each operation
export const useAddSetMutation = () => {
  const { addSet, isLoading, error } = useSetStore();
  return {
    mutateAsync: addSet,
    isPending: isLoading,
    error
  };
};

export const useDeleteSetMutation = () => {
  const { deleteSet, isLoading, error } = useSetStore();
  return {
    mutateAsync: deleteSet,
    isPending: isLoading,
    error
  };
};

export const useSetListQuery = () => {
  const { getSets, sets, isLoading, error } = useSetStore();
  return {
    refetch: getSets,
    data: sets,
    isLoading,
    error
  };
};

export const useGetSetQuery = () => {
  const { getSet, set, isLoading, error } = useSetStore();
  return {
    refetch: getSet,
    data: set,
    isLoading,
    error
  };
};

export const useUpdateSetMutation = () => {
  const { updateSet, isLoading, error } = useSetStore();
  return {
    mutateAsync: updateSet,
    isPending: isLoading,
    error
  };
};



// import React, { useEffect, useState } from 'react';
// import { 
//   useAddSetMutation, 
//   useDeleteSetMutation, 
//   useSetListQuery, 
//   useGetSetQuery, 
//   useUpdateSetMutation 
// } from './useSetStore';
// import { SetType } from '@/schemaValidations/set.schema';

// const SetManagement: React.FC = () => {
//   const [selectedSetId, setSelectedSetId] = useState<number | null>(null);
//   const [newSetName, setNewSetName] = useState('');

//   // Fetch all sets
//   const { data: sets, isLoading: setsLoading, error: setsError, refetch: refetchSets } = useSetListQuery();

//   // Fetch a single set
//   const { data: selectedSet, refetch: refetchSet } = useGetSetQuery();

//   // Mutations
//   const addSetMutation = useAddSetMutation();
//   const updateSetMutation = useUpdateSetMutation();
//   const deleteSetMutation = useDeleteSetMutation();

//   useEffect(() => {
//     if (selectedSetId) {
//       refetchSet(selectedSetId);
//     }
//   }, [selectedSetId, refetchSet]);

//   const handleAddSet = async () => {
//     try {
//       await addSetMutation.mutateAsync({
//         name: newSetName,
//         description: '',
//         dishes: [],
//         userId: 1, // Assuming a user ID, adjust as needed
//         isFavourite: 0,
//         isPopulate: 0,
//       });
//       setNewSetName('');
//       refetchSets();
//     } catch (error) {
//       console.error('Failed to add set:', error);
//     }
//   };

//   const handleUpdateSet = async (set: SetType) => {
//     try {
//       await updateSetMutation.mutateAsync({
//         ...set,
//         name: `${set.name} (Updated)`,
//       });
//       refetchSets();
//       if (selectedSetId === set.id) {
//         refetchSet(set.id);
//       }
//     } catch (error) {
//       console.error('Failed to update set:', error);
//     }
//   };

//   const handleDeleteSet = async (id: number) => {
//     try {
//       await deleteSetMutation.mutateAsync(id);
//       refetchSets();
//       if (selectedSetId === id) {
//         setSelectedSetId(null);
//       }
//     } catch (error) {
//       console.error('Failed to delete set:', error);
//     }
//   };

//   if (setsLoading) return <div>Loading sets...</div>;
//   if (setsError) return <div>Error loading sets: {setsError}</div>;

//   return (
//     <div>
//       <h1>Set Management</h1>
      
//       <div>
//         <input 
//           type="text" 
//           value={newSetName} 
//           onChange={(e) => setNewSetName(e.target.value)} 
//           placeholder="New Set Name"
//         />
//         <button onClick={handleAddSet} disabled={addSetMutation.isPending}>
//           {addSetMutation.isPending ? 'Adding...' : 'Add Set'}
//         </button>
//       </div>

//       <h2>All Sets</h2>
//       <ul>
//         {sets?.map((set) => (
//           <li key={set.id}>
//             {set.name}
//             <button onClick={() => setSelectedSetId(set.id)}>View</button>
//             <button onClick={() => handleUpdateSet(set)}>Update</button>
//             <button onClick={() => handleDeleteSet(set.id)}>Delete</button>
//           </li>
//         ))}
//       </ul>

//       {selectedSet && (
//         <div>
//           <h2>Selected Set Details</h2>
//           <p>Name: {selectedSet.name}</p>
//           <p>Description: {selectedSet.description}</p>
//           <p>Dishes: {selectedSet.dishes.length}</p>
//           <p>Created At: {selectedSet.createdAt.toLocaleString()}</p>
//           <p>Updated At: {selectedSet.updatedAt.toLocaleString()}</p>
//         </div>
//       )}
//     </div>
//   );
// };

// export default SetManagement;