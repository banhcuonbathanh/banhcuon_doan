"use client";

import { ColumnDef } from "@tanstack/react-table";

import { CellAction } from "./table-cell-action";
import { CellActionImageBillboards } from "./table-cell-action-image-table";

import { Table } from "@/zusstand/table/table.schema";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger
} from "@/components/ui/tooltip";
import QRCodeTable from "@/components/qrcode-table";

const truncateToken = (token: string, maxLength: number = 50) => {
  return token.length > maxLength ? `${token.slice(0, maxLength)}...` : token;
};
const generateTestLink = (tableNumber: number, token: string) => {
  const baseUrl = "http://localhost:3000/table";
  return `${baseUrl}/${tableNumber}?token=${encodeURIComponent(token)}`;
};

export const columns: ColumnDef<Table>[] = [
  {
    accessorKey: "number",
    header: "Number"
  },
  {
    accessorKey: "capacity",
    header: "Capacity"
  },
  {
    accessorKey: "status",
    header: "Status"
  },
  {
    accessorKey: "token",
    header: "Token",
    cell: ({ row }) => {
      const token = row.original.token;
      const truncatedToken = truncateToken(token, 10);

      return (
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <span
                className="cursor-pointer"
                onClick={() => navigator.clipboard.writeText(token)}
              >
                {truncatedToken}
              </span>
            </TooltipTrigger>
            <TooltipContent>
              <p>{token}</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      );
    }
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
    accessorKey: "qrCode",
    header: "QR Code",
    cell: ({ row }) => {
      const truncatedTokenForQR = truncateToken(row.original.token, 50);
      const testLink = generateTestLink(
        row.original.number,
        truncatedTokenForQR
      );
      return (
        <div>
          <QRCodeTable
            token={testLink}
            tableNumber={row.original.number}
            width={300}
   
          />
          <a
            href={testLink}
            target="_blank"
            rel="noopener noreferrer"
            className="text-sm text-blue-500 hover:underline mt-2 block"
          >
            Test Link
          </a>
        </div>
      );
    }
  },
  {
    id: "actions",
    cell: ({ row }) => <CellAction data={row.original} />
  }
];
