"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";

import { cn } from "@/lib/utils";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger
} from "@/components/ui/alert-dialog";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Role } from "@/constants/type";
import { RoleType } from "@/types/jwt.types";
import { useAuthStore } from "../../../auth/controller/auth-controller";
import LoginDialog from "./public-component/login-dialog";
import RegisterDialog from "./public-component/register-dialog";
import { useRouter } from "next/navigation";
import GuestLoginDialog from "./tables/[number]/guest-login-form";

const menuItems: {
  title: string;
  href: string;
  role?: RoleType[];
  hideWhenLogin?: boolean;
}[] = [
  {
    title: "Trang chủ",
    href: "/"
  },
  {
    title: "Menu",
    href: "/guest/menu",
    role: ["Guest"]
  },
  {
    title: "Đơn hàng",
    href: "/guest/orders",
    role: ["Guest"]
  },
  {
    title: "Quản lý",
    href: "/admin",
    role: ["Owner", "Employee"]
  }
];

function isValidRole(role: string): role is RoleType {
  return Object.values(Role).includes(role as RoleType);
}

export default function NavItems({ className }: { className?: string }) {
  const { user, logout: logoutAction, error, clearError } = useAuthStore();
  const router = useRouter();
  const [showError, setShowError] = useState(false);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [isErrorDialogOpen, setIsErrorDialogOpen] = useState(false);
  useEffect(() => {
    if (error) {
      setIsErrorDialogOpen(true);
    }
  }, [error]);

  useEffect(() => {
    if (error) {
      setShowError(true);
    }
  }, [error]);

  const handleDismissError = () => {
    setShowError(false);
    clearError();
  };

  // const logout = async () => {
  //   console.log("quananqr1/app/(public)/nav-items.tsx logout");
  //   try {
  //     await logoutAction();
  //     router.push("/");
  //   } catch (error: any) {
  //     console.error("Logout error:", error);
  //   }
  // };
  const logout = async () => {
    console.log("quananqr1/app/(public)/nav-items.tsx logout");
    try {
      await logoutAction();
      router.back(); // Go back to the previous page
    } catch (error: any) {
      console.error("Logout error:", error);
      setShowError(true);
    } finally {
      setIsDialogOpen(false);
    }
  };

  return (
    <>
      {menuItems.map((item) => {
        return (
          <Link href={item.href} key={item.href} className={className}>
            {item.title}
          </Link>
        );

        return null;
      })}

      {!user && (
        <>
          <LoginDialog />
          <RegisterDialog />

          <GuestLoginDialog />
        </>
      )}

      {user && (
        <AlertDialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <AlertDialogTrigger asChild>
            <div className={cn(className, "cursor-pointer")}>Đăng xuất</div>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Bạn có muốn đăng xuất không?</AlertDialogTitle>
              <AlertDialogDescription>
                Việc đăng xuất có thể làm mất đi hóa đơn của bạn
              </AlertDialogDescription>
            </AlertDialogHeader>

            <AlertDialogFooter>
              <AlertDialogCancel>Thoát</AlertDialogCancel>
              <AlertDialogAction onClick={logout}>OK</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}

      {showError && error && (
        <AlertDialog
          open={isErrorDialogOpen}
          onOpenChange={setIsErrorDialogOpen}
        >
          <AlertDialogContent className="bg-opacity-0">
            <AlertDialogHeader>
              <Alert variant="destructive" className="bg-gray-500">
                <AlertTitle>Error</AlertTitle>
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogAction onClick={handleDismissError}>
                Dismiss
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}
    </>
  );
}
