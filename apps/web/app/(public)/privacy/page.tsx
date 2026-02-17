import { Footer } from "../_components/footer";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Privacy Policy - Hookie",
  description: "Privacy Policy for Hookie webhook relay service",
};

export default function PrivacyPage() {
  return (
    <>
      <main className="container mx-auto px-4 py-12 max-w-3xl">
        <h1 className="text-3xl font-bold mb-2">Privacy Policy</h1>
        <p className="text-muted-foreground text-sm mb-10">
          Last updated: February 2025
        </p>

        <div className="prose prose-neutral dark:prose-invert max-w-none space-y-8 text-sm">
          <section>
            <h2 className="text-xl font-semibold mb-2">1. Introduction</h2>
            <p>
              Hookie (&quot;we&quot;, &quot;our&quot;, or &quot;us&quot;) operates a webhook ingestion
              and relay platform. This Privacy Policy describes how we collect,
              use, and disclose information when you use our web application,
              API, CLI, and related services (the &quot;Services&quot;).
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold mb-2">2. Information We Collect</h2>
            <p className="mb-2">We may collect:</p>
            <ul className="list-disc pl-6 space-y-1">
              <li>
                <strong>Account information:</strong> email, name, and profile
                data you provide when signing up (e.g. via our authentication
                provider);
              </li>
              <li>
                <strong>Usage data:</strong> how you use the Services (e.g.
                applications created, webhook endpoints, API and CLI usage);
              </li>
              <li>
                <strong>Webhook payloads:</strong> the content of webhooks you
                send through our ingest service, which we process and relay
                according to your configuration;
              </li>
              <li>
                <strong>Technical data:</strong> IP address, browser/device
                information, and logs necessary to operate and secure the
                Services.
              </li>
            </ul>
          </section>

          <section>
            <h2 className="text-xl font-semibold mb-2">3. How We Use Information</h2>
            <p>
              We use the information to provide, operate, and improve the
              Services; to authenticate you and manage your account; to relay
              and stream webhook events as you configure; to respond to support
              requests; and to comply with law and protect our rights. We may use
              aggregated, non-identifying data for analytics and product
              improvement.
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold mb-2">4. Sharing and Disclosure</h2>
            <p>
              We do not sell your personal information. We may share data with
              service providers that help us run the Services (e.g. hosting,
              authentication, analytics), subject to confidentiality and use
              restrictions. We may disclose information where required by law or
              to protect our rights, safety, or property.
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold mb-2">5. Data Retention</h2>
            <p>
              We retain account and usage data for as long as your account is
              active and as needed to provide the Services and comply with legal
              obligations. Webhook payload data is processed and relayed in
              real-time; retention of payload content depends on your usage and
              our operational requirements and may be described in product
              documentation.
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold mb-2">6. Security</h2>
            <p>
              We implement technical and organizational measures to protect
              your data. No method of transmission or storage is completely
              secure; you provide webhook and other data at your own risk.
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold mb-2">7. Your Rights</h2>
            <p>
              Depending on your location, you may have rights to access, correct,
              delete, or port your personal data, or to object to or restrict
              certain processing. You can manage account details through the
              Services or our authentication provider. To exercise other rights
              or ask questions, contact us at the address below.
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold mb-2">8. Changes</h2>
            <p>
              We may update this Privacy Policy from time to time. We will post
              the updated policy on this page and indicate the last updated
              date. Continued use of the Services after changes constitutes
              acceptance.
            </p>
          </section>

          <section>
            <h2 className="text-xl font-semibold mb-2">9. Contact</h2>
            <p>
              For privacy-related questions or requests, contact us at{" "}
              <a
                href="mailto:privacy@hookie.sh"
                className="text-primary underline"
              >
                privacy@hookie.sh
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
