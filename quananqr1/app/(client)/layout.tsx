import React from "react";
import DarkModeToggle from "@/components/dark-mode-toggle";
import AuthButtons from "@/components/form/auth-dialog";

export default function Layout({
  children
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <div className="flex min-h-screen w-full flex-col bg-muted/40">
      <div className="flex flex-col sm:gap-4 sm:py-4 sm:pl-14">
        <header className="sticky top-0 z-30 flex h-14 items-center gap-4 border-b bg-background px-4 sm:static sm:h-auto sm:border-0 sm:bg-transparent sm:px-6">
          <div className="relative ml-auto flex items-center gap-4">
            <DarkModeToggle />
            <AuthButtons />
          </div>
        </header>
        {children}
      </div>
    </div>
  );
}
