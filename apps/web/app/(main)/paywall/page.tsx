import { Plans } from "@/features/subscriptions/components/server/plans";

export default async function PaywallPage() {
  return (
    <div className="min-h-screen">
      <Plans />
    </div>
  );
}
