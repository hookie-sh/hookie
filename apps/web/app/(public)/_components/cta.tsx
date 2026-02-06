import { Waitlist } from "@clerk/nextjs";

export function CTA() {
  return (
    <section id="waitlist" className="relative overflow-hidden py-24">
      {/* Background with primary color */}
      <div className="absolute inset-0 bg-primary" />
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_50%,rgba(255,255,255,0.1),transparent_70%)]" />
      
      {/* Content */}
      <div className="relative container mx-auto px-4">
        <div className="grid md:grid-cols-2 gap-12 items-center">
          {/* Copy on the left */}
          <div className="text-left">
            <h2 className="text-3xl md:text-4xl font-bold mb-4 text-primary-foreground">
              Join the waitlist for early access
            </h2>
            <p className="text-lg text-primary-foreground/90">
              Be among the first developers to experience seamless webhook management. We'll notify you as soon as Hookie launches.
            </p>
          </div>
          
          {/* Waitlist form on the right */}
          <div className="flex justify-center md:justify-end">
            <Waitlist 
              appearance={{
                elements: {
                  rootBox: "w-full max-w-md",
                  card: "bg-background shadow-lg",
                }
              }}
            />
          </div>
        </div>
      </div>
    </section>
  );
}
