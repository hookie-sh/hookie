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
