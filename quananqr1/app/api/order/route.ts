import { NextResponse } from "next/server";
import { DishRes } from "@/schemaValidations/dish.schema"; // Adjust the import based on your project structure
import envConfig from "@/config";
import { CreateTableBody } from "@/zusstand/table/table.schema";

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const page = searchParams.get("page") || "1";
  const page_size = searchParams.get("page_size") || "10";

  const link_order = `${envConfig.NEXT_SERVER_API_ENDPOINT}${envConfig.Order_External_End_Point}`;
  const queryParams = new URLSearchParams({
    page,
    page_size
  });

  const requestUrl = `${link_order}?${queryParams}`;
  console.log("Request URL:", requestUrl);

  try {
    const response = await fetch(requestUrl, {
      method: "GET",
      cache: "no-store",
      headers: {
        Accept: "application/json"
      }
    });

    if (!response.ok) {
      const errorText = await response.text();
      console.error("API Error Response:", errorText);
      throw new Error(
        `HTTP error! status: ${response.status}, message: ${errorText}`
      );
    }

    const data = await response.json();

    return NextResponse.json({
      data: data.data,
      pagination: data.pagination,
      message: "Orders retrieved successfully"
    });
  } catch (error) {
    console.error("Error fetching orders:", error);
    return NextResponse.json(
      {
        message: "Failed to fetch orders",
        error: "error at quananqr1/app/api/order/route.ts"
      },
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
