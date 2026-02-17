import { Footer } from "../_components/footer";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Terms of Service - Hookie",
  description: "Terms of Service for Hookie webhook relay service",
};

export default function TermsPage() {
  return (
    <>
      <main className="container mx-auto px-4 py-12 max-w-3xl">
        <h1 className="text-3xl font-bold mb-2">Terms of Service</h1>
        <p className="text-muted-foreground text-sm mb-10">
          Last updated: February 2025
        </p>

        <div className="prose prose-neutral dark:prose-invert max-w-none space-y-8 text-sm">
          <section>
            <h2 className="text-xl font-semibold mb-2">1. Agreement</h2>
            <p>
              These Terms of Service (&quot;Terms&quot;) govern your access to and use of
              Hookie&apos;s web application, API, CLI, and related services (the
              &quot;Services&quot;) operated by Hookie. By accessing or using the Services,
              you agree to these Terms.
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold mb-2">2. Use of the Services</h2>
            <p>
              You may use the Services only in compliance with these Terms and
              applicable law. You are responsible for the webhook payloads and
              data you send through the Services and for ensuring that your use
              does not violate any third-party rights or applicable regulations.
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold mb-2">3. Restrictions</h2>
            <p>You may not:</p>
            <ul className="list-disc pl-6 space-y-1 mt-2">
              <li>
                Use the Services for any purpose that is to Hookie&apos;s detriment or
                commercial disadvantage;
              </li>
              <li>
                Use the Services to build, provide, or operate a competing
                webhook relay or ingestion product or service;
              </li>
              <li>
                Use the Services for competitive analysis of Hookie or the
                Services;
              </li>
              <li>
                Reverse engineer, attempt to gain unauthorized access to, or
                disrupt the Services or their infrastructure;
              </li>
              <li>
                Use the Services to transmit harmful code, spam, or illegal
                content.
              </li>
            </ul>
          </section>

          <section>
            <h2 className="text-xl font-semibold mb-2">4. Account and Data</h2>
            <p>
              You may need an account to use certain features. You are
              responsible for keeping your account credentials secure and for
              all activity under your account. We process data as described in
              our{" "}
              <a href="/privacy" className="text-primary underline">
                Privacy Policy
              </a>
              .
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold mb-2">5. Availability and Changes</h2>
            <p>
              We strive to keep the Services available but do not guarantee
              uptime. We may change, suspend, or discontinue features with
              reasonable notice where practicable. We may update these Terms;
              continued use after changes constitutes acceptance.
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold mb-2">6. Disclaimer</h2>
            <p>
              The Services are provided &quot;as is&quot; and &quot;as available&quot; without
              warranties of any kind, express or implied. We do not warrant that
              the Services will be uninterrupted, secure, or error-free.
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold mb-2">7. Limitation of Liability</h2>
            <p>
              To the maximum extent permitted by law, Hookie and its affiliates
              shall not be liable for any indirect, incidental, special,
              consequential, or punitive damages, or for loss of profits, data,
              or use, arising out of or in connection with the Services or these
              Terms.
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold mb-2">8. Contact</h2>
            <p>
              For questions about these Terms, contact us at{" "}
              <a
                href="mailto:legal@hookie.sh"
                className="text-primary underline"
              >
                legal@hookie.sh
              </a>
              .
            </p>
          </section>
        </div>
      </main>
      <Footer />
    </>
  );
}
