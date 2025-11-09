import { getStripeProducts } from '@/clients/stripe.server'
import { enhancePlansWithStripe } from '@/data/stripe/plans'
import {
  Card,
  CardHeader,
  CardTitle,
  CardContent,
  CardFooter,
} from '@hookie/ui/components/card'
import { Badge } from '@hookie/ui/components/badge'
import { Check } from 'lucide-react'
import { PurchasePlan } from '../purchase-plan'

export async function Plans() {
  const stripeProducts = await getStripeProducts()
  const plans = enhancePlansWithStripe(stripeProducts)
  return (
    <>
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
            <PurchasePlan plan={plan} />
          </CardFooter>
        </Card>
      ))}
    </>
  )
}
