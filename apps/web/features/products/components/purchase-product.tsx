'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Button } from '@hookie/ui/components/button'
import {
  checkoutSessionSchema,
  type CheckoutSessionInput,
} from '@/data/stripe/validation'
import type { EnhancedProduct } from '../types'

export function PurchaseProduct({ product }: { product: EnhancedProduct }) {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const {
    handleSubmit,
    formState: { errors },
  } = useForm<CheckoutSessionInput>({
    resolver: zodResolver(checkoutSessionSchema),
    defaultValues: {
      priceId: product.stripePriceId || '',
    },
  })

  const onSubmit = async (data: CheckoutSessionInput) => {
    if (!data.priceId) {
      setError('Price ID is required')
      return
    }

    setIsLoading(true)
    setError(null)

    try {
      const response = await fetch('/api/stripe/checkout', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ priceId: data.priceId }),
      })

      const result = await response.json()

      if (!response.ok) {
        throw new Error(result.error || 'Failed to create checkout session')
      }

      if (result.url) {
        window.location.href = result.url
      } else {
        throw new Error('No checkout URL returned')
      }
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Failed to create checkout session'
      )
      setIsLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="w-full">
      {error && (
        <div className="mb-2 text-sm text-destructive" role="alert">
          {error}
        </div>
      )}
      {errors.priceId && (
        <div className="mb-2 text-sm text-destructive" role="alert">
          {errors.priceId.message}
        </div>
      )}
      <Button
        type="submit"
        className="w-full"
        variant={product.cta.variant || 'default'}
        disabled={isLoading || !product.stripePriceId}
      >
        {isLoading ? 'Loading...' : product.cta.label}
      </Button>
    </form>
  )
}
