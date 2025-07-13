"use client";

import axios from "axios";
import { useState } from "react";
import { Copy, Edit, MoreHorizontal, Trash } from "lucide-react";
import { toast } from "react-hot-toast";
import { useParams, useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger
} from "@/components/ui/dropdown-menu";

import { AlertModal } from "@/components/modal/alert-modal";


import { Table } from "@/zusstand/table/table.schema";
import { useDeleteTableMutation } from "@/zusstand/table/table-client-controler";
import EditTable from "../edit-table";
interface CellActionProps {
  data: Table;
}

export const CellAction: React.FC<CellActionProps> = ({ data }) => {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [tableIdEdit, setTableIdEdit] = useState<number | undefined>(undefined);

  const { mutateAsync: deleteTable, isPending, error } = useDeleteTableMutation();

  const onConfirm = async () => {
    try {
      await deleteTable(data.number);
      toast.success("Table deleted successfully.");
      router.refresh();
    } catch (error) {
      toast.error("Failed to delete table. Please try again.");
    } finally {
      setOpen(false);
    }
  };

  const onCopy = (id: number) => {
    navigator.clipboard.writeText(id.toString());
    toast.success("Table ID copied to clipboard.");
  };

  return (
    <>
      <AlertModal
        isOpen={open}
        onClose={() => setOpen(false)}
        onConfirm={onConfirm}
        loading={isPending}
      />

      {tableIdEdit && (
        <EditTable
          id={tableIdEdit}
          setId={setTableIdEdit}
          onSubmitSuccess={() => {
            setTableIdEdit(undefined);
            router.refresh();
          }}
        />
      )}

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" className="h-8 w-8 p-0">
            <span className="sr-only">Open menu</span>
            <MoreHorizontal className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>

        <DropdownMenuContent align="end">
          <DropdownMenuLabel>Actions</DropdownMenuLabel>

          <DropdownMenuItem onClick={() => onCopy(data.number)}>
            <Copy className="mr-2 h-4 w-4" /> Copy Id
          </DropdownMenuItem>

          <DropdownMenuItem onClick={() => setTableIdEdit(data.number)}>
            <Edit className="mr-2 h-4 w-4" /> Update
          </DropdownMenuItem>

          <DropdownMenuItem onClick={() => setOpen(true)}>
            <Trash className="mr-2 h-4 w-4" /> Delete
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </>
  );
};

export default CellAction;