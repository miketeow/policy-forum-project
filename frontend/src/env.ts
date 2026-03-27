import { createEnv } from "@t3-oss/env-nextjs";
import * as z from "zod";

export const env = createEnv({
  // server: {
  //   DATABASE_URL: z.url(),
  // },
  client: {
    NEXT_PUBLIC_API_URL: z.string().min(1),
  },
  experimental__runtimeEnv: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL,
  },
});
