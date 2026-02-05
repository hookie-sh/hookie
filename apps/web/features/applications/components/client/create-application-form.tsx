"use client";

import { zodResolver } from "@hookform/resolvers/zod";
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
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { mutate } from "swr";
import {
  createApplicationSchema,
  type CreateApplicationInput,
} from "../../schemas/application";

interface Application {
  id: string;
  name: string;
  description?: string;
  topicCount: number;
}

interface CreateApplicationFormProps {
  onSuccess?: (application: Application) => void;
  onError?: (error: string) => void;
}

export function CreateApplicationForm({
  onSuccess,
  onError,
}: CreateApplicationFormProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

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
      mutate(
        "/api/applications",
        (current: Application[] | undefined) => {
          return current ? [newApp, ...current] : [newApp];
        },
        false
      );

      // Revalidate to confirm with server
      await mutate("/api/applications");

      setIsOpen(false);
      reset();
      setSubmitError(null);
      onSuccess?.(newApp);
    } catch (err) {
      console.error("Failed to create application:", err);
      const errorMessage =
        err instanceof Error ? err.message : "Failed to create application";
      setSubmitError(errorMessage);
      onError?.(errorMessage);
    }
  };

  return (
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
            {submitError && (
              <div className="p-3 bg-destructive/10 text-destructive rounded-md text-sm">
                {submitError}
              </div>
            )}
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Controller
                name="name"
                control={control}
                render={({ field }) => (
                  <Input
                    id="name"
                    placeholder="My Application"
                    data-1p-ignore
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
  );
}
