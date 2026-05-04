import { getSession } from "@/lib/session";
import { redirect } from "next/navigation";
import { CreatePostForm } from "./_components/create-post-form";
import { PostList } from "./_components/post-list";
import { cookies } from "next/headers";
import { fetchPostAction } from "../actions/forum";
import { Post } from "./_components/post-card";

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
  const user: UserProfile = await getSession();

  const cookieStore = await cookies();
  const hasCookie = cookieStore.has("session");

  if (!user && hasCookie) {
    redirect("/api/auth/logout");
  }

  if (!user) {
    redirect("/sign-in");
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

      {/*create post*/}
      <CreatePostForm />

      {/*the feed*/}

      <PostList initialPosts={posts} initialSort={sort} />
    </div>
  );
}
