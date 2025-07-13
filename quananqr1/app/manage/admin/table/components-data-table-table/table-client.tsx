"use client";

import { useRouter } from "next/navigation";

import { columns } from "./table-columns";


import { DataTable } from "@/components/ui/data-table";
import { Heading } from "@/components/ui/heading";

import { Separator } from "@/components/ui/separator";
import EditDish from "../edit-table";
import { Table } from "@/zusstand/table/table.schema";
import AddTable from "../add-table";

interface TableClientProps {
  data: Table[];
}

export const TableClient: React.FC<TableClientProps> = ({ data }) => {
  const router = useRouter();
  console.log(
    "quananqr1/app/admin/table/components-data-table-table/table-client.tsx",
    data
  );
  return (
    <>
      {data && (
        <div className="flex flex-col">
          <div className="flex flex-rol">
            <Heading
              title={`Table (${data.length})`}
              description="Manage Table for your store"
            />
            <AddTable />
          </div>

          <Separator />
          <DataTable searchKey="name" columns={columns} data={data} />
        </div>
      )}
    </>
  );
};
