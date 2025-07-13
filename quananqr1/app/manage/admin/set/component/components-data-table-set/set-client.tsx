"use client";

import { useRouter } from "next/navigation";

import { columns } from "./set-columns";

import { DataTable } from "@/components/ui/data-table";
import { Heading } from "@/components/ui/heading";

import { Separator } from "@/components/ui/separator";

import { SetType } from "@/schemaValidations/dish.schema";

import AddSetPage from "../../add-set-page";

import {
  useDishListQuery,
  useDishStore
} from "@/zusstand/dished/dished-controller";

import { useEffect } from "react";
import { SetDishClient } from "../components-data-table-dish/set-dish-client";
import { SetInterface } from "@/schemaValidations/interface/types_set";
//   const { data: sets, isLoading: setsLoading, error: setsError, refetch: refetchSets } = useSetListQuery();
interface SetClientProps {
  data:  SetInterface[];
}

export const SetClient: React.FC<SetClientProps> = ({ data }) => {
  const { dishes, getDishes, isLoading, error } = useDishStore();
  console.log(
    "quananqr1/app/admin/set/component/components-data-table-set/set-client.tsx dishes adsfd",
    dishes
  );
  useEffect(() => {
    getDishes();
  }, []);

  // const { data: sets, isLoading: setsLoading, error: setsError, refetch: refetchSets } = useSetListQuery();
  const router = useRouter();
  // console.log(
  //   "quananqr1/app/admin/Set/components/Set-client.tsx asdfasd in set",
  //   dishes
  // );
  return (
    <>
      {data && (
        <div className="flex flex-col">
          <div className="flex flex-rol">
            <Heading
              title={`set (${data.length})`}
              description="Manage set for your store"
            />
            {/* <AddSet /> */}
          </div>

          <Separator />
          <DataTable searchKey="name" columns={columns} data={data} />

          <AddSetPage />

          <SetDishClient data={dishes} />
        </div>
      )}
    </>
  );
};
