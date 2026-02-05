"use client";

import { CreateApplicationForm } from "@/features/applications/components/client/create-application-form";
import { ApplicationsList } from "@/features/applications/components/client/applications-list";

export default function ApplicationsPage() {
  return (
    <main className="container mx-auto px-4 py-8">
      <div className="mb-8 flex justify-between items-center">
        <div>
          <h2 className="text-3xl font-bold mb-2">Applications</h2>
          <p className="text-muted-foreground">
            Manage your webhook applications and topics
          </p>
        </div>
        <CreateApplicationForm />
      </div>

      <ApplicationsList />
    </main>
  );
}
