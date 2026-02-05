"use client";

import {
  OrganizationSwitcher,
  UserButton,
  useOrganization,
} from "@clerk/nextjs";
import { Button } from "@hookie/ui/components/button";
import { LogoWordmark } from "@hookie/ui/components/logo-wordmark";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useRef } from "react";

export function AppHeader() {
  const pathname = usePathname();
  const { organization } = useOrganization();
  const prevOrgIdRef = useRef<string | null>(null);

  useEffect(() => {
    if (organization?.id !== prevOrgIdRef.current) {
      if (prevOrgIdRef.current !== null) {
        // Organization changed, refresh the page
        window.location.reload();
      }
      prevOrgIdRef.current = organization?.id || null;
    }
  }, [organization?.id]);

  return (
    <header className="border-b border-border sticky top-0 z-50 w-full bg-background">
      <div className="container mx-auto px-4 py-4 flex justify-between items-center">
        <div className="flex items-center gap-4">
          <Link href="/dashboard" className="mr-3">
            <LogoWordmark className="h-7 text-foreground" />
          </Link>
          <Link href="/dashboard">
            <Button variant={pathname === "/dashboard" ? "default" : "ghost"}>
              Dashboard
            </Button>
          </Link>
          <Link href="/applications">
            <Button
              variant={
                pathname?.startsWith("/applications") ? "default" : "ghost"
              }
            >
              Applications
            </Button>
          </Link>
          <Link href="/settings/connected-clients">
            <Button
              variant={pathname?.startsWith("/settings") ? "default" : "ghost"}
            >
              Settings
            </Button>
          </Link>
        </div>
        <div className="flex items-center gap-4">
          <OrganizationSwitcher
            afterSwitchOrganizationUrl={pathname || "/dashboard"}
          />
          <UserButton />
        </div>
      </div>
    </header>
  );
}
