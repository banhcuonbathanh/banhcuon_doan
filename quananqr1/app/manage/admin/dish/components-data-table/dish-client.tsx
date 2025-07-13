"use client";

import { useRouter } from "next/navigation";

import { columns } from "./dish-columns";

import { Dish } from "@/schemaValidations/dish.schema";
import { DataTable } from "@/components/ui/data-table";
import { Heading } from "@/components/ui/heading";
import AddDish from "../add-dish";
import { Separator } from "@/components/ui/separator";
import EditDish from "../edit-dish";

interface DishClientProps {
  data: Dish[];
}

export const DishClient: React.FC<DishClientProps> = ({ data }) => {
  const router = useRouter();
  // console.log("quananqr1/app/admin/dish/components/dish-client.tsx", data);
  return (
    <>
      {data && (
        <div className="flex flex-col">
          <div className="flex flex-rol">
            <Heading
              title={`Billboards (${data.length})`}
              description="Manage billboards for your store"
            />
            <AddDish />
          </div>

          <Separator />
          <DataTable searchKey="name" columns={columns} data={data} />
        </div>
      )}
    </>
  );
};
