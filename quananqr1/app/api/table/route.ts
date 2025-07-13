import { NextResponse } from "next/server";
import {
  CreateDishBody,
  DishListRes,
  DishRes
} from "@/schemaValidations/dish.schema"; // Adjust the import based on your project structure
import envConfig from "@/config";
import { CreateTableBody } from "@/zusstand/table/table.schema";

export async function GET() {
  const link_table = `${envConfig.NEXT_SERVER_API_ENDPOINT}${envConfig.NEXT_PUBLIC_Table_List}`;

  // console.log("quananqr1/app/api/table/route.ts link_table", link_table);
  try {
    const response = await fetch(
      `${envConfig.NEXT_SERVER_API_ENDPOINT}${envConfig.NEXT_PUBLIC_Table_List}`,
      {
        method: "GET",
        cache: "no-store"
      }
    );

    // Check for response.ok before parsing JSON
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const tables = await response.json(); // Parse response once
    // console.log("quananqr1/app/api/table/route.ts tables", tables);

    return NextResponse.json({
      data: tables.data,
      message: "tables retrieved successfully"
    });
  } catch (error) {
    console.error("Error fetching dishes:", error);
    return NextResponse.json(
      { message: "Failed to fetch dishes asdfasdf" },
      { status: 500 }
    );
  }
}

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const validatedBody = CreateTableBody.parse(body);

    const response = await fetch(
      `${envConfig.NEXT_SERVER_API_ENDPOINT}${envConfig.NEXT_PUBLIC_Table_List}`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify(validatedBody)
      }
    );

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const newDish = await response.json();
    return NextResponse.json(
      DishRes.parse({ data: newDish, message: "table created successfully" }),
      { status: 201 }
    );
  } catch (error) {
    console.error("Error creating table:", error);
    if (error instanceof Error) {
      return NextResponse.json(
        { message: `Failed to create table: ${error.message}` },
        { status: 400 }
      );
    }
    return NextResponse.json(
      { message: "Failed to create table" },
      { status: 400 }
    );
  }
}
