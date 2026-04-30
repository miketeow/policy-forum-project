import { cookies } from "next/headers";
import { redirect } from "next/navigation";

export async function GET() {
  const cookieStore = await cookies();
  // 1. Destoy ghost token
  cookieStore.delete("session");

  // 2. Redirect to sign in
  redirect("/sign-in");
}
