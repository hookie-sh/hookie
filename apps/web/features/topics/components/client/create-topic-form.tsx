"use client";

import { useState } from "react";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { mutate } from "swr";
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
import { createTopicSchema, type CreateTopicInput } from "../../schemas/topic";

interface Topic {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

interface CreateTopicFormProps {
  applicationId: string;
  onSuccess?: (topic: Topic) => void;
  onError?: (error: string) => void;
}

export function CreateTopicForm({
  applicationId,
  onSuccess,
  onError,
}: CreateTopicFormProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const {
    control,
    handleSubmit,
    formState: { errors, isSubmitting },
    reset,
  } = useForm<CreateTopicInput>({
    resolver: zodResolver(createTopicSchema),
    defaultValues: {
      name: "",
      description: "",
    },
  });

  const onSubmit = async (data: CreateTopicInput) => {
    try {
      setSubmitError(null);
      const response = await fetch(
        `/api/applications/${applicationId}/topics`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(data),
        },
      );

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || "Failed to create topic");
      }

      const newTopic = await response.json();

      // Optimistically update the cache, then revalidate
      mutate(
        `/api/applications/${applicationId}/topics`,
        (current: Topic[] | undefined) => {
          return current ? [newTopic, ...current] : [newTopic];
        },
        false,
      );

      // Revalidate to confirm with server
      await mutate(`/api/applications/${applicationId}/topics`);

      // Also update the applications list to reflect new topic count
      await mutate("/api/applications");

      setIsOpen(false);
      reset();
      setSubmitError(null);
      onSuccess?.(newTopic);
    } catch (err) {
      console.error("Failed to create topic:", err);
      const errorMessage =
        err instanceof Error ? err.message : "Failed to create topic";
      setSubmitError(errorMessage);
      onError?.(errorMessage);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        <Button>Create Topic</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create New Topic</DialogTitle>
          <DialogDescription>
            Create a new topic to receive webhooks at a unique endpoint
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
              <Label htmlFor="topic-name">Name</Label>
              <Controller
                name="name"
                control={control}
                render={({ field }) => (
                  <Input
                    id="topic-name"
                    placeholder="payment-success"
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
              <Label htmlFor="topic-description">Description (optional)</Label>
              <Controller
                name="description"
                control={control}
                render={({ field }) => (
                  <Input
                    id="topic-description"
                    placeholder="Receives payment success events"
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
