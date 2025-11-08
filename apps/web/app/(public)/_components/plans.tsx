import { getStripeProducts } from '@/services/stripe.server'
import { enhancePlansWithStripe } from '@/data/plans'
import {
  Card,
  CardHeader,
  CardTitle,
  CardContent,
  CardFooter,
} from '@hookie/ui/components/card'
import { Badge } from '@hookie/ui/components/badge'
import { Check } from 'lucide-react'
import { Button } from '@hookie/ui/components/button'

export async function Plans() {
  const stripeProducts = await getStripeProducts()
  const plans = enhancePlansWithStripe(stripeProducts)
  return (
    <section id="plans" className="container mx-auto px-4 py-16">
      <h3 className="text-3xl md:text-4xl font-bold text-center mb-4">Plans</h3>
      <p className="text-center text-muted-foreground mb-16 max-w-2xl mx-auto">
        Choose the perfect plan for your needs. All plans include everything
        from previous tiers plus new features.
      </p>

      <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6 max-w-7xl mx-auto">
        {plans.map((plan) => (
          <Card
            key={plan.name}
            className={`flex flex-col ${
              plan.highlight ? 'border-primary/20 shadow-md' : ''
            }`}
          >
            <CardHeader>
              <div className="flex items-center justify-between mb-2">
                <CardTitle className="text-2xl">{plan.displayName}</CardTitle>
                {plan.badge && (
                  <Badge variant={plan.badge.variant}>{plan.badge.label}</Badge>
                )}
              </div>
              <div className="mt-4">
                <div className="text-4xl font-bold">{plan.price.display}</div>
                {plan.price.monthly && (
                  <div className="text-sm text-muted-foreground mt-1">
                    {plan.price.monthly}
                  </div>
                )}
                <div className="text-sm text-muted-foreground mt-1">
                  {plan.price.webhookLimit}
                </div>
              </div>
            </CardHeader>
            <CardContent className="flex-1">
              {plan.previousPlanName && (
                <p className="text-xs text-muted-foreground mb-4">
                  Everything in {plan.previousPlanName}, plus:
                </p>
              )}
              <ul className="space-y-3">
                {plan.features.map((feature, index) => (
                  <li key={index} className="flex items-start gap-2">
                    <Check className="h-5 w-5 text-primary mt-0.5 shrink-0" />
                    <span className="text-sm">{feature.text}</span>
                  </li>
                ))}
              </ul>
            </CardContent>
            <CardFooter>
              <Button
                className="w-full"
                variant={plan.cta.variant || 'default'}
              >
                {plan.cta.label}
              </Button>
            </CardFooter>
          </Card>
        ))}
      </div>

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
  )
}
