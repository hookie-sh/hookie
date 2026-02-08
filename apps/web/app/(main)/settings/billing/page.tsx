import { stripe } from "@/clients/stripe.server";
import { createSupabaseServerClient } from "@/clients/supabase.server";
import { Plans } from "@/features/subscriptions/components/server/plans";
import { getSubscriptionByOrgId } from "@/features/subscriptions/db/server";
import { auth } from "@clerk/nextjs/server";
import { Badge } from "@hookie/ui/components/badge";
import { Button } from "@hookie/ui/components/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@hookie/ui/components/card";
import Link from "next/link";
import { redirect } from "next/navigation";

export default async function BillingPage() {
  const { userId, orgId } = await auth();

  if (!userId) {
    redirect("/sign-in");
  }

  const supabase = createSupabaseServerClient();

  const subscription = orgId
    ? await getSubscriptionByOrgId(supabase, orgId)
    : null;

  let stripeSubscription = null;
  let customerPortalUrl: string | null = null;

  if (subscription?.stripe_customer_id) {
    try {
      // Get Stripe customer to check subscription status
      const customer = await stripe.customers.retrieve(
        subscription.stripe_customer_id,
        {
          expand: ["subscriptions"],
        }
      );

      if (!customer.deleted && "subscriptions" in customer) {
        const subscriptions = customer.subscriptions?.data || [];
        stripeSubscription = subscriptions[0] || null;

        // Create customer portal session if subscription exists
        if (stripeSubscription) {
          const baseUrl =
            process.env.NEXT_PUBLIC_APP_URL || "http://localhost:3000";
          const portalSession = await stripe.billingPortal.sessions.create({
            customer: subscription.stripe_customer_id,
            return_url: `${baseUrl}/settings/billing`,
          });
          customerPortalUrl = portalSession.url;
        }
      }
    } catch (error) {
      console.error("Error fetching Stripe subscription:", error);
    }
  }

  const formatDate = (date: Date | number) => {
    return new Date(date).toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  };

  return (
    <div>
      <div className="mb-8">
        <h2 className="text-3xl font-bold mb-2">Billing</h2>
        <p className="text-muted-foreground mb-4">
          Manage your organization&apos;s subscription and billing information
        </p>
      </div>

      {!subscription || !stripeSubscription ? (
        <Plans hideHeader />
      ) : (
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Current Plan</CardTitle>
                  <CardDescription>
                    Your active subscription details
                  </CardDescription>
                </div>
                <Badge variant="default">Active</Badge>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <div className="text-sm text-muted-foreground mb-1">
                  Subscription Status
                </div>
                <div className="font-medium capitalize">
                  {stripeSubscription.status}
                </div>
              </div>

              {/* {stripeSubscription.current_period_start && (
                <div>
                  <div className="text-sm text-muted-foreground mb-1">
                    Current Period
                  </div>
                  <div className="font-medium">
                    {formatDate(stripeSubscription.current_period_start * 1000)}{" "}
                    - {formatDate(stripeSubscription.current_period_end * 1000)}
                  </div>
                </div>
              )} */}

              {stripeSubscription.cancel_at_period_end && (
                <div>
                  <Badge variant="secondary">
                    Subscription will cancel at period end
                  </Badge>
                </div>
              )}

              {customerPortalUrl && (
                <div className="pt-4">
                  <a
                    href={customerPortalUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <Button variant="outline" className="w-full md:w-auto">
                      Manage Subscription
                    </Button>
                  </a>
                </div>
              )}
            </CardContent>
          </Card>

          {!customerPortalUrl && (
            <Card>
              <CardHeader>
                <CardTitle>Upgrade or Change Plan</CardTitle>
                <CardDescription>
                  View available plans and upgrade your subscription
                </CardDescription>
              </CardHeader>
              <CardContent>
                <Link href="/paywall">
                  <Button variant="outline">View Plans</Button>
                </Link>
              </CardContent>
            </Card>
          )}
        </div>
      )}
    </div>
  );
}
