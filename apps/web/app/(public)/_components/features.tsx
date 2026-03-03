import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@hookie/ui/components/card";
import { Code, Shield, Webhook, Zap } from "lucide-react";
import { WebhookFlow } from "./webhook-flow";

export function Features() {
  return (
    <section id="features" className="relative overflow-hidden py-24">
      {/* Background gradient with subtle pattern - distinct from hero */}
      <div className="absolute inset-0 bg-linear-to-b from-background via-muted/15 to-muted/5" />
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top_left,oklch(0.6171_0.1375_39.0427/0.06),transparent_50%)]" />

      {/* Content */}
      <div className="relative container mx-auto px-4">
        <div className="text-center mb-12">
          <h3 className="text-3xl md:text-4xl font-bold mb-4">
            Ingest, relay, listen
          </h3>
          <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
            Webhooks hit our ingest service, get relayed in real time, and land
            on your machine when you run the CLI. Built for local development.
          </p>
        </div>
        <div className="mb-16 max-w-7xl mx-auto">
          <WebhookFlow />
        </div>
        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
          <Card className="group border-border/50 hover:border-primary/20 transition-all duration-300 hover:shadow-lg">
            <CardHeader>
              <div className="inline-flex items-center justify-center p-3 rounded-lg bg-muted group-hover:bg-primary transition-colors mb-4 w-fit">
                <Webhook className="h-6 w-6 text-primary group-hover:text-primary-foreground transition-colors" />
              </div>
              <CardTitle className="mb-2">Ingest</CardTitle>
              <CardDescription>
                Create a topic for each service you want to receive—Stripe,
                GitHub, your own APIs.
              </CardDescription>
            </CardHeader>
          </Card>
          <Card className="group border-border/50 hover:border-primary/20 transition-all duration-300 hover:shadow-lg">
            <CardHeader>
              <div className="inline-flex items-center justify-center p-3 rounded-lg bg-muted group-hover:bg-primary transition-colors mb-4 w-fit">
                <Zap className="h-6 w-6 text-primary group-hover:text-primary-foreground transition-colors" />
              </div>
              <CardTitle className="mb-2">Relay</CardTitle>
              <CardDescription>
                Real-time delivery through our relay, no polling, no delay.
              </CardDescription>
            </CardHeader>
          </Card>
          <Card className="group border-border/50 hover:border-primary/20 transition-all duration-300 hover:shadow-lg">
            <CardHeader>
              <div className="inline-flex items-center justify-center p-3 rounded-lg bg-muted group-hover:bg-primary transition-colors mb-4 w-fit">
                <Code className="h-6 w-6 text-primary group-hover:text-primary-foreground transition-colors" />
              </div>
              <CardTitle className="mb-2">CLI</CardTitle>
              <CardDescription>
                Run our CLI and listen to the topics you need. Inspect, debug,
                and develop locally without exposing your dev environment.
              </CardDescription>
            </CardHeader>
          </Card>
          <Card className="group border-border/50 hover:border-primary/20 transition-all duration-300 hover:shadow-lg">
            <CardHeader>
              <div className="inline-flex items-center justify-center p-3 rounded-lg bg-muted group-hover:bg-primary transition-colors mb-4 w-fit">
                <Shield className="h-6 w-6 text-primary group-hover:text-primary-foreground transition-colors" />
              </div>
              <CardTitle className="mb-2">For developers</CardTitle>
              <CardDescription>
                No ngrok, no tunnels, no public URLs. Your local setup stays
                private. Get webhooks where you need them on your machine.
              </CardDescription>
            </CardHeader>
          </Card>
        </div>
      </div>
    </section>
  );
}
