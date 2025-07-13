import { create } from "zustand";
import { useApiStore } from "@/zusstand/api/api-controller";

import envConfig from "@/config";
import { z } from "zod";
import { TableSchema, UpdateTableBodyType, TableParamsType, CreateTableBodyType, TableResType, TableListResType } from "./table.schema";

type Table = z.infer<typeof TableSchema>;

interface TableStore {
  table: Table | null;
  tables: Table[];
  isLoading: boolean;
  error: string | null;
  getTable: (number: number) => Promise<void>;
  updateTable: (body: UpdateTableBodyType & TableParamsType) => Promise<void>;
  getTables: () => Promise<Table[]>;
  addTable: (body: CreateTableBodyType) => Promise<void>;
  deleteTable: (number: number) => Promise<void>;
}

export const useTableStore = create<TableStore>((set) => ({
  table: null,
  tables: [],
  isLoading: false,
  error: null,
  getTable: async (number: number) => {
    set({ isLoading: true, error: null });
    try {
      const response = await useApiStore
        .getState()
        .http.get<TableResType>(`/api/tables/${number}`);
      set({ table: response.data.data, isLoading: false });
    } catch (error) {
      set({ isLoading: false, error: "Failed to fetch table" });
      throw error;
    }
  },
  updateTable: async (body: UpdateTableBodyType & TableParamsType) => {
    set({ isLoading: true, error: null });
    try {
      const response = await useApiStore
        .getState()
        .http.put<TableResType>(`/api/tables/${body.number}`, body);
      set({ table: response.data.data, isLoading: false });
    } catch (error) {
      set({ isLoading: false, error: "Failed to update table" });
      throw error;
    }
  },
  getTables: async () => {
    const link = envConfig.NEXT_PUBLIC_API_ENDPOINT + envConfig.NEXT_PUBLIC_Table_List;
    set({ isLoading: true, error: null });
    
    try {
      const response = await useApiStore.getState().http.get<TableListResType>(link);

      // Create a more lenient schema for parsing
      const LenientTableSchema = TableSchema.extend({
        createdAt: z.string().or(z.date()),
        updatedAt: z.string().or(z.date())
      }).transform((table) => ({
        ...table,
        createdAt: new Date(table.createdAt),
        updatedAt: new Date(table.updatedAt)
      }));

      // Validate and transform the response data
      const validatedData = z.array(LenientTableSchema).parse(response.data.data);

      set({ tables: validatedData, isLoading: false });
      return validatedData;
    } catch (error) {
      console.error("Error fetching or parsing tables:", error);
      if (error instanceof z.ZodError) {
        console.error("Zod validation errors:", JSON.stringify(error.errors, null, 2));
      }
      set({ isLoading: false, error: "Failed to fetch tables" });
      throw error;
    }
  },
  addTable: async (body: CreateTableBodyType) => {
    const link = envConfig.NEXT_PUBLIC_API_ENDPOINT + envConfig.NEXT_PUBLIC_Table_End_Point;
    set({ isLoading: true, error: null });
    try {
      const response = await useApiStore
        .getState()
        .http.post<TableResType>(link, body);
      set((state) => ({
        tables: [...state.tables, response.data.data],
        isLoading: false
      }));
    } catch (error) {
      set({ isLoading: false, error: "Failed to add table" });
      throw error;
    }
  },
  deleteTable: async (number: number) => {
    set({ isLoading: true, error: null });
    try {
      await useApiStore
        .getState()
        .http.delete<TableResType>(`/api/tables/${number}`);
      set((state) => ({
        tables: state.tables.filter((table) => table.number !== number),
        isLoading: false
      }));
    } catch (error) {
      set({ isLoading: false, error: "Failed to delete table" });
      throw error;
    }
  }
}));

// Custom hooks for each operation
export const useAddTableMutation = () => {
  const { addTable, isLoading, error } = useTableStore();
  return {
    mutateAsync: addTable,
    isPending: isLoading,
    error
  };
};

export const useDeleteTableMutation = () => {
  const { deleteTable, isLoading, error } = useTableStore();
  return {
    mutateAsync: deleteTable,
    isPending: isLoading,
    error
  };
};

export const useTableListQuery = () => {
  const { getTables, tables, isLoading, error } = useTableStore();
  return {
    refetch: getTables,
    data: tables,
    isLoading,
    error
  };
};

export const useGetTableQuery = (p0: { id: any; }) => {
  const { getTable, table, isLoading, error } = useTableStore();
  return {
    refetch: getTable,
    data: table,
    isLoading,
    error
  };
};

export const useUpdateTableMutation = () => {
  const { updateTable, isLoading, error } = useTableStore();
  return {
    mutateAsync: updateTable,
    isPending: isLoading,
    error
  };
};