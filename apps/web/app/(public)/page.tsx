// import { Plans } from "@/features/subscriptions/components/server/plans";
import { CTA } from "./_components/cta";
import { Features } from "./_components/features";
import { Footer } from "./_components/footer";
import { Hero } from "./_components/hero";

export default async function Home() {
  return (
    <>
      <Hero />
      <Features />
      {/* <Plans /> */}
      <CTA />
      <Footer />
    </>
  );
}
