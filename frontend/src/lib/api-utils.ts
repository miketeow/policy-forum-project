import { env } from "@/env";
import { ActionState } from "./api";

export const API_BASE_URL = env.NEXT_PUBLIC_API_URL;

export class ApiError extends Error {
  public status: number;
  public fields?: Record<string, string>;

  constructor(
    message: string,
    status: number,
    fields?: Record<string, string>,
  ) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.fields = fields;
  }
}

export function getApiUrl(path: string) {
  // make sure no double slashes
  const base = API_BASE_URL.replace(/\/$/, "");
  const cleanPath = path.startsWith("/") ? path : `/${path}`;
  return `${base}${cleanPath}`;
}

export function handleActionError<T = undefined>(
  error: unknown,
  actionName: string,
): ActionState<T> {
  if (error instanceof ApiError) {
    console.error(`[${actionName}] API Error`, error.message);
    return {
      success: false,
      message: "Action failed",
      error: error.message,
      fields: error.fields,
    };
  }

  if (error instanceof Error) {
    console.error(`[${actionName}] Network Error`, error.message);
  } else {
    console.error(`[${actionName}]Unknown Error`, error);
  }

  // generic safe fallback
  return {
    success: false,
    message: "Connection failure",
    error: "An unexpected network error occurred. Please try again.",
  };
}
