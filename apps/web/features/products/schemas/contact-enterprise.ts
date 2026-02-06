import { z } from "zod";

export const contactEnterpriseSchema = z.object({
  name: z.string().min(1, "Name is required"),
  email: z.string().email("Invalid email address"),
  company: z.string().optional(),
  message: z.string().min(10, "Message must be at least 10 characters"),
});

export type ContactEnterpriseInput = z.infer<typeof contactEnterpriseSchema>;
