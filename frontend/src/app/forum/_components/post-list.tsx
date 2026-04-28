"use client";

import { keepPreviousData, useInfiniteQuery } from "@tanstack/react-query";
import { Post, PostCard } from "./post-card";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useRouter, useSearchParams } from "next/navigation";
import { fetchPostAction } from "@/app/actions/forum";

interface PostListProps {
  initialPosts: Post[];
  initialSort: "asc" | "desc" | "popular";
}

export function PostList({ initialPosts, initialSort }: PostListProps) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const sortOrder =
    (searchParams.get("sort") as "desc" | "asc" | "popular") || initialSort;

  const handleSortChange = (newSort: "desc" | "asc" | "popular") => {
    const params = new URLSearchParams(searchParams.toString());
    params.set("sort", newSort);

    router.push(`?${params.toString()}`, { scroll: false });
  };
  const {
    data,
    status,
    error,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteQuery({
    queryKey: ["posts", sortOrder],
    queryFn: ({ pageParam }) => fetchPostAction(pageParam, sortOrder),

    initialPageParam: 0,
    placeholderData: keepPreviousData,
    initialData: {
      pages: [initialPosts], // first page
      pageParams: [0], // cursor for first page is 0
    },
    // how to figure out the next cursor, TanStack provides the "lastPage" which is the array of 20 posts we just fetched
    getNextPageParam: (lastPage, allPages) => {
      // If Go backend return less than 20 items or an empty array
      // it means we hit the end of the database, hence return undefined to notify
      // proceed to set hasNextPage as false
      if (!lastPage || lastPage.length < 20) {
        return undefined;
      }
      if (sortOrder === "popular") {
        return allPages.length * 20;
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
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold pb-2">Recent Discussions</h2>
        <DropdownMenu modal={false}>
          <DropdownMenuTrigger asChild>
            <Button
              variant="outline"
              size="sm"
              className="text-muted-foreground"
            >
              Sort by:{" "}
              {sortOrder === "desc"
                ? "Newest"
                : sortOrder === "asc"
                  ? "Oldest"
                  : "Popular"}
            </Button>
          </DropdownMenuTrigger>

          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => handleSortChange("desc")}>
              Newest
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => handleSortChange("asc")}>
              Oldest
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => handleSortChange("popular")}>
              Popular
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

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
    </div>
  );
}
