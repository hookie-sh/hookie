"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAuth } from "@clerk/nextjs";
import useSWR, { mutate } from "swr";
import { Button } from "@hookie/ui/components/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@hookie/ui/components/dialog";
import { Input } from "@hookie/ui/components/input";
import { Label } from "@hookie/ui/components/label";
import { ApplicationCard } from "@/components/application-card";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { createApplicationSchema, type CreateApplicationInput } from "@/data/apps/validation";
import { fetcher } from "@/utils/api";

interface Application {
  id: string;
  name: string;
  description?: string;
  topicCount: number;
}

export default function ApplicationsPage() {
  const { userId } = useAuth();
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const {
    data: applications,
    error,
    isLoading,
  } = useSWR<Application[]>(
    userId ? "/api/applications" : null,
    fetcher,
    {
      revalidateOnFocus: false,
      revalidateOnReconnect: true,
    }
  );

  const {
    control,
    handleSubmit,
    formState: { errors, isSubmitting },
    reset,
  } = useForm<CreateApplicationInput>({
    resolver: zodResolver(createApplicationSchema),
    defaultValues: {
      name: "",
      description: "",
    },
  });

  const onSubmit = async (data: CreateApplicationInput) => {
    try {
      setSubmitError(null);
      const response = await fetch("/api/applications", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || "Failed to create application");
      }

      const newApp = await response.json();
      
      // Optimistically update the cache, then revalidate
      mutate("/api/applications", (current: Application[] | undefined) => {
        return current ? [newApp, ...current] : [newApp];
      }, false);

      // Revalidate to confirm with server
      await mutate("/api/applications");

      setIsOpen(false);
      reset();
      setSubmitError(null);
    } catch (err) {
      console.error("Failed to create application:", err);
      setSubmitError(err instanceof Error ? err.message : "Failed to create application");
    }
  };

  return (
    <>
      {/* Main Content */}
      <main className="container mx-auto px-4 py-8">
        <div className="mb-8 flex justify-between items-center">
          <div>
            <h2 className="text-3xl font-bold mb-2">Applications</h2>
            <p className="text-muted-foreground">
              Manage your webhook applications and topics
            </p>
          </div>
          <Dialog open={isOpen} onOpenChange={setIsOpen}>
            <DialogTrigger asChild>
              <Button>Create Application</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Create New Application</DialogTitle>
                <DialogDescription>
                  Create a new application to start receiving webhooks
                </DialogDescription>
              </DialogHeader>
              <form onSubmit={handleSubmit(onSubmit)}>
                <div className="space-y-4 py-4">
                  <div className="space-y-2">
                    <Label htmlFor="name">Name</Label>
                    <Controller
                      name="name"
                      control={control}
                      render={({ field }) => (
                        <Input
                          id="name"
                          placeholder="My Application"
                          {...field}
                        />
                      )}
                    />
                    {errors.name && (
                      <p className="text-sm text-destructive">
                        {errors.name.message}
                      </p>
                    )}
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="description">Description (optional)</Label>
                    <Controller
                      name="description"
                      control={control}
                      render={({ field }) => (
                        <Input
                          id="description"
                          placeholder="A brief description"
                          {...field}
                        />
                      )}
                    />
                    {errors.description && (
                      <p className="text-sm text-destructive">
                        {errors.description.message}
                      </p>
                    )}
                  </div>
                </div>
                <DialogFooter>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setIsOpen(false)}
                  >
                    Cancel
                  </Button>
                  <Button type="submit" disabled={isSubmitting}>
                    {isSubmitting ? "Creating..." : "Create"}
                  </Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>
        </div>

        {/* Error Messages */}
        {(error || submitError) && (
          <div className="mb-4 p-4 bg-destructive/10 text-destructive rounded-md">
            {submitError || (error instanceof Error ? error.message : "Failed to load applications")}
          </div>
        )}

        {/* Applications Grid */}
        {isLoading ? (
          <div className="text-center py-12">
            <p className="text-muted-foreground">Loading applications...</p>
          </div>
        ) : !applications || applications.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-muted-foreground mb-4">
              No applications yet. Create your first application to get started.
            </p>
          </div>
        ) : (
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {applications.map((app) => (
              <ApplicationCard
                key={app.id}
                {...app}
                href={`/applications/${app.id}`}
              />
            ))}
          </div>
        )}
      </main>
    </>
  );
}

