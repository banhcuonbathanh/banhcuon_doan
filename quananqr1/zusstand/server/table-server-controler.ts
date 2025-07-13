import envConfig from "@/config";
import { TableStatusValues } from "@/constants/type";

import { z } from "zod";
import { TableListResType, TableSchema } from "../table/table.schema";

const get_tables = async (): Promise<TableListResType> => {
  // console.log(
  //   "Starting get_tables function quananqr1/zusstand/server/table-server-controler.ts"
  // );
  try {
    const baseUrl =
      envConfig.NEXT_PUBLIC_URL + envConfig.NEXT_PUBLIC_intern_table_end_point;
    // console.log(
    //   `Fetching tables from: quananqr1/zusstand/server/table-server-controler.ts ${baseUrl}`
    // );

    const response = await fetch(baseUrl, {
      method: "GET",
      cache: "no-store"
    });

    if (!response.ok) {
      const errorData = await response.json();
      console.error(
        "Error response data: quananqr1/zusstand/server/table-server-controler.ts",
        errorData
      );
      throw new Error(
        `quananqr1/zusstand/server/table-server-controler.ts HTTP error! status: ${response.status}, message: ${errorData.message}`
      );
    }

    const data = await response.json();
    // console.log(
    //   "Received raw data: quananqr1/zusstand/server/table-server-controler.ts",
    //   JSON.stringify(data, null, 2)
    // );

    // Create a more lenient schema for parsing
    const LenientTableSchema = TableSchema.extend({
      createdAt: z.string().or(z.date()).optional(),
      updatedAt: z.string().or(z.date()).optional(),
      status: z.enum(TableStatusValues).optional()
    }).transform((table) => ({
      ...table,
      number: z.coerce.number().parse(table.number),
      capacity: z.coerce.number().parse(table.capacity),
      status: table.status || TableStatusValues[0], // Use first status as default if not provided
      token: table.token || "", // Provide a default empty string if token is missing
      createdAt: table.createdAt ? new Date(table.createdAt) : new Date(),
      updatedAt: table.updatedAt ? new Date(table.updatedAt) : new Date()
    }));

    // console.log(
    //   "Validating data against schema quananqr1/zusstand/server/table-server-controler.ts"
    // );
    // Validate the response data against the lenient schema
    const validatedData = z.array(LenientTableSchema).parse(data.data || data);

    // console.log(
    //   `Successfully validated ${validatedData.length} tables quananqr1/zusstand/server/table-server-controler.ts`
    // );

    return {
      data: validatedData,
      message: data.message || "Tables fetched successfully"
    };
  } catch (error) {
    console.error(
      "Error in get_tables function: quananqr1/zusstand/server/table-server-controler.ts",
      error
    );
    if (error instanceof z.ZodError) {
      console.error(
        "Zod validation errors:",
        JSON.stringify(error.errors, null, 2)
      );
    }
    throw error;
  } finally {
    // console.log(
    //   "Finished get_tables function quananqr1/zusstand/server/table-server-controler.ts"
    // );
  }
};

export { get_tables };
