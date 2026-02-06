import { auth } from "@clerk/nextjs/server";
import { NextRequest, NextResponse } from "next/server";
import { Resend } from "resend";
import { ZodError } from "zod";
import { contactEnterpriseSchema } from "@/features/products/schemas/contact-enterprise";

const resend = new Resend(process.env.RESEND_API_KEY);

export async function POST(req: NextRequest) {
  try {
    const { userId } = await auth();

    if (!userId) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const body = await req.json();

    // Validate input
    const validatedData = contactEnterpriseSchema.parse(body);
    const { name, email, company, message } = validatedData;

    const recipientEmail =
      process.env.ENTERPRISE_CONTACT_EMAIL || "sales@hookie.dev";

    // Send email via Resend
    const { data, error } = await resend.emails.send({
      from: "Hookie <noreply@hookie.dev>",
      to: [recipientEmail],
      replyTo: email,
      subject: `Enterprise Inquiry from ${name}${company ? ` at ${company}` : ""}`,
      html: `
        <h2>New Enterprise Inquiry</h2>
        <p><strong>Name:</strong> ${name}</p>
        <p><strong>Email:</strong> ${email}</p>
        ${company ? `<p><strong>Company:</strong> ${company}</p>` : ""}
        <p><strong>User ID:</strong> ${userId}</p>
        <hr />
        <h3>Message:</h3>
        <p>${message.replace(/\n/g, "<br />")}</p>
      `,
      text: `
New Enterprise Inquiry

Name: ${name}
Email: ${email}
${company ? `Company: ${company}` : ""}
User ID: ${userId}

Message:
${message}
      `,
    });

    if (error) {
      console.error("Resend error:", error);
      return NextResponse.json(
        { error: "Failed to send email" },
        { status: 500 },
      );
    }

    return NextResponse.json({ success: true, data });
  } catch (error) {
    // Handle Zod validation errors
    if (error instanceof ZodError) {
      return NextResponse.json(
        { error: "Invalid input", details: error.issues },
        { status: 400 },
      );
    }

    // Handle other errors
    if (error instanceof Error) {
      console.error("Error sending enterprise contact email:", error);
      return NextResponse.json(
        { error: error.message || "Failed to send email" },
        { status: 500 },
      );
    }

    // Handle unknown errors
    console.error("Unknown error sending enterprise contact email:", error);
    return NextResponse.json(
      { error: "Internal server error" },
      { status: 500 },
    );
  }
}
