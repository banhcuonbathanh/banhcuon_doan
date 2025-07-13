import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent
} from "@/components/ui/card";
import { get_tables } from "@/zusstand/server/table-server-controler";
import React from "react";

import { TableClient } from "./components-data-table-table/table-client";

const HomeTable = async () => {
  const tables = await get_tables();

  // console.log("quananqr1/app/admin/table/page.tsx tables", tables.data);
  return (
    <main className="grid flex-1 items-start gap-4 p-4 sm:px-6 sm:py-0 md:gap-8">
      <div className="space-y-2">
        <Card x-chunk="dashboard-06-chunk-0">
          <CardHeader>
            <CardTitle>Món ăn</CardTitle>
            <CardDescription>Quản lý món ăn</CardDescription>
          </CardHeader>
          <CardContent>
            <TableClient data={tables.data} />
          </CardContent>
        </Card>
      </div>
    </main>
  );
};

export default HomeTable;
