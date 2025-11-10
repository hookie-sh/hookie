import {
  Card,
  CardHeader,
  CardTitle,
  CardContent,
  CardFooter,
} from '@hookie/ui/components/card'
import { Badge } from '@hookie/ui/components/badge'
import { Check } from 'lucide-react'
import { listProducts } from '../../db/server'
import { PurchaseProduct } from '../purchase-product'

export async function Products() {
  const products = await listProducts()
  return (
    <>
      {products.map((product) => (
        <Card
          key={product.name}
          className={`flex flex-col ${
            product.highlight ? 'border-primary/20 shadow-md' : ''
          }`}
        >
          <CardHeader>
            <div className="flex items-center justify-between mb-2">
              <CardTitle className="text-2xl">{product.displayName}</CardTitle>
              {product.badge && (
                <Badge variant={product.badge.variant}>
                  {product.badge.label}
                </Badge>
              )}
            </div>
            <div className="mt-4">
              <div className="text-4xl font-bold">{product.price.display}</div>
              {product.price.monthly && (
                <div className="text-sm text-muted-foreground mt-1">
                  {product.price.monthly}
                </div>
              )}
              <div className="text-sm text-muted-foreground mt-1">
                {product.price.webhookLimit}
              </div>
            </div>
          </CardHeader>
          <CardContent className="flex-1">
            {product.previousPlanName && (
              <p className="text-xs text-muted-foreground mb-4">
                Everything in {product.previousPlanName}, plus:
              </p>
            )}
            <ul className="space-y-3">
              {product.features.map((feature, index) => (
                <li key={index} className="flex items-start gap-2">
                  <Check className="h-5 w-5 text-primary mt-0.5 shrink-0" />
                  <span className="text-sm">{feature.text}</span>
                </li>
              ))}
            </ul>
          </CardContent>
          <CardFooter>
            <PurchaseProduct product={product} />
          </CardFooter>
        </Card>
      ))}
    </>
  )
}
