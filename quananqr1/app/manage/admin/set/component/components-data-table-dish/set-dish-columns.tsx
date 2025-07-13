"use client";

import { ColumnDef } from "@tanstack/react-table";

import { CellAction } from "./cell-action";
import { CellActionImageBillboards } from "./cell-action-image-billboards";

import { DishInterface } from "@/schemaValidations/interface/type_dish";

export const set_dish_columns: ColumnDef<DishInterface>[] = [
  {
    accessorKey: "id",
    header: "ID"
  },
  {
    accessorKey: "name",
    header: "Name"
  },
  {
    accessorKey: "price",
    header: "Price"
  },
  {
    accessorKey: "description",
    header: "Description"
  },
  {
    accessorKey: "image",
    header: "Image",
    cell: ({ row }) => {
      return <CellActionImageBillboards data={row.original.image} />;
    }
  },
  {
    accessorKey: "status",
    header: "Status"
  },
  {
    accessorKey: "createdAt",
    header: "Created At",
    cell: ({ row }) =>  row.original.created_at
  },
  {
    accessorKey: "updatedAt",
    header: "Updated At",
    cell: ({ row }) => row.original.updated_at
  },
  {
    id: "actions",
    cell: ({ row }) => <CellAction data={row.original} />
  }
];
