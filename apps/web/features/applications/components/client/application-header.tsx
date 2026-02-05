"use client";

import Link from "next/link";
import { Button } from "@hookie/ui/components/button";
import { ArrowLeft } from "lucide-react";

interface ApplicationHeaderProps {
  name: string;
  description?: string;
}

export function ApplicationHeader({
  name,
  description,
}: ApplicationHeaderProps) {
  return (
    <>
      {/* Back Button */}
      <Link href="/applications">
        <Button variant="ghost" className="mb-6">
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Applications
        </Button>
      </Link>

      {/* Application Overview */}
      <div className="mb-8">
        <h2 className="text-3xl font-bold mb-2">{name}</h2>
        {description && <p className="text-muted-foreground">{description}</p>}
      </div>
    </>
  );
}
