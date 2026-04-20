import { getSession } from "@/lib/session";
import { redirect } from "next/navigation";
import { CreatePostForm } from "./_components/create-post-form";
import { Post, PostCard } from "./_components/post-card";

interface UserProfile {
  id: string;
  name: string;
  email: string;
  kyc_status: string;
  created_at: string;
  updated_at: string;
}

async function getPosts(): Promise<Post[]> {
  try {
    // hit the go backend directly, set "no-store" to avoid aggresive cache
    const res = await fetch("http://localhost:8080/api/posts", {
      cache: "no-store",
    });

    if (!res.ok) {
      console.error("Failed to fetch posts from backend");
      return [];
    }

    return res.json();
  } catch (error) {
    console.error("Network error fetching posts: ", error);
    return [];
  }
}
export default async function Forum() {
  const user: UserProfile = await getSession();

  if (!user) {
    redirect("/sign-in");
  }

  const posts = await getPosts();

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
      <div className="flex flex-col gap-4">
        <h2 className="text-xl font-semibold pb-2">Recent Discussions</h2>
        {posts.length === 0 ? (
          <div className="pb-4 rounded-md opacity-50 flex items-center justify-center h-32 bg-muted/50">
            <p className="text-muted-foreground text-sm">
              No discussion yet. Be the first to post!
            </p>
          </div>
        ) : (
          <div className="flex flex-col gap-4">
            {posts.map((post) => (
              <PostCard key={post.id} post={post} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
