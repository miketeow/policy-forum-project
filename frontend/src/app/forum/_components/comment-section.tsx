"use client";

import { keepPreviousData, useInfiniteQuery } from "@tanstack/react-query";
import { CommentThread } from "./comment-thread";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useRouter, useSearchParams } from "next/navigation";
import { fetchCommentsAction } from "@/app/actions/forum";

export interface CommentsDetail {
  id: string;
  parent_id: string | null;
  content: string;
  created_at: string;
  updated_at: string;
  author_id: string;
  author_name: string;
  reply_count: number;
  score: number;
  user_vote: number;
}

export function CommentSection({
  postId,
  initialSort,
  currentUserId,
}: {
  postId: string;
  initialSort: "desc" | "asc" | "popular";
  currentUserId: string;
}) {
  const router = useRouter();
  const searchParams = useSearchParams();

  const sortOrder =
    (searchParams.get("sort") as "desc" | "asc" | "popular") || initialSort;

  const handleSortChange = (newSort: "desc" | "asc" | "popular") => {
    const params = new URLSearchParams(searchParams.toString());
    params.set("sort", newSort);

    router.push(`?${params.toString()}`, { scroll: false });
  };

  const { data, status, fetchNextPage, hasNextPage, isFetchingNextPage } =
    useInfiniteQuery({
      queryKey: ["comments", postId, "root", sortOrder, currentUserId],
      queryFn: ({ pageParam }) =>
        fetchCommentsAction(postId, null, pageParam, sortOrder),
      initialPageParam: 0 as string | number,
      placeholderData: keepPreviousData,
      getNextPageParam: (lastPage, allPages) => {
        if (!lastPage || lastPage.length < 5) return undefined;
        if (sortOrder === "popular") {
          return allPages.length * 5;
        }
        return lastPage[lastPage.length - 1].created_at;
      },
    });

  if (status === "pending")
    return <div className="py-4">Loading comments...</div>;
  if (status === "error")
    return <div className="text-destructive py-4">Error loading comments.</div>;

  return (
    <div className="mt-8 flex flex-col gap-4 border-t pt-8">
      {/*header*/}
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">Comments</h3>
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

      {data.pages[0].length === 0 ? (
        <p className="text-sm text-muted-foreground">
          No comments yet. Start the conversation!
        </p>
      ) : (
        <div className="flex flex-col gap-4">
          {data.pages.map((page, i) => (
            <div key={i} className="flex flex-col gap-4">
              {page.map((comment: CommentsDetail) => (
                <CommentThread
                  key={comment.id}
                  comment={comment}
                  postId={postId}
                  currentUserId={currentUserId}
                />
              ))}
            </div>
          ))}
        </div>
      )}

      {hasNextPage && (
        <Button
          className="w-full mt-2"
          variant="ghost"
          onClick={() => fetchNextPage()}
          disabled={isFetchingNextPage}
        >
          {isFetchingNextPage ? "Loading..." : "Load More Comments"}
        </Button>
      )}
    </div>
  );
}
