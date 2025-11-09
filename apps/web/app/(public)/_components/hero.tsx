import { Button } from '@hookie/ui/components/button'
import Link from 'next/link'

export function Hero() {
  return (
    <section className="container mx-auto px-4 py-32 flex flex-col items-center text-center">
      <h2 className="text-4xl md:text-6xl font-bold mb-6">
        Webhook Relay for Developers
      </h2>
      <p className="text-xl text-muted-foreground mb-10 max-w-3xl">
        Create reliable webhook endpoints in seconds. Hookie helps you receive,
        inspect, and route webhooks from any service with ease.
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
  )
}
