"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAuth } from "@clerk/nextjs";
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

export default function ApplicationsPage() {
  const { userId } = useAuth();
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const [applications, setApplications] = useState<
    Array<{ id: string; name: string; description?: string; topicCount: number }>
  >([]);

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
      // TODO: Replace with actual API call
      const newApp = {
        id: Math.random().toString(36).substring(7),
        name: data.name,
        description: data.description,
        topicCount: 0,
      };
      setApplications([...applications, newApp]);
      setIsOpen(false);
      reset();
    } catch (error) {
      console.error("Failed to create application:", error);
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

        {/* Applications Grid */}
        {applications.length === 0 ? (
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

