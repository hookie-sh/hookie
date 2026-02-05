import { Button } from "@hookie/ui/components/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@hookie/ui/components/card";
import { Link } from "lucide-react";

export function CTA() {
  return (
    <section className="container mx-auto px-4 py-32">
      <Card className="max-w-4xl mx-auto">
        <CardHeader className="text-center">
          <CardTitle className="text-3xl md:text-4xl font-bold mb-4">
            Ready to get started?
          </CardTitle>
          <CardDescription className="text-lg">
            Join developers who trust Hookie for their webhook infrastructure.
          </CardDescription>
        </CardHeader>
        <CardContent className="text-center">
          <Link href="/sign-up">
            <Button size="lg">Create Free Account</Button>
          </Link>
        </CardContent>
      </Card>
    </section>
  );
}
