import { NextResponse } from "next/server";
import {
  CreateDishBody,
  DishListRes,
  DishRes
} from "@/schemaValidations/dish.schema"; // Adjust the import based on your project structure
import envConfig from "@/config";
import { CreateTableBody } from "@/zusstand/table/table.schema";

export async function GET() {
  const link_set = `${envConfig.NEXT_SERVER_API_ENDPOINT}${envConfig.NEXT_PUBLIC_Set_End_Point}`;

  // console.log("quananqr1/app/api/sets/route.ts link_set", link_set);
  try {
    const response = await fetch(link_set, {
      method: "GET",
      cache: "no-store"
    });

    // Check for response.ok before parsing JSON
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const tables = await response.json(); // Parse response once
    // console.log(
    //   "quananqr1/app/api/sets/route.ts set 2121212121212121212121212",
    //   tables
    // );

    return NextResponse.json({
      data: tables.data,
      message: "set retrieved successfully"
    });
  } catch (error) {
    console.error(
      "Error fetching set: quananqr1/app/api/sets/route.ts set",
      error
    );
    return NextResponse.json(
      { message: "Failed to fetch quananqr1/app/api/sets/route.ts set" },
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
