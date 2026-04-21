"use client";

import { useInfiniteQuery } from "@tanstack/react-query";
import { Post, PostCard } from "./post-card";
import { Button } from "@/components/ui/button";

interface PostListProps {
  initialPosts: Post[];
}

async function fetchPost({ pageParam = 0 }: { pageParam: number | string }) {
  // if cursor is 0, don't attach it to the url, get page 1 instead
  // otherwise append it
  const cursorQuery = pageParam ? `&cursor=${pageParam}` : "";

  const res = await fetch(
    `http://localhost:8080/api/posts?limit=20${cursorQuery}`,
  );

  if (!res.ok) {
    throw new Error("Failed to fetch posts");
  }

  return res.json();
}

export function PostList({ initialPosts }: PostListProps) {
  const {
    data,
    status,
    error,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteQuery({
    queryKey: ["posts"],
    queryFn: fetchPost,

    initialPageParam: 0,
    initialData: {
      pages: [initialPosts], // first page
      pageParams: [0], // cursor for first page is 0
    },

    // how to figure out the next cursor, TanStack provides the "lastPage" which is the array of 20 posts we just fetched
    getNextPageParam: (lastPage) => {
      // If Go backend return less than 20 items or an empty array
      // it means we hit the end of the database, hence return undefined to notify
      // proceed to set hasNextPage as false
      if (!lastPage || lastPage.length < 20) {
        return undefined;
      }
      const lastPost = lastPage[lastPage.length - 1];
      return lastPost.created_at;
    },
  });

  // rendering ui

  if (status === "error") {
    return <p className="p-4 text-destructive">Error: {error.message}</p>;
  }

  if (data.pages[0].length === 0) {
    return (
      <div className="pb-4 rounded-md opacity-50 flex items-center justify-center h-32 bg-muted/50">
        <p className="text-muted-foreground text-sm">
          No discussion yet. Be the first to post!
        </p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      {data.pages.map((page, pageIndex) => (
        <div key={pageIndex} className="flex flex-col gap-4">
          {page.map((post: Post) => (
            <PostCard key={post.id} post={post} />
          ))}
        </div>
      ))}

      {/*load more action*/}
      <div className="mt-8 flex justify-center pb-8">
        <Button
          onClick={() => fetchNextPage()}
          disabled={!hasNextPage || isFetchingNextPage}
          className="px-6 py-2 rounded-md"
          variant="default"
        >
          {isFetchingNextPage
            ? "Loading more..."
            : hasNextPage
              ? "Load More"
              : "Nothing more to load"}
        </Button>
      </div>
    </div>
  );
}
