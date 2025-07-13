"use client";
import React, { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Info } from "lucide-react";
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
import { LoginBodyType, LoginBody } from "@/schemaValidations/auth.schema";
import { useAuthStore } from "@/zusstand/new_auth/new_auth_controller";
// import { handleErrorApi, handleLoginRedirect } from "@/lib/utils";
import { usePathname, useRouter } from "next/navigation";
import { handleErrorApi, handleLoginRedirect } from "@/lib/utils";

const LoginDialog1 = () => {
  console.log("quananqr1/components/form/login-dialog.tsx LoginDialog1");
  const {
    login,
    isLoginDialogOpen,
    closeLoginDialog,
    openRegisterDialog,
    openGuestDialog
  } = useAuthStore();
  const pathname = usePathname();
  const router = useRouter();

  useEffect(() => {
    console.log(
      "quananqr1/components/form/login-dialog.tsx pathname:",
      pathname
    );
  }, [pathname]); // Only log when pathname changes
  const form = useForm<LoginBodyType>({
    resolver: zodResolver(LoginBody),
    defaultValues: {
      email: "",
      password: ""
    }
  });

  const onSubmit = async (data: LoginBodyType) => {
    console.log("quananqr1/components/form/login-dialog.tsx onSubmit");
    try {
      await login(data);
      handleLoginRedirect(pathname, router);
    } catch (error: any) {
      handleErrorApi({
        error,
        setError: form.setError
      });
    }
  };

  const handleRegisterClick = () => {
    closeLoginDialog();
    openRegisterDialog();
  };

  const handleGuestClick = () => {
    closeLoginDialog();
    openGuestDialog();
  };

  return (
    <Dialog
      open={isLoginDialogOpen}
      onOpenChange={(open) => !open && closeLoginDialog()}
    >
      <DialogContent className="sm:max-w-[425px] bg-white dark:bg-gray-800 shadow-lg">
        <DialogHeader>
          <DialogTitle className="text-2xl">Đăng nhập</DialogTitle>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Nhập email và mật khẩu của bạn để đăng nhập vào hệ thống
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
                name="email"
                render={({ field }) => (
                  <FormItem>
                    <div className="grid gap-2">
                      <Label htmlFor="email">Email</Label>
                      <Input
                        id="email"
                        type="email"
                        placeholder="m@example.com"
                        required
                        className="border-2 border-gray-300 dark:border-gray-600"
                        {...field}
                      />
                      <FormMessage />
                    </div>
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="password"
                render={({ field }) => (
                  <FormItem>
                    <div className="grid gap-2">
                      <Label htmlFor="password">Password</Label>
                      <Input
                        id="password"
                        type="password"
                        placeholder="••••••••"
                        required
                        className="border-2 border-gray-300 dark:border-gray-600"
                        {...field}
                      />
                      <div className="text-sm text-gray-500 dark:text-gray-400 flex items-center gap-1">
                        <Info size={16} />
                        <span>
                          Password should be at least 8 characters long and
                          include a mix of letters, numbers, and symbols.
                        </span>
                      </div>
                      <FormMessage />
                    </div>
                  </FormItem>
                )}
              />
              <Button type="submit" className="w-full">
                Đăng nhập
              </Button>
              <div className="flex flex-col gap-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={handleRegisterClick}
                  className="w-full"
                >
                  Đăng ký tài khoản mới
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={handleGuestClick}
                  className="w-full"
                >
                  Đăng nhập với tư cách khách
                </Button>
              </div>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
};

export default LoginDialog1;
