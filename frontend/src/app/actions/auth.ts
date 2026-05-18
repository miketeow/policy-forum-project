"use server";

import { fetchAPI } from "@/lib/api";
import { handleActionError } from "@/lib/api-utils";
import { SignInSchema, SignUpSchema } from "@/schemas/auth";
import { revalidatePath } from "next/cache";
import { cookies } from "next/headers";
import z from "zod";

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

export async function signUpAction(data: z.infer<typeof SignUpSchema>) {
  const parsed = SignUpSchema.safeParse(data);

  if (!parsed.success) {
    return { success: false, error: parsed.error.issues[0].message };
  }

  try {
    const resData = await fetchAPI<{ message?: string }>("/api/auth/register", {
      method: "POST",
      body: JSON.stringify(parsed.data),
    });
    return {
      success: true,
      message: resData.message || "Account created successfully",
    };
  } catch (error) {
    return handleActionError(error, "signUpAction");
  }
}

export async function signInAction(data: z.infer<typeof SignInSchema>) {
  const parsed = SignInSchema.safeParse(data);
  if (!parsed.success) {
    return { success: false, error: parsed.error.issues[0].message };
  }

  try {
    const resData = await fetchAPI<{ message?: string; token: string }>(
      "/api/auth/login",
      {
        method: "POST",
        body: JSON.stringify(parsed.data),
      },
    );

    await createSession(resData.token);

    return {
      success: true,
      message: resData.message || "Successfully signed in",
    };
  } catch (error) {
    return handleActionError(error, "signInAction");
  }
}
