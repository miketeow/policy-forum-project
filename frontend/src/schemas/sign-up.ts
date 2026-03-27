import z from "zod";

export const SignUpSchema = z.object({
  email: z.email(),
  name: z.string().min(3),
  password: z.string().min(6),
});
