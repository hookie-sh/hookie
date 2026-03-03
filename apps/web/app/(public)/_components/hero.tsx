"use client";

import { Button } from "@hookie/ui/components/button";

export function Hero() {
  const scrollToWaitlist = () => {
    const waitlistElement = document.getElementById("waitlist");
    if (waitlistElement) {
      waitlistElement.scrollIntoView({ behavior: "smooth", block: "start" });
    }
  };

  return (
    <section className="relative overflow-hidden">
      {/* Background gradient with subtle pattern - entry point */}
      <div className="absolute inset-0 bg-gradient-to-br from-background via-background to-muted/20" />
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_30%_20%,oklch(0.6171_0.1375_39.0427/0.12),transparent_50%)]" />
      <div className="absolute inset-0 bg-[linear-gradient(to_right,#80808008_1px,transparent_1px),linear-gradient(to_bottom,#80808008_1px,transparent_1px)] bg-[size:24px_24px]" />
      
      {/* Content */}
      <div className="relative container mx-auto px-4 py-32 flex flex-col items-center text-center">
        <h2 className="text-4xl md:text-6xl font-bold mb-6">
          Receive webhooks locally
        </h2>
        <p className="text-xl text-muted-foreground mb-10 max-w-3xl">
          For developers. Send webhooks to our ingest service—we relay them to your machine through the CLI. No ngrok, no port forwarding. Just run the CLI and start listening.
        </p>
        <div className="flex flex-col sm:flex-row gap-4">
          <Button size="lg" onClick={scrollToWaitlist}>
            Join Waitlist
          </Button>
        </div>
      </div>
    </section>
  );
}
