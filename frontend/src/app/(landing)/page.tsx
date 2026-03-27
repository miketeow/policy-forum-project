import { Button } from "@/components/ui/button";
import { getSession } from "@/lib/session";
import Link from "next/link";

export default async function LandingPage() {
  const user = await getSession();
  return (
    <div className="flex flex-col min-h-dvh bg-background text-foreground max-w-5xl mx-auto items-center justify-center">
      <div className=" px-6 flex text-center flex-col items-center justify-center gap-8 w-full">
        <h1 className="text-4xl font-semibold tracking-tight lg:text-5xl">
          Welcome to Circle Policy
        </h1>
        <p className="text-muted-foreground text-lg max-w-xl">
          Join the discussion and shape the policies that matters
        </p>
        <div className="flex gap-4 justify-center w-full flex-col sm:flex-row max-w-sm sm:max-w-none">
          {user ? (
            <>
              <Button
                asChild
                variant="outline"
                className="w-full sm:w-40 h-12 text-base"
              >
                <Link href="/forum">Enter Forum</Link>
              </Button>
              <Button
                asChild
                className="w-full sm:w-40 bg-primary h-12 text-base"
              >
                <Link href="/dashboard">My Dashboard</Link>
              </Button>
            </>
          ) : (
            <>
              <Button
                asChild
                variant="outline"
                className="w-full sm:w-40 h-12 text-base"
              >
                <Link href="/sign-in">Sign In</Link>
              </Button>
              <Button
                asChild
                className="w-full sm:w-40 bg-primary h-12 text-base"
              >
                <Link href="/sign-up">Sign Up</Link>
              </Button>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
