import { Plans } from '@/features/subscriptions/components/server/plans'
import { Hero } from './_components/hero'
import { Features } from './_components/features'
import { CTA } from './_components/cta'
import { Footer } from './_components/footer'

export default async function Home() {
  return (
    <>
      <Hero />
      <Features />
      {/* Plans */}
      <section id="plans" className="container mx-auto px-4 py-16">
        <h3 className="text-3xl md:text-4xl font-bold text-center mb-4">
          Plans
        </h3>
        <p className="text-center text-muted-foreground mb-16 max-w-2xl mx-auto">
          Choose the perfect plan for your needs. All plans include everything
          from previous tiers plus new features.
        </p>

        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6 max-w-7xl mx-auto">
          <Plans />
        </div>
      </section>
      <CTA />
      <Footer />
    </>
  )
}
