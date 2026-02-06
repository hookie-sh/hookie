import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@hookie/ui/components/card";
import { Webhook, Zap, Shield, Code } from "lucide-react";

export function Features() {
  return (
    <section id="features" className="relative overflow-hidden py-24">
      {/* Background gradient with subtle pattern - distinct from hero */}
      <div className="absolute inset-0 bg-gradient-to-b from-background via-muted/15 to-muted/5" />
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top_left,oklch(0.6171_0.1375_39.0427/0.06),transparent_50%)]" />
      
      {/* Content */}
      <div className="relative container mx-auto px-4">
        <div className="text-center mb-16">
          <h3 className="text-3xl md:text-4xl font-bold mb-4">
            Built for developers, designed for scale
          </h3>
          <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
            Everything you need to build, debug, and scale your webhook infrastructure without the complexity
          </p>
        </div>
        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
          <Card className="group border-border/50 hover:border-primary/20 transition-all duration-300 hover:shadow-lg">
            <CardHeader>
              <div className="inline-flex items-center justify-center p-3 rounded-lg bg-muted group-hover:bg-primary transition-colors mb-4 w-fit">
                <Webhook className="h-6 w-6 text-primary group-hover:text-primary-foreground transition-colors" />
              </div>
              <CardTitle className="mb-2">Instant Setup</CardTitle>
              <CardDescription>
                Get your webhook endpoints up and running in seconds. No infrastructure to manage, no configuration headaches.
              </CardDescription>
            </CardHeader>
          </Card>
          <Card className="group border-border/50 hover:border-primary/20 transition-all duration-300 hover:shadow-lg">
            <CardHeader>
              <div className="inline-flex items-center justify-center p-3 rounded-lg bg-muted group-hover:bg-primary transition-colors mb-4 w-fit">
                <Zap className="h-6 w-6 text-primary group-hover:text-primary-foreground transition-colors" />
              </div>
              <CardTitle className="mb-2">Lightning Fast</CardTitle>
              <CardDescription>
                Sub-millisecond delivery with automatic retries. Built to handle millions of webhooks without breaking a sweat.
              </CardDescription>
            </CardHeader>
          </Card>
          <Card className="group border-border/50 hover:border-primary/20 transition-all duration-300 hover:shadow-lg">
            <CardHeader>
              <div className="inline-flex items-center justify-center p-3 rounded-lg bg-muted group-hover:bg-primary transition-colors mb-4 w-fit">
                <Shield className="h-6 w-6 text-primary group-hover:text-primary-foreground transition-colors" />
              </div>
              <CardTitle className="mb-2">Enterprise Security</CardTitle>
              <CardDescription>
                End-to-end encryption, webhook signing, and complete data privacy. Your webhooks stay yours, always.
              </CardDescription>
            </CardHeader>
          </Card>
          <Card className="group border-border/50 hover:border-primary/20 transition-all duration-300 hover:shadow-lg">
            <CardHeader>
              <div className="inline-flex items-center justify-center p-3 rounded-lg bg-muted group-hover:bg-primary transition-colors mb-4 w-fit">
                <Code className="h-6 w-6 text-primary group-hover:text-primary-foreground transition-colors" />
              </div>
              <CardTitle className="mb-2">Developer First</CardTitle>
              <CardDescription>
                Clean REST API, webhook replay, request inspection, and comprehensive docs. Built by developers, for developers.
              </CardDescription>
            </CardHeader>
          </Card>
        </div>
      </div>
    </section>
  );
}
