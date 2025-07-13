import DarkModeToggle from "@/components/dark-mode-toggle";
import { redirect } from "next/navigation";
import DropdownAvatar from "@/components/dropdown-avatar";
import NavLinks from "./admin/admin_component/nav-links";
import { cookies } from "next/headers";
import { decodeToken } from "@/lib/utils";
import { Role } from "@/constants/type";
import React from "react";
export default function Layout({
  children
}: Readonly<{
  children: React.ReactNode;
}>) {
  const cookieStore = cookies();
  const accessToken = cookieStore.get("accessToken")?.value;

  if (!accessToken) {
    redirect("/login");
  }
  const decoded = decodeToken(accessToken);
  if (!(decoded.role === Role.Admin || decoded.role === Role.Employee)) {
    redirect("/manage/employee");
  }
  return (
    <div className="flex min-h-screen w-full flex-col bg-muted/40">
      {/* <WebSocketWrapper accessToken={accessToken} /> */}
      <p>this is admin lay out</p>
      <NavLinks />
      <div className="flex flex-col sm:gap-4 sm:py-4 sm:pl-14">
        <header className="sticky top-0 z-30 flex h-14 items-center gap-4 border-b bg-background px-4 sm:static sm:h-auto sm:border-0 sm:bg-transparent sm:px-6">
          {/* <MobileNavLinks /> */}
          <div className="relative ml-auto flex-1 md:grow-0">
            <div className="flex justify-end">
              <DarkModeToggle />
            </div>
          </div>
          <DropdownAvatar />
        </header>
        {children}
      </div>
    </div>
  );
}
