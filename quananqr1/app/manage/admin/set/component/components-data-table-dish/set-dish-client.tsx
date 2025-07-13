"use client";

import { useRouter } from "next/navigation";



import { Dish } from "@/schemaValidations/dish.schema";
import { DataTable } from "@/components/ui/data-table";
import { Heading } from "@/components/ui/heading";

import { Separator } from "@/components/ui/separator";
import { set_dish_columns } from "./set-dish-columns";
import { useEffect, useState } from "react";
import { DishInterface } from "@/schemaValidations/interface/type_dish";

interface DishClientProps {
  data: DishInterface[];
}

export const SetDishClient: React.FC<DishClientProps> = ({ data }) => {
  const [hydrated, setHydrated] = useState(false);
  const router = useRouter();

  useEffect(() => {
    setHydrated(true);
  }, []);

  console.log(
    "quananqr1/app/admin/set/component/components-data-table-dish/set-dish-client.tsx in set",
    data
  );

  if (!hydrated) {
    return null; // Prevent rendering until hydration is complete
  }

  return (
    <>
      {data && (
        <div className="flex flex-col">
          <div className="flex flex-row">
            <Heading
              title={`dish (${data.length})`}
              description="Manage dish for your store"
            />
            {/* <AddDish /> */}
          </div>

          <Separator />
          <DataTable searchKey="name" columns={set_dish_columns} data={data} />
        </div>
      )}
    </>
  );
};