"use client";

import React, { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter, useSearchParams, useParams } from "next/navigation";
import { Loader2 } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from "@/components/ui/dialog";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { handleErrorApi } from "@/lib/utils";
import {
  GuestLoginBodyType,
  GuestLoginBody
} from "@/schemaValidations/guest.schema";
import { useAuthStore } from "@/zusstand/new_auth/new_auth_controller";

const GuestLoginDialog: React.FC = () => {
  const [open, setOpen] = useState(false);
  const { guestLogin, loading, error, clearError, isLoginDialogOpen } =
    useAuthStore();

  // setTimeout(() => {
  //   console.log(
  //     " quananqr1/app/(public)/tables/[number]/guest-login-form.tsx set time out"
  //   );
  //   setOpen(true);

  //   console.log(
  //     " quananqr1/app/(public)/tables/[number]/guest-login-form.tsx set time out",
  //     open
  //   );
  // }, 100); // 3 seconds
  const searchParams = useSearchParams();
  const params = useParams();
  const tableNumber = Number(params.number);
  const token = searchParams.get("token");
  const router = useRouter();

  const form = useForm<GuestLoginBodyType>({
    resolver: zodResolver(GuestLoginBody),
    defaultValues: {
      name: "",
      token: token ?? "",
      tableNumber
    }
  });

  useEffect(() => {
    if (!token) {
      router.push("/");
    }
  }, [token, router]);

  useEffect(() => {
    clearError();
  }, [clearError]);

  async function onSubmit(values: GuestLoginBodyType) {
    try {
      await guestLogin(values);
      if (!error) {
        setOpen(false);
        router.push("/guest/menu");
      }
    } catch (error) {
      handleErrorApi({
        error,
        setError: form.setError
      });
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button
          onClick={() => {
            setOpen(true);
          }}
        >
          Đăng nhập gọi món
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px] bg-white dark:bg-gray-800 shadow-lg">
        <DialogHeader>
          <DialogTitle className="text-2xl">Đăng nhập gọi món</DialogTitle>
        </DialogHeader>

        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        <Form {...form}>
          <form
            className="space-y-4 w-full"
            noValidate
            onSubmit={form.handleSubmit(onSubmit)}
          >
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <div className="grid gap-2">
                    <Label htmlFor="name">Tên khách hàng</Label>
                    <Input
                      id="name"
                      type="text"
                      required
                      {...field}
                      disabled={loading}
                      className="border-2 border-gray-300 dark:border-gray-600"
                    />
                    <FormMessage />
                  </div>
                </FormItem>
              )}
            />

            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Đang xử lý...
                </>
              ) : (
                "Đăng nhập"
              )}
            </Button>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
};

export default GuestLoginDialog;
