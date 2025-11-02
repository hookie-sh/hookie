import { Button } from "@hookie/ui/components/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@hookie/ui/components/card";
import { Code, Shield, Webhook, Zap } from "lucide-react";
import Link from "next/link";

export default function Home() {
  return (
    <div className="min-h-screen flex flex-col">
      {/* Header */}
      <header className="border-b">
        <div className="container mx-auto px-4 py-4 flex justify-between items-center">
          <h1 className="text-2xl font-bold">Hookie</h1>
          <div className="flex gap-4">
            <Link href="/sign-in">
              <Button variant="ghost">Sign In</Button>
            </Link>
            <Link href="/sign-up">
              <Button>Get Started</Button>
            </Link>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <section className="container mx-auto px-4 py-24 flex flex-col items-center text-center">
        <h2 className="text-5xl font-bold mb-6">
          Webhook Relay for Developers
        </h2>
        <p className="text-xl text-muted-foreground mb-8 max-w-2xl">
          Create reliable webhook endpoints in seconds. Hookie helps you
          receive, inspect, and route webhooks from any service with ease.
        </p>
        <div className="flex gap-4">
          <Link href="/sign-up">
            <Button size="lg">Start Free Trial</Button>
          </Link>
          <Link href="#features">
            <Button size="lg" variant="outline">
              Learn More
            </Button>
          </Link>
        </div>
      </section>

      {/* Features Section */}
      <section id="features" className="container mx-auto px-4 py-24">
        <h3 className="text-3xl font-bold text-center mb-12">Features</h3>
        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
          <Card>
            <CardHeader>
              <Webhook className="h-10 w-10 mb-4 text-primary" />
              <CardTitle>Easy Setup</CardTitle>
              <CardDescription>
                Create webhook endpoints in seconds. No complex configuration
                needed.
              </CardDescription>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader>
              <Zap className="h-10 w-10 mb-4 text-primary" />
              <CardTitle>Real-time Delivery</CardTitle>
              <CardDescription>
                Receive webhooks instantly. Built for reliability and speed.
              </CardDescription>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader>
              <Shield className="h-10 w-10 mb-4 text-primary" />
              <CardTitle>Secure & Private</CardTitle>
              <CardDescription>
                Your webhooks are secure and private. We never access your data.
              </CardDescription>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader>
              <Code className="h-10 w-10 mb-4 text-primary" />
              <CardTitle>Developer Friendly</CardTitle>
              <CardDescription>
                Simple REST API and comprehensive documentation for integration.
              </CardDescription>
            </CardHeader>
          </Card>
        </div>
      </section>

      {/* CTA Section */}
      <section className="container mx-auto px-4 py-24">
        <Card className="max-w-3xl mx-auto">
          <CardHeader className="text-center">
            <CardTitle className="text-3xl">Ready to get started?</CardTitle>
            <CardDescription className="text-lg">
              Join developers who trust Hookie for their webhook infrastructure.
            </CardDescription>
          </CardHeader>
          <CardContent className="text-center">
            <Link href="/sign-up">
              <Button size="lg">Create Free Account</Button>
            </Link>
          </CardContent>
        </Card>
      </section>

      {/* Footer */}
      <footer className="border-t mt-auto py-8">
        <div className="container mx-auto px-4 text-center text-sm text-muted-foreground">
          <p>© 2024 Hookie. All rights reserved.</p>
        </div>
      </footer>
    </div>
  );
}
