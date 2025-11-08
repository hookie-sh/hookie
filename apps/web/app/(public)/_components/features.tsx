import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@hookie/ui/components/card'
import { Webhook, Zap, Shield, Code } from 'lucide-react'

export function Features() {
  return (
    <section id="features" className="container mx-auto px-4 py-16">
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
  )
}
