"use client";

import { SignUp } from "@clerk/nextjs";
import { useSearchParams } from "next/navigation";

export default function SignUpPage() {
  const searchParams = useSearchParams();
  const redirectUrl = searchParams.get("redirect");

  return (
    <div className="flex min-h-screen items-center justify-center">
      <SignUp
        fallbackRedirectUrl={redirectUrl || "/dashboard"}
        forceRedirectUrl={redirectUrl || undefined}
      />
    </div>
  );
}
