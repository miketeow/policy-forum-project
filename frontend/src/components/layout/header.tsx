import { ThemeToggle } from "./theme-toggle";
import MainNav from "./main-nav";
import { Profile } from "./profile";
import { getSession } from "@/lib/session";
import { Button } from "../ui/button";
import Link from "next/link";

export async function Header() {
  const user = await getSession();
  console.log("User is logged in now", user);
  return (
    <header className="bg-background/75 fixed inset-x-0 top-0 z-50 py-4 backdrop-blur-md">
      <div className="container flex max-w-6xl items-center justify-between">
        {/* Left Side: Brand / Nav */}
        <MainNav />

        {/* Right Side: Interactive Client Components */}
        <div className="flex items-center gap-2">
          <ThemeToggle />
          {user ? (
            <Profile user={user} />
          ) : (
            <Button asChild className="w-full rounded-sm">
              <Link href="/sign-in">Sign In</Link>
            </Button>
          )}
        </div>
      </div>
    </header>
  );
}
