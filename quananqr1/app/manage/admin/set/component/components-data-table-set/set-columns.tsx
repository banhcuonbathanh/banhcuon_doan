"use client";

import { ColumnDef } from "@tanstack/react-table";
import { MoreHorizontal, ChevronDown, ChevronUp } from "lucide-react";
import Image from "next/image";
import React, { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { DataTable } from "@/components/ui/data-table";
import { DishInterface } from "@/schemaValidations/interface/type_dish";
import { set_dish_columns } from "../components-data-table-dish/set-dish-columns";
import {
  SetInterface,
  SetProtoDish
} from "@/schemaValidations/interface/types_set";
import { CellActionImageBillboards } from "./cell-action-set";

const DishDialog: React.FC<{ dish: DishInterface }> = ({ dish }) => {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        <Button variant="link" className="p-0 h-auto font-normal">
          {dish.name}
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px] bg-white dark:bg-gray-800 shadow-lg">
        <DialogHeader>
          <DialogTitle className="text-2xl">{dish.name}</DialogTitle>
        </DialogHeader>
        <div className="grid gap-4">
          <div className="relative w-full h-48">
            <Image
              src={dish.image}
              alt={dish.name}
              layout="fill"
              objectFit="cover"
              className="rounded-md"
            />
          </div>
          <div className="grid gap-2">
            <InfoRow label="ID" value={dish.id} />
            <InfoRow label="Name" value={dish.name} />
            <InfoRow label="Description" value={dish.description} />
            <InfoRow label="Price" value={`$${dish.price.toFixed(2)}`} />
            <InfoRow label="Status" value={dish.status} />
            <InfoRow label="Created At" value={String(dish.created_at)} />
            <InfoRow label="Updated At" value={String(dish.updated_at)} />
            {dish.set_id && <InfoRow label="Set ID" value={dish.set_id} />}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
};

const InfoRow: React.FC<{ label: string; value: string | number }> = ({
  label,
  value
}) => (
  <div className="grid grid-cols-3 items-center gap-4">
    <span className="text-sm font-medium">{label}:</span>
    <span className="col-span-2 text-sm">{value}</span>
  </div>
);

export const columns: ColumnDef<SetInterface>[] = [
  {
    accessorKey: "id",
    header: "ID",
    size: 60
  },
  {
    accessorKey: "name",
    header: "Set Name",
    size: 200
  },

  {
    accessorKey: "image",
    header: "Image",
    cell: ({ row }) => {
      return <CellActionImageBillboards data={row.original.image} />;
    }
  },
  {
    accessorKey: "description",
    header: "Description",
    size: 300,
    cell: ({ row }) => {
      const description = row.original.description;
      return description ? description : "N/A";
    }
  },

  {
    accessorKey: "is_public",
    header: "is_public",
    size: 60
  },
  {
    accessorKey: "dishes",
    header: "Dishes",
    size: 300,
    cell: ({ row }) => {
      const [isExpanded, setIsExpanded] = React.useState(false);
      const dishes = row.original.dishes;

      console.log(
        "quananqr1/app/admin/set/component/components-data-table-set/set-columns.tsx dishes",
        dishes
      );

      return (
        <div>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setIsExpanded(!isExpanded)}
          >
            {dishes.length} Dishes{" "}
            {isExpanded ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
          </Button>
          {isExpanded && (
            <ul className="mt-2 space-y-1">
              {dishes.map((setProtoDish: SetProtoDish) => (
                <li key={setProtoDish.dish.id} className="text-sm">
                  <DishDialog dish={setProtoDish.dish} /> - $
                  {setProtoDish.dish.price.toFixed(2)} (Qty:{" "}
                  {setProtoDish.quantity})
                </li>
              ))}
            </ul>
          )}
        </div>
      );
    }
  },
  {
    accessorKey: "createdAt",
    header: "Created At",
    size: 150,
    cell: ({ row }) => new Date(row.original.created_at).toLocaleDateString()
  },
  {
    accessorKey: "updatedAt",
    header: "Updated At",
    size: 150,
    cell: ({ row }) => new Date(row.original.updated_at).toLocaleDateString()
  },
  {
    id: "actions",
    size: 60,
    cell: ({ row }) => (
      <Button variant="ghost" size="icon">
        <MoreHorizontal className="h-4 w-4" />
      </Button>
    )
  }
];
