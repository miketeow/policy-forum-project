import z from "zod";

export const SignUpSchema = z.object({
  email: z.email({ error: "Invalid email address format" }),
  name: z
    .string({ error: () => "Name is required" })
    .min(3, { error: "Name must be at least 3 characters" })
    .max(100, { error: "Name cannot exceed 100 characters" }),
  password: z
    .string({ error: () => "Password is required" })
    .min(7, { error: "Password must be at least 8 characters" })
    .max(72, { error: "Password is too long" }),
});

export const SignInSchema = z.object({
  email: z.email({ error: "Invalid email address format" }),
  password: z.string().min(1, { error: "Password is required" }),
});
