import { Products } from "@/features/products/components/server/products";

interface PlansProps {
  hideHeader?: boolean;
}

export async function Plans({ hideHeader = false }: PlansProps = {}) {
  return (
    <section id="plans" className={`relative ${hideHeader ? '' : 'overflow-hidden'} ${hideHeader ? 'py-0' : 'py-24'}`}>
      {!hideHeader && (
        <>
          {/* Background gradient - different from features */}
          <div className="absolute inset-0 bg-gradient-to-b from-background via-muted/10 to-background" />
          <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top_right,oklch(0.6171_0.1375_39.0427/0.08),transparent_50%)]" />
        </>
      )}
      
      {/* Content */}
      <div className="relative container mx-auto px-4">
        {!hideHeader && (
          <>
            <h3 className="text-3xl md:text-4xl font-bold text-center mb-4">Simple, transparent pricing</h3>
            <p className="text-center text-lg text-muted-foreground mb-16 max-w-2xl mx-auto">
              Start free and scale as you grow. Every plan includes all features from previous tiers—no feature gating, just more capacity.
            </p>
          </>
        )}

        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
          <Products />
        </div>
      </div>
    </section>
  );
}
