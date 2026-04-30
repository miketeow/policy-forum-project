import { cookies } from "next/headers";
import { cache } from "react";
import { getApiUrl } from "./api";

export const getSession = cache(async () => {
  const cookieStore = await cookies();
  const token = cookieStore.get("session")?.value;

  if (!token) return null;

  try {
    const res = await fetch(getApiUrl("/api/users/me"), {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      cache: "no-store",
    });
    if (!res.ok) {
      // BEST PRACTICE: Treat 404 (Deleted User) identically to 401 (Invalid Token)
      if (res.status === 401 || res.status === 403 || res.status === 404) {
        return null;
      }

      console.error(`[Auth] Backend error checking session: ${res.status}`);
      return null;
    }
    const user = await res.json();
    return user;
  } catch (error) {
    console.error("[Auth] Fatal error connecting to Go backend", error);
    return null;
  }
});
