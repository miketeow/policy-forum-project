import { env } from "@/env";

export const API_BASE_URL = env.NEXT_PUBLIC_API_URL;

export function getApiUrl(path: string) {
  // make sure no double slashes
  const base = API_BASE_URL.replace(/\/$/, "");
  const cleanPath = path.startsWith("/") ? path : `/${path}`;
  return `${base}${cleanPath}`;
}
