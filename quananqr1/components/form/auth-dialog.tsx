"use client";

import React from "react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from "@/components/ui/dropdown-menu";

import {
  User,
  LogOut,
  Settings,
  ShieldCheck,
  ShoppingCart,
  UserCircle2
} from "lucide-react";
import { useAuthStore } from "@/zusstand/new_auth/new_auth_controller";
import Cookies from "js-cookie";
import useOrderStore from "@/zusstand/order/order_zustand";
import { useRouter } from "next/navigation";

import LoginDialog from "./login-dialog";
import GuestLoginDialog from "./guest-dialog";
import RegisterDialog from "./register-dialog";

const AuthButtons = () => {
  const router = useRouter();
  const {
    user,
    guest,
    logout,
    guestLogout,
    loading,
    isGuest,
    isLogin,
    openLoginDialog,
    openGuestDialog,
    openRegisterDialog
  } = useAuthStore();
  const { clearOrder } = useOrderStore();

  const handleLogout = async () => {
    try {
      if (isGuest && guest) {
        await guestLogout({
          refresh_token: Cookies.get("refreshToken") || ""
        });
        clearOrder();
      } else {
        await logout();
        clearOrder();
      }
      router.push("/");
    } catch (error) {
      console.error("Logout failed:", error);
    }
  };

  const navigateToPage = (path: string) => {
    router.push(path);
  };

  // Custom user icon component
  const UserTrigger = () => (
    <div className="flex items-center gap-2 p-2 hover:bg-gray-100 rounded-md cursor-pointer">
      <UserCircle2 className="w-8 h-8 text-gray-600" />
      <div className="flex flex-col">
        <span className="font-medium text-sm">{user?.name || "User"}</span>
        <span className="text-xs text-gray-500">
          {user?.email || "user@example.com"}
        </span>
      </div>
    </div>
  );

  // Render buttons based on login state
  const renderAuthComponent = () => {
    if (isLogin) {
      return (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <div>
              <UserTrigger />
            </div>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="w-64">
            <DropdownMenuLabel>
              <div className="flex flex-col">
                <span className="font-medium">{user?.name || "User"}</span>
                <span className="text-xs text-gray-500">
                  {user?.email || "user@example.com"}
                </span>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => navigateToPage("/profile")}
              className="cursor-pointer"
            >
              <User className="mr-2 h-4 w-4" />
              <span>Hồ sơ cá nhân</span>
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => navigateToPage("/orders")}
              className="cursor-pointer"
            >
              <ShoppingCart className="mr-2 h-4 w-4" />
              <span>Đơn hàng</span>
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => navigateToPage("/settings")}
              className="cursor-pointer"
            >
              <Settings className="mr-2 h-4 w-4" />
              <span>Cài đặt</span>
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => navigateToPage("/security")}
              className="cursor-pointer"
            >
              <ShieldCheck className="mr-2 h-4 w-4" />
              <span>Bảo mật</span>
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={handleLogout}
              className="cursor-pointer text-red-600"
            >
              <LogOut className="mr-2 h-4 w-4" />
              <span>Đăng xuất</span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      );
    }

    return (
      <div className="flex gap-2">
        <Button onClick={openLoginDialog}>Đăng nhập</Button>
        <Button variant="secondary" onClick={openGuestDialog}>
          Đăng nhập khách
        </Button>
        <Button variant="outline" onClick={openRegisterDialog}>
          Đăng ký
        </Button>
      </div>
    );
  };

  return (
    <>
      {renderAuthComponent()}

      {/* Authentication Dialogs */}
      <LoginDialog />
      <GuestLoginDialog />
      <RegisterDialog />
    </>
  );
};

export default AuthButtons;
