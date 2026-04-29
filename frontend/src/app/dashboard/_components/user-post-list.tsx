"use client";

import { fetchUserPostsAction } from "@/app/actions/user";
import { Post, PostCard } from "@/app/forum/_components/post-card";
import { Button } from "@/components/ui/button";
import { useInfiniteQuery } from "@tanstack/react-query";

export function UserPostList() {
  const { data, status, fetchNextPage, hasNextPage, isFetchingNextPage } =
    useInfiniteQuery({
      queryKey: ["users", "posts"],
      queryFn: ({ pageParam }) => fetchUserPostsAction(pageParam),
      initialPageParam: 0 as string | number,
      getNextPageParam: (lastPage) => {
        if (!lastPage || lastPage.length < 10) return undefined;
        return lastPage[lastPage.length - 1].created_at;
      },
    });

  if (status === "pending")
    return (
      <div className="py-8 text-center text-muted-foreground">
        Loading posts...
      </div>
    );
  if (status === "error")
    return (
      <div className=" py-8 text-center text-destructive">
        Failed to load posts.
      </div>
    );

  const allPosts = data.pages.flat();

  if (allPosts.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-32 text-muted-foreground border-2 border-dashed rounded-md bg-muted/20">
        <p>You haven&apos;t start any discussion yet.</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      {allPosts.map((post: Post) => (
        <PostCard key={post.id} post={post} isDashboardView={true} />
      ))}

      {hasNextPage && (
        <Button
          variant="outline"
          className="mt-4"
          onClick={() => fetchNextPage()}
          disabled={isFetchingNextPage}
        >
          {isFetchingNextPage ? "Loading..." : "Load More"}
        </Button>
      )}
    </div>
  );
}
