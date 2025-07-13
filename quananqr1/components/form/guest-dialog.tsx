"use client";
import React from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle
} from "@/components/ui/dialog";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useAuthStore } from "@/zusstand/new_auth/new_auth_controller";
import { useSearchParams, useParams, useRouter, usePathname } from "next/navigation";
import {
  GuestLoginBody,
  GuestLoginBodyType
} from "@/schemaValidations/guest.schema";
import { handleErrorApi, handleLoginRedirect } from "@/lib/utils";

const GuestLoginDialog = () => {
  const {
    guestLogin,
    isGuestDialogOpen,
    closeGuestDialog,
    openLoginDialog,
    openRegisterDialog
  } = useAuthStore();

  const searchParams = useSearchParams();
  const params = useParams();

  const pathname = usePathname();
  const router = useRouter();


  const tableNumber = Number(params.number);
  const token = searchParams.get("token");

  const form = useForm<GuestLoginBodyType>({
    resolver: zodResolver(GuestLoginBody),
    defaultValues: {
      name: "",
      token: token ?? "",
      tableNumber: tableNumber
    }
  });

  const onSubmit = async (data: GuestLoginBodyType) => {
    console.log("quananqr1/components/form/guest-dialog.tsx data 123123", data);
    try {
      await guestLogin(data);
      handleLoginRedirect(pathname, router);
    } catch (error: any) {
      handleErrorApi({
        error,
        setError: form.setError
      });
    }
  };

  const handleLoginClick = () => {
    closeGuestDialog();
    openLoginDialog();
  };

  const handleRegisterClick = () => {
    closeGuestDialog();
    openRegisterDialog();
  };

  return (
    <Dialog
      open={isGuestDialogOpen}
      onOpenChange={(open) => !open && closeGuestDialog()}
    >
      <DialogContent className="sm:max-w-[425px] bg-white dark:bg-gray-800 shadow-lg">
        <DialogHeader>
          <DialogTitle className="text-2xl">Guest Login</DialogTitle>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Please enter your guest information to access the system
          </p>
        </DialogHeader>
        <Form {...form}>
          <form
            className="space-y-4 w-full"
            noValidate
            onSubmit={form.handleSubmit(onSubmit)}
          >
            <div className="grid gap-4">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <div className="grid gap-2">
                      <Label htmlFor="name">Name</Label>
                      <Input
                        id="name"
                        type="text"
                        placeholder="Enter your name"
                        required
                        className="border-2 border-gray-300 dark:border-gray-600"
                        {...field}
                      />
                      <FormMessage />
                    </div>
                  </FormItem>
                )}
              />
              <Button type="submit" className="w-full">
                Login as Guest
              </Button>
              <div className="flex flex-col gap-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={handleLoginClick}
                  className="w-full"
                >
                  Sign in with an account
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={handleRegisterClick}
                  className="w-full"
                >
                  Register new account
                </Button>
              </div>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
};

export default GuestLoginDialog;
