"use client";

import { ColumnDef } from "@tanstack/react-table";

import { CellAction } from "./cell-action";
import { CellActionImageBillboards } from "./cell-action-image-billboards";
import { Dish } from "@/schemaValidations/dish.schema";

export const columns: ColumnDef<Dish>[] = [
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
    cell: ({ row }) => new Date(row.original.createdAt).toLocaleDateString()
  },
  {
    accessorKey: "updatedAt",
    header: "Updated At",
    cell: ({ row }) => new Date(row.original.updatedAt).toLocaleDateString()
  },
  {
    id: "actions",
    cell: ({ row }) => <CellAction data={row.original} />
  }
];
