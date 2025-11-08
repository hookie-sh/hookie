import { getStripeProducts } from '@/services/stripe.server'
import { enhancePlansWithStripe } from '@/data/plans'
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
import { Plans } from './_components/plans'
import { Hero } from './_components/hero'
import { Features } from './_components/features'
import { CTA } from './_components/cta'
import { Footer } from './_components/footer'

export default async function Home() {
  return (
    <>
      <Hero />
      <Features />
      <Plans />
      <CTA />
      <Footer />
    </>
  )
}
