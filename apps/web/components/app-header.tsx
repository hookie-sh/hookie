"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { OrganizationSwitcher, UserButton } from "@clerk/nextjs";
import { Button } from "@hookie/ui/components/button";

export function AppHeader() {
  const pathname = usePathname();

  return (
    <header className="border-b border-border sticky top-0 z-50 w-full bg-background">
      <div className="container mx-auto px-4 py-4 flex justify-between items-center">
        <div className="flex items-center gap-4">
          <Link href="/dashboard">
            <h1 className="text-2xl font-bold">Hookie</h1>
          </Link>
          <Link href="/dashboard">
            <Button variant={pathname === "/dashboard" ? "default" : "ghost"}>
              Dashboard
            </Button>
          </Link>
          <Link href="/applications">
            <Button variant={pathname?.startsWith("/applications") ? "default" : "ghost"}>
              Applications
            </Button>
          </Link>
          <Link href="/settings">
            <Button variant={pathname === "/settings" ? "default" : "ghost"}>
              Settings
            </Button>
          </Link>
        </div>
        <div className="flex items-center gap-4">
          <OrganizationSwitcher />
          <UserButton />
        </div>
      </div>
    </header>
  );
}

