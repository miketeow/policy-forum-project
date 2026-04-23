"use client";

import { fetchComments } from "@/lib/api-comments";
import { useInfiniteQuery } from "@tanstack/react-query";
import { CommentThread } from "./comment-thread";
import { Button } from "@/components/ui/button";

export interface CommentsDetail {
  id: string;
  parent_id: string | null;
  content: string;
  created_at: string;
  updated_at: string;
  author_id: string;
  author_name: string;
  reply_count: number;
}

export function CommentSection({ postId }: { postId: string }) {
  const { data, status, fetchNextPage, hasNextPage, isFetchingNextPage } =
    useInfiniteQuery({
      queryKey: ["comments", postId, "root"],
      queryFn: ({ pageParam }) =>
        fetchComments({ pageParam, postId, parentId: null }),
      initialPageParam: 0 as string | number,
      getNextPageParam: (lastPage) => {
        if (!lastPage || lastPage.length < 5) return undefined;
        return lastPage[lastPage.length - 1].created_at;
      },
    });

  if (status === "pending")
    return <div className="py-4">Loading comments...</div>;
  if (status === "error")
    return <div className="text-destructive py-4">Error loading comments.</div>;

  return (
    <div className="mt-8 flex flex-col gap-4 border-t pt-8">
      <h3 className="text-lg font-semibold">Comments</h3>

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
