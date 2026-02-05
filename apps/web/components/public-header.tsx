"use client";

import { SignedIn, SignedOut, UserButton } from "@clerk/nextjs";
import { Button } from "@hookie/ui/components/button";
import { LogoWordmark } from "@hookie/ui/components/logo-wordmark";
import Link from "next/link";

export function PublicHeader() {
  return (
    <header className="border-b border-border sticky top-0 z-50 w-full bg-background">
      <div className="container mx-auto px-4 py-4 flex justify-between items-center">
        <Link href="/" className="mr-3">
          <LogoWordmark className="h-7 text-foreground" />
        </Link>
        <div className="flex gap-4">
          <SignedOut>
            <Link href="/sign-in">
              <Button variant="ghost">Sign In</Button>
            </Link>
            <Link href="/sign-up">
              <Button>Get Started</Button>
            </Link>
          </SignedOut>
          <SignedIn>
            <Link href="/dashboard">
              <Button variant="ghost">Dashboard</Button>
            </Link>
            <UserButton />
          </SignedIn>
        </div>
      </div>
    </header>
  );
}
