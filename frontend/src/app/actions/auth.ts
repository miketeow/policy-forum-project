"use server";

import { revalidatePath } from "next/cache";
import { cookies } from "next/headers";

export async function createSession(token: string) {
  // store the token in http-only cookie
  (await cookies()).set("session", token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    // token expires in 24 hours
    maxAge: 60 * 60 * 24,
  });

  revalidatePath("/", "layout");
}

export async function deleteSession() {
  (await cookies()).delete("session");
  revalidatePath("/", "layout");
}
