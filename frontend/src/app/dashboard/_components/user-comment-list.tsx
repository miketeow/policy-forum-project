"use client";

import { fetchUserCommentsAction } from "@/app/actions/user";
import {
  CommentNode,
  CommentThread,
} from "@/app/forum/_components/comment-thread";

import { Button } from "@/components/ui/button";
import { useInfiniteQuery } from "@tanstack/react-query";

export function UserCommentList({ currentUserId }: { currentUserId: string }) {
  const { data, status, fetchNextPage, hasNextPage, isFetchingNextPage } =
    useInfiniteQuery({
      queryKey: ["users", "comments"],
      queryFn: ({ pageParam }) => fetchUserCommentsAction(pageParam),
      initialPageParam: 0 as string | number,
      getNextPageParam: (lastPage) => {
        if (!lastPage || lastPage.length < 10) return undefined;
        return lastPage[lastPage.length - 1].created_at;
      },
    });

  if (status === "pending")
    return (
      <div className="py-8 text-center text-muted-foreground">
        Loading comments...
      </div>
    );
  if (status === "error")
    return (
      <div className=" py-8 text-center text-destructive">
        Failed to load comments.
      </div>
    );

  const allComments = data.pages.flat();

  if (allComments.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-32 text-muted-foreground border-2 border-dashed rounded-md bg-muted/20">
        <p>You haven&apos;t start any discussion yet.</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      {allComments.map((comment: CommentNode) => (
        <CommentThread
          key={comment.id}
          comment={comment}
          postId={comment.post_id}
          currentUserId={currentUserId}
          showPostLink={true}
          isDashboardView={true}
        />
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
