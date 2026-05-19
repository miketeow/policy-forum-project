"use server";
import { cookies } from "next/headers";
import { ApiError, getApiUrl } from "./api-utils";

export interface ActionState<T = undefined> {
  success: boolean;
  message?: string;
  error?: string;
  fields?: Record<string, string>;
  data?: T;
}

export async function requireAuth<
  T = undefined,
>(): Promise<ActionState<T> | null> {
  const cookieStore = await cookies();
  if (!cookieStore.get("session")?.value) {
    return {
      success: false,
      message: "Authentication required",
      error: "Your session has expired. Please log in again",
    };
  }
  return null;
}

export async function fetchAPI<T = unknown>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const url = getApiUrl(path);
  const headers = new Headers(options.headers);

  // 1. auto-inject content-type for mutation
  if (
    !headers.has("Content-Type") &&
    options.method &&
    options.method !== "GET" &&
    options.method !== "DELETE"
  ) {
    headers.set("Content-Type", "application/json");
  }

  // 2. auto-inject authentication token
  const cookieStore = await cookies();
  const token = cookieStore.get("session")?.value;
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  // 3. execute fetch
  const response = await fetch(url, { cache: "no-store", ...options, headers });

  // 4. centralized error handling
  if (!response.ok) {
    let errorMessage = `API Error: ${response.status} ${response.statusText}`;
    let errorFields: Record<string, string> | undefined = undefined;
    try {
      const errorBody = await response.json();
      if (errorBody.error) errorMessage = errorBody.error;

      if (errorBody.field) errorFields = errorBody.field;
    } catch {}

    throw new ApiError(errorMessage, response.status, errorFields);
  }

  const text = await response.text();

  if (!text || text.trim() === "") {
    return {} as T;
  }

  try {
    return JSON.parse(text) as T;
  } catch {
    console.warn(
      `[fetchAPI] Failed to parse JSON for ${path}, returning raw text.`,
    );
    return text as unknown as T;
  }
}
