"use client";

import Link from "next/link";

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
import { cn, handleErrorApi } from "@/lib/utils";

import { Role } from "@/constants/type";
import { RoleType } from "@/types/jwt.types";
import { useRouter } from "next/navigation";
import { useAuth } from "../../../auth/useauth";
import LoginDialog from "../../components/form/login-dialog";
import RegisterDialog from "../../components/form/register-dialog";

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
    href: "/manage/dashboard",
    role: ["Owner", "Employee"]
  }
];

// Type guard function
function isValidRole(role: string): role is RoleType {
  return Object.values(Role).includes(role as RoleType);
}

export default function NavItems({ className }: { className?: string }) {
  const { user, logout: logoutAction, login } = useAuth();
  const router = useRouter();

  const logout = async () => {
    try {
      await logoutAction();
      router.push("/");
    } catch (error: any) {
      handleErrorApi({
        error
      });
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
      })}
      <LoginDialog />
      <RegisterDialog />

      <AlertDialog>
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
    </>
  );
}

{
  /* {user && (
        <AlertDialog>
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
      )} */
}
