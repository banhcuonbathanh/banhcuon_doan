"use client";

import React from "react";
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
import {
  RegisterBodyType,
  RegisterBody
} from "@/schemaValidations/auth.schema";
import { useAuthStore } from "@/zusstand/new_auth/new_auth_controller";
import { handleErrorApi } from "@/lib/utils";
import { toast } from "@/hooks/use-toast";

const RegisterDialog = () => {
  const {
    register,
    isRegisterDialogOpen,
    closeRegisterDialog,
    openLoginDialog,
    openGuestDialog
  } = useAuthStore();

  const form = useForm<RegisterBodyType>({
    resolver: zodResolver(RegisterBody),
    defaultValues: {
      name: "",
      email: "",
      password: "",
      role: "",
      phone: "",
      image: "",
      address: "",
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    }
  });

  const onSubmit = async (data: RegisterBodyType) => {
    try {
      await register({
        name: data.name,
        email: data.email,
        password: data.password,
        role: "Guest",
        phone: data.phone,
        image: data.image,
        address: data.address,
        created_at: data.created_at,
        updated_at: data.updated_at
      });
      toast({
        title: "dang ki thanh cong",
        description: "Friday, February 10, 2023 at 5:57 PM"
      });
      closeRegisterDialog();
      openLoginDialog();
    } catch (error: any) {
      handleErrorApi({
        error,
        setError: form.setError
      });
    }
  };

  const handleLoginClick = () => {
    closeRegisterDialog();
    openLoginDialog();
  };

  const handleGuestClick = () => {
    closeRegisterDialog();
    openGuestDialog();
  };

  return (
    <Dialog
      open={isRegisterDialogOpen}
      onOpenChange={(open) => !open && closeRegisterDialog()}
    >
      <DialogContent className="sm:max-w-[425px] bg-white dark:bg-gray-800 shadow-lg">
        <DialogHeader>
          <DialogTitle className="text-2xl font-semibold">Đăng ký</DialogTitle>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Điền thông tin của bạn để tạo tài khoản mới
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
                      <Label htmlFor="name">Họ và tên</Label>
                      <Input
                        id="name"
                        type="text"
                        placeholder="Alice Johnson"
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
                name="email"
                render={({ field }) => (
                  <FormItem>
                    <div className="grid gap-2">
                      <Label htmlFor="email">Email</Label>
                      <Input
                        id="email"
                        type="email"
                        placeholder="alice.johnson@example.com"
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
                      <Label htmlFor="password">Mật khẩu</Label>
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
                          Mật khẩu phải có ít nhất 8 ký tự và bao gồm chữ cái,
                          số và ký tự đặc biệt.
                        </span>
                      </div>
                      <FormMessage />
                    </div>
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="phone"
                render={({ field }) => (
                  <FormItem>
                    <div className="grid gap-2">
                      <Label htmlFor="phone">Số điện thoại</Label>
                      <Input
                        id="phone"
                        type="tel"
                        placeholder="1234567890"
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
                name="address"
                render={({ field }) => (
                  <FormItem>
                    <div className="grid gap-2">
                      <Label htmlFor="address">Địa chỉ</Label>
                      <Input
                        id="address"
                        type="text"
                        placeholder="123 Main St, Anytown, USA"
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
                Đăng ký
              </Button>
              <div className="flex flex-col gap-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={handleLoginClick}
                  className="w-full"
                >
                  Đã có tài khoản? Đăng nhập
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

export default RegisterDialog;
