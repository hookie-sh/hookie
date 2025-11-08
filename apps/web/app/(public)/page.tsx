import { getStripeProducts } from '@/services/stripe.server'
import { Button } from '@hookie/ui/components/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@hookie/ui/components/card'
import { Badge } from '@hookie/ui/components/badge'
import { Check, Code, Shield, Webhook, Zap } from 'lucide-react'
import Link from 'next/link'

export default async function Home() {
  const products = await getStripeProducts()

  console.log(products)

  return (
    <>
      {/* Hero Section */}
      <section className="container mx-auto px-4 py-32 flex flex-col items-center text-center">
        <h2 className="text-4xl md:text-6xl font-bold mb-6">
          Webhook Relay for Developers
        </h2>
        <p className="text-xl text-muted-foreground mb-10 max-w-3xl">
          Create reliable webhook endpoints in seconds. Hookie helps you
          receive, inspect, and route webhooks from any service with ease.
        </p>
        <div className="flex flex-col sm:flex-row gap-4">
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
      <section id="features" className="container mx-auto px-4 py-32">
        <h3 className="text-3xl md:text-4xl font-bold text-center mb-16">
          Features
        </h3>
        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
          <Card>
            <CardHeader>
              <div className="inline-flex p-3 rounded-lg bg-primary/10 mb-4">
                <Webhook className="h-8 w-8 text-primary" />
              </div>
              <CardTitle>Easy Setup</CardTitle>
              <CardDescription>
                Create webhook endpoints in seconds. No complex configuration
                needed.
              </CardDescription>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader>
              <div className="inline-flex p-3 rounded-lg bg-primary/10 mb-4">
                <Zap className="h-8 w-8 text-primary" />
              </div>
              <CardTitle>Real-time Delivery</CardTitle>
              <CardDescription>
                Receive webhooks instantly. Built for reliability and speed.
              </CardDescription>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader>
              <div className="inline-flex p-3 rounded-lg bg-primary/10 mb-4">
                <Shield className="h-8 w-8 text-primary" />
              </div>
              <CardTitle>Secure & Private</CardTitle>
              <CardDescription>
                Your webhooks are secure and private. We never access your data.
              </CardDescription>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader>
              <div className="inline-flex p-3 rounded-lg bg-primary/10 mb-4">
                <Code className="h-8 w-8 text-primary" />
              </div>
              <CardTitle>Developer Friendly</CardTitle>
              <CardDescription>
                Simple REST API and comprehensive documentation for integration.
              </CardDescription>
            </CardHeader>
          </Card>
        </div>
      </section>

      {/* Plans Section */}
      <section id="plans" className="container mx-auto px-4 py-32">
        <h3 className="text-3xl md:text-4xl font-bold text-center mb-4">
          Plans
        </h3>
        <p className="text-center text-muted-foreground mb-16 max-w-2xl mx-auto">
          Choose the perfect plan for your needs. All plans include everything
          from previous tiers plus new features.
        </p>
        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6 max-w-7xl mx-auto">
          {/* Free Plan */}
          <Card className="flex flex-col">
            <CardHeader>
              <div className="flex items-center justify-between mb-2">
                <CardTitle className="text-2xl">Free</CardTitle>
                <Badge variant="secondary">Individual</Badge>
              </div>
              <div className="mt-4">
                <div className="text-4xl font-bold">$0</div>
                <div className="text-sm text-muted-foreground mt-1">
                  10k webhooks/month
                </div>
              </div>
            </CardHeader>
            <CardContent className="flex-1">
              <ul className="space-y-3">
                <li className="flex items-start gap-2">
                  <Check className="h-5 w-5 text-primary mt-0.5 shrink-0" />
                  <span className="text-sm">Personal organization</span>
                </li>
                <li className="flex items-start gap-2">
                  <Check className="h-5 w-5 text-primary mt-0.5 shrink-0" />
                  <span className="text-sm">7-day retention</span>
                </li>
                <li className="flex items-start gap-2">
                  <Check className="h-5 w-5 text-primary mt-0.5 shrink-0" />
                  <span className="text-sm">Basic webhook delivery</span>
                </li>
              </ul>
            </CardContent>
            <CardFooter>
              <Button className="w-full" variant="outline">
                Get Started
              </Button>
            </CardFooter>
          </Card>

          {/* Pro Plan */}
          <Card className="flex flex-col border-primary/20 shadow-md">
            <CardHeader>
              <div className="flex items-center justify-between mb-2">
                <CardTitle className="text-2xl">Pro</CardTitle>
                <Badge variant="default">Popular</Badge>
              </div>
              <div className="mt-4">
                <div className="text-4xl font-bold">$29</div>
                <div className="text-sm text-muted-foreground mt-1">
                  per month
                </div>
                <div className="text-sm text-muted-foreground mt-1">
                  1M webhooks included
                </div>
              </div>
            </CardHeader>
            <CardContent className="flex-1">
              <p className="text-xs text-muted-foreground mb-4">
                Everything in Free, plus:
              </p>
              <ul className="space-y-3">
                <li className="flex items-start gap-2">
                  <Check className="h-5 w-5 text-primary mt-0.5 shrink-0" />
                  <span className="text-sm">Unlimited apps & members</span>
                </li>
                <li className="flex items-start gap-2">
                  <Check className="h-5 w-5 text-primary mt-0.5 shrink-0" />
                  <span className="text-sm">30-day retention</span>
                </li>
                <li className="flex items-start gap-2">
                  <Check className="h-5 w-5 text-primary mt-0.5 shrink-0" />
                  <span className="text-sm">Team collaboration</span>
                </li>
                <li className="flex items-start gap-2">
                  <Check className="h-5 w-5 text-primary mt-0.5 shrink-0" />
                  <span className="text-sm">Advanced analytics</span>
                </li>
              </ul>
            </CardContent>
            <CardFooter>
              <Button className="w-full">Get Started</Button>
            </CardFooter>
          </Card>

          {/* Scale Plan */}
          <Card className="flex flex-col">
            <CardHeader>
              <div className="flex items-center justify-between mb-2">
                <CardTitle className="text-2xl">Scale</CardTitle>
                <Badge variant="secondary">Mid-market</Badge>
              </div>
              <div className="mt-4">
                <div className="text-4xl font-bold">$99</div>
                <div className="text-sm text-muted-foreground mt-1">
                  per month
                </div>
                <div className="text-sm text-muted-foreground mt-1">
                  10M webhooks included
                </div>
              </div>
            </CardHeader>
            <CardContent className="flex-1">
              <p className="text-xs text-muted-foreground mb-4">
                Everything in Pro, plus:
              </p>
              <ul className="space-y-3">
                <li className="flex items-start gap-2">
                  <Check className="h-5 w-5 text-primary mt-0.5 shrink-0" />
                  <span className="text-sm">Priority support</span>
                </li>
                <li className="flex items-start gap-2">
                  <Check className="h-5 w-5 text-primary mt-0.5 shrink-0" />
                  <span className="text-sm">Webhook signing</span>
                </li>
                <li className="flex items-start gap-2">
                  <Check className="h-5 w-5 text-primary mt-0.5 shrink-0" />
                  <span className="text-sm">Custom integrations</span>
                </li>
                <li className="flex items-start gap-2">
                  <Check className="h-5 w-5 text-primary mt-0.5 shrink-0" />
                  <span className="text-sm">Extended retention options</span>
                </li>
              </ul>
            </CardContent>
            <CardFooter>
              <Button className="w-full" variant="outline">
                Get Started
              </Button>
            </CardFooter>
          </Card>

          {/* Enterprise Plan */}
          <Card className="flex flex-col">
            <CardHeader>
              <div className="flex items-center justify-between mb-2">
                <CardTitle className="text-2xl">Enterprise</CardTitle>
                <Badge variant="secondary">Custom</Badge>
              </div>
              <div className="mt-4">
                <div className="text-4xl font-bold">Custom</div>
                <div className="text-sm text-muted-foreground mt-1">
                  Custom webhook limits
                </div>
                <div className="text-sm text-muted-foreground mt-1">
                  Volume pricing
                </div>
              </div>
            </CardHeader>
            <CardContent className="flex-1">
              <p className="text-xs text-muted-foreground mb-4">
                Everything in Scale, plus:
              </p>
              <ul className="space-y-3">
                <li className="flex items-start gap-2">
                  <Check className="h-5 w-5 text-primary mt-0.5 shrink-0" />
                  <span className="text-sm">Dedicated support</span>
                </li>
                <li className="flex items-start gap-2">
                  <Check className="h-5 w-5 text-primary mt-0.5 shrink-0" />
                  <span className="text-sm">SLA guarantee</span>
                </li>
                <li className="flex items-start gap-2">
                  <Check className="h-5 w-5 text-primary mt-0.5 shrink-0" />
                  <span className="text-sm">SSO integration</span>
                </li>
                <li className="flex items-start gap-2">
                  <Check className="h-5 w-5 text-primary mt-0.5 shrink-0" />
                  <span className="text-sm">Audit logs</span>
                </li>
                <li className="flex items-start gap-2">
                  <Check className="h-5 w-5 text-primary mt-0.5 shrink-0" />
                  <span className="text-sm">Custom contracts</span>
                </li>
              </ul>
            </CardContent>
            <CardFooter>
              <Button className="w-full" variant="outline">
                Contact Sales
              </Button>
            </CardFooter>
          </Card>
        </div>

        {/* Volume Discounts Note */}
        <div className="mt-12 text-center">
          <Card className="max-w-2xl mx-auto border-dashed">
            <CardContent className="pt-6">
              <p className="text-sm text-muted-foreground">
                <strong className="text-foreground">Volume discounts:</strong>{' '}
                15-25% off available for 100M+ webhooks/month on Scale and
                Enterprise plans. Contact us for custom pricing.
              </p>
            </CardContent>
          </Card>
        </div>
      </section>

      {/* CTA Section */}
      <section className="container mx-auto px-4 py-32">
        <Card className="max-w-4xl mx-auto">
          <CardHeader className="text-center">
            <CardTitle className="text-3xl md:text-4xl font-bold mb-4">
              Ready to get started?
            </CardTitle>
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
          <p>© 2025 Hookie. All rights reserved.</p>
        </div>
      </footer>
    </>
  )
}
