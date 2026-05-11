import { getSession } from "@/lib/session";
import { redirect } from "next/navigation";
import { CreatePostForm } from "./_components/create-post-form";
import { PostList } from "./_components/post-list";
import { cookies } from "next/headers";
import { fetchPostAction } from "../actions/forum";
import { Post } from "./_components/post-card";
import { Button } from "@/components/ui/button";
import Link from "next/link";

interface UserProfile {
  id: string;
  name: string;
  email: string;
  kyc_status: string;
  created_at: string;
  updated_at: string;
}

export default async function Forum({
  searchParams,
}: {
  searchParams: Promise<{ sort?: string }>;
}) {
  const user: UserProfile | null = await getSession();

  const cookieStore = await cookies();
  const hasCookie = cookieStore.has("session");

  if (!user && hasCookie) {
    redirect("/api/auth/logout");
  }

  const resolvedParams = await searchParams;
  let sort: "desc" | "asc" | "popular" = "desc";
  if (resolvedParams.sort === "asc") sort = "asc";
  if (resolvedParams.sort === "popular") sort = "popular";

  let posts: Post[] = [];
  try {
    posts = await fetchPostAction(0, sort);
  } catch (error) {
    console.error("Failed to fetch initial posts:", error);
    posts = [];
  }

  return (
    <div className="mx-auto w-full max-w-3xl px-4 flex flex-col gap-8">
      {/*page header*/}
      <div>
        <h1 className="text-3xl font-bold tracking-tight">
          Public Policy Forum
        </h1>
        <p className="text-muted-foreground mt-2">
          Share your ideas, discuss local policies, and engage with the
          communities
        </p>
      </div>

      {user ? (
        <CreatePostForm />
      ) : (
        <div className="flex flex-col items-center justify-center p-8 border-2 border-dashed rounded-lg bg-muted/20 text-center">
          <h3 className="font-semibold text-lg mb-2">Join the Discussion</h3>
          <p className="text-muted-foreground text-sm mb-6 max-w-sm">
            You must be logged in to create a post, participate in comments, or
            vote in community policies
          </p>
          <div className="flex gap-4">
            <Button asChild>
              <Link href="/sign-in">Sign In</Link>
            </Button>
            <Button asChild variant="outline">
              <Link href="/sign-up">Create Account</Link>
            </Button>
          </div>
        </div>
      )}
      {/*the feed*/}

      <PostList initialPosts={posts} initialSort={sort} />
    </div>
  );
}
