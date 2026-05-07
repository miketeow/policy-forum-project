"use server";

import { getApiUrl } from "@/lib/api";
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
    const res = await fetch(getApiUrl("/api/auth/register"), {
      method: "POST",
      headers: { "Content-Type": "applicatio/json" },
      body: JSON.stringify(parsed.data),
    });

    if (!res.ok) {
      const errData = await res.json();

      if (res.status === 422 && errData.fields) {
        return {
          success: false,
          error: "validation failed",
          fields: errData.fields,
        };
      }
      return {
        success: false,
        error: errData.error || errData.message || "Registration failed",
      };
    }

    const resData = await res.json();
    return {
      success: true,
      message: resData.message || "Account created successfully",
    };
  } catch (error) {
    console.error("SignUp Network Error:", error);
    return { success: false, error: "Unable to connect to server" };
  }
}

export async function signInAction(data: z.infer<typeof SignInSchema>) {
  const parsed = SignInSchema.safeParse(data);
  if (!parsed.success) {
    return { success: false, error: parsed.error.issues[0].message };
  }

  try {
    const res = await fetch(getApiUrl("/api/auth/login"), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(parsed.data),
    });

    if (!res.ok) {
      const errData = await res.json();

      if (res.status === 422 && errData.fields) {
        return {
          success: false,
          error: "Validation failed",
          fields: errData.fields,
        };
      }

      return {
        success: false,
        error: errData.error || errData.message || "Invalid credentials",
      };
    }

    const resData = await res.json();
    await createSession(resData.token);

    return {
      success: true,
      message: resData.message || "Successfully signed in",
    };
  } catch (error) {
    console.error("SignIn Network Error:", error);
    return { success: false, error: "Unable to connect to the server." };
  }
}
