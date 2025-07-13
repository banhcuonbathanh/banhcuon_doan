"use client";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from "@/components/ui/dropdown-menu";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import Link from "next/link";
import { useAuthStore } from "@/zusstand/new_auth/new_auth_controller";
import Cookies from "js-cookie";
import { Role, RoleType } from "@/constants/type";
import { useEffect, useRef } from "react";
import useOrderStore from "@/zusstand/order/order_zustand";

export default function DropdownAvatar() {
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
    openRegisterDialog,

    initializeAuthFromCookies
  } = useAuthStore();
  const { clearOrder } = useOrderStore();

  const initialized = useRef(false);

  useEffect(() => {
    // Only run initialization once
    if (!initialized.current && !isLogin) {
      initialized.current = true;
      initializeAuthFromCookies();
    }

    return () => {
      // No need to reset initialized on cleanup as the component
      // will get a new ref if it remounts
    };
  }, [isLogin, initializeAuthFromCookies]);
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
      window.location.href = "/";
    } catch (error) {
      console.error("Logout failed:", error);
    }
  };

  // Get user's role
  const getUserRole = (): RoleType => {
    if (isGuest) return Role.Guest;
    return (user?.role as RoleType) || Role.User;
  };

  // Get display name based on role and user type
  const getDisplayName = () => {
    if (!isLogin) return "Tài khoản";
    if (isGuest && guest) {
      return `Khách ${guest.name} - Bàn ${guest.table_number}`;
    }
    const role = getUserRole();
    const prefix = {
      [Role.Admin]: "Admin",
      [Role.Employee]: "NV",
      [Role.Manager]: "QL",
      [Role.User]: "",
      [Role.Guest]: "Khách"
    }[role];

    return prefix ? `${prefix} ${user?.name || ""}` : user?.name || "User";
  };

  // Get avatar initials based on role
  const getAvatarInitials = () => {
    if (!isLogin) return "G";
    if (isGuest && guest) {
      return `K${guest.table_number}`;
    }
    const role = getUserRole();
    const prefix = {
      [Role.Admin]: "A",
      [Role.Employee]: "E",
      [Role.Manager]: "M",
      [Role.User]: "",
      [Role.Guest]: "G"
    }[role];

    return prefix + (user?.name ? user.name.charAt(0).toUpperCase() : "U");
  };

  // Get role-specific menu items
  const getRoleSpecificItems = () => {
    if (!isLogin) return null;

    const role = getUserRole();
    switch (role) {
      case Role.Admin:
        return (
          <>
            <DropdownMenuItem asChild>
              <Link href="/admin/dashboard" className="cursor-pointer">
                Quản lý hệ thống
              </Link>
            </DropdownMenuItem>
            <DropdownMenuItem asChild>
              <Link href="/admin/users" className="cursor-pointer">
                Quản lý người dùng
              </Link>
            </DropdownMenuItem>
          </>
        );
      case Role.Manager:
        return (
          <DropdownMenuItem asChild>
            <Link href="/manage/dashboard" className="cursor-pointer">
              Quản lý cửa hàng
            </Link>
          </DropdownMenuItem>
        );
      case Role.Employee:
        return (
          <DropdownMenuItem asChild>
            <Link href="/employee/orders" className="cursor-pointer">
              Quản lý đơn hàng
            </Link>
          </DropdownMenuItem>
        );
      default:
        return null;
    }
  };

  // Get authentication menu items
  const getAuthMenuItems = () => {
    if (isLogin) {
      return (
        <DropdownMenuItem
          onClick={handleLogout}
          disabled={loading}
          className="text-red-600 focus:text-red-600"
        >
          {loading ? "Đang đăng xuất..." : "Đăng xuất"}
        </DropdownMenuItem>
      );
    }

    return (
      <>
        <DropdownMenuItem onClick={openLoginDialog} className="cursor-pointer">
          Đăng nhập
        </DropdownMenuItem>
        <DropdownMenuItem onClick={openGuestDialog} className="cursor-pointer">
          Đăng nhập khách
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={openRegisterDialog}
          className="cursor-pointer"
        >
          Đăng ký
        </DropdownMenuItem>
      </>
    );
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          size="icon"
          className="overflow-hidden rounded-full"
        >
          <Avatar>
            <AvatarImage
              src={isGuest ? undefined : user?.image ?? undefined}
              alt={getDisplayName()}
            />
            <AvatarFallback>{getAvatarInitials()}</AvatarFallback>
          </Avatar>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-56">
        <DropdownMenuLabel className="font-semibold">
          {getDisplayName()}
        </DropdownMenuLabel>
        <DropdownMenuSeparator />

        {/* Role-specific menu items */}
        {getRoleSpecificItems()}

        {/* Show settings and support for logged-in users only */}
        {isLogin && !isGuest && (
          <>
            {getRoleSpecificItems() && <DropdownMenuSeparator />}
            <DropdownMenuItem asChild>
              <Link href="/manage/setting" className="cursor-pointer">
                Cài đặt
              </Link>
            </DropdownMenuItem>
            <DropdownMenuItem asChild>
              <Link href="/support" className="cursor-pointer">
                Hỗ trợ
              </Link>
            </DropdownMenuItem>
            <DropdownMenuSeparator />
          </>
        )}

        {/* Authentication menu items */}
        {getAuthMenuItems()}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
