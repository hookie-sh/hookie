"use client";

import { useState } from "react";
import { useAuth } from "@clerk/nextjs";
import { useRouter, usePathname } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@hookie/ui/components/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@hookie/ui/components/dialog";
import {
  checkoutSessionSchema,
  type CheckoutSessionInput,
} from "../schemas/checkout-session";
import type { EnhancedProduct } from "../types";
import { EnterpriseContactForm } from "./enterprise-contact-form";

export function PurchaseProduct({ product }: { product: EnhancedProduct }) {
  const { isSignedIn, isLoaded } = useAuth();
  const router = useRouter();
  const pathname = usePathname();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isEnterpriseDialogOpen, setIsEnterpriseDialogOpen] = useState(false);
  const isPaywallPage = pathname === "/paywall";
  const isEnterprise = product.name.toLowerCase() === "enterprise";

  const {
    handleSubmit,
    formState: { errors },
  } = useForm<CheckoutSessionInput>({
    resolver: zodResolver(checkoutSessionSchema),
    defaultValues: {
      priceId: product.stripePriceId || "",
      returnUrl: pathname ?? "/",
    },
  });

  const handleClick = async () => {
    if (!isLoaded) {
      return;
    }

    // Enterprise plan: open dialog
    if (isEnterprise) {
      setIsEnterpriseDialogOpen(true);
      return;
    }

    if (!isSignedIn) {
      // Redirect to sign-up with redirect to paywall
      router.push(`/sign-up?redirect=${encodeURIComponent("/paywall")}`);
      return;
    }

    // If logged in and not Enterprise, proceed with checkout
    handleSubmit(onSubmit)();
  };

  const onSubmit = async (data: CheckoutSessionInput) => {
    if (!data.priceId || !product.stripePriceId) {
      setError("This plan is not available for purchase at the moment. Please contact support.");
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const response = await fetch("/api/stripe/checkout", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          priceId: data.priceId,
          returnUrl: data.returnUrl || pathname || "/",
        }),
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || "Failed to create checkout session");
      }

      if (result.url) {
        window.location.href = result.url;
      } else {
        throw new Error("No checkout URL returned");
      }
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to create checkout session",
      );
      setIsLoading(false);
    }
  };

  return (
    <>
      <div className="w-full">
        {error && (
          <div className="mb-2 text-sm text-destructive" role="alert">
            {error}
          </div>
        )}
        {errors.priceId && (
          <div className="mb-2 text-sm text-destructive" role="alert">
            {errors.priceId.message}
          </div>
        )}
        <Button
          onClick={handleClick}
          className="w-full"
          variant={product.cta.variant || "default"}
          disabled={!isLoaded || isLoading}
        >
          {isLoading ? "Loading..." : product.cta.label}
        </Button>
      </div>

      {isEnterprise && (
        <Dialog open={isEnterpriseDialogOpen} onOpenChange={setIsEnterpriseDialogOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Contact Sales</DialogTitle>
              <DialogDescription>
                Get in touch with our team to discuss Enterprise pricing and
                features.
              </DialogDescription>
            </DialogHeader>
            <EnterpriseContactForm
              onSuccess={() => setIsEnterpriseDialogOpen(false)}
            />
          </DialogContent>
        </Dialog>
      )}
    </>
  );
}
