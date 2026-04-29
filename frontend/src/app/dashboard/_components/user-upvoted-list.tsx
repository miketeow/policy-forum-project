"use client";

import {
  fetchUserUpvotedCommentsAction,
  fetchUserUpvotedPostsAction,
} from "@/app/actions/user";
import {
  CommentNode,
  CommentThread,
} from "@/app/forum/_components/comment-thread";
import { Post, PostCard } from "@/app/forum/_components/post-card";
import { Button } from "@/components/ui/button";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { useInfiniteQuery } from "@tanstack/react-query";
import { useState } from "react";

export function UserUpvotedList({ currentUserId }: { currentUserId: string }) {
  const [activeTab, setActiveTab] = useState<"posts" | "comments">("posts");
  // fetch upvoted posts
  const {
    data: postsData,
    status: postsStatus,
    fetchNextPage: fetchPosts,
    hasNextPage: hasMorePosts,
    isFetchingNextPage: fetchingPosts,
  } = useInfiniteQuery({
    queryKey: ["users", "upvoted", "posts"],
    queryFn: ({ pageParam }) => fetchUserUpvotedPostsAction(pageParam),
    initialPageParam: 0 as string | number,
    getNextPageParam: (lastPage) =>
      lastPage?.length === 10 ? lastPage[9].created_at : undefined,
    enabled: activeTab === "posts",
  });
  // fetch upvoted comments
  const {
    data: commentsData,
    status: commentsStatus,
    fetchNextPage: fetchComments,
    hasNextPage: hasMoreComments,
    isFetchingNextPage: fetchingComments,
  } = useInfiniteQuery({
    queryKey: ["users", "upvoted", "comments"],
    queryFn: ({ pageParam }) => fetchUserUpvotedCommentsAction(pageParam),
    initialPageParam: 0 as string | number,
    getNextPageParam: (lastPage) =>
      lastPage?.length === 10 ? lastPage[9].created_at : undefined,
    enabled: activeTab === "comments",
  });

  return (
    <div className="flex flex-col gap-4">
      {/*toggle switch*/}
      <ToggleGroup
        type="single"
        value={activeTab}
        onValueChange={(value) => {
          if (value) setActiveTab(value as "posts" | "comments");
        }}
        className="justify-start mb-2 bg-muted/30 p-1 rounded-lg w-fit"
      >
        <ToggleGroupItem
          value="posts"
          className="px-4 data-[state=on]:bg-background data-[state=on]:shadow-sm"
        >
          Posts
        </ToggleGroupItem>
        <ToggleGroupItem
          value="comments"
          className="px-4 data-[state=on]:bg-background data-[state=on]:shadow-sm"
        >
          Comments
        </ToggleGroupItem>
      </ToggleGroup>

      {/*render posts*/}
      {activeTab === "posts" && (
        <div className="flex flex-col gap-4">
          {postsStatus === "pending" && (
            <p className="text-muted-foreground text-sm">Loading posts...</p>
          )}
          {postsData?.pages.flat().map((post: Post) => (
            <PostCard key={post.id} post={post} isDashboardView={true} />
          ))}
          {hasMorePosts && (
            <Button
              variant="outline"
              onClick={() => fetchPosts()}
              disabled={fetchingPosts}
            >
              Load More
            </Button>
          )}
        </div>
      )}

      {/*render comments*/}
      {activeTab === "comments" && (
        <div className="flex flex-col gap-4">
          {commentsStatus === "pending" && (
            <p className="text-muted-foreground text-sm">Loading comments...</p>
          )}
          {commentsData?.pages.flat().map((comment: CommentNode) => (
            <CommentThread
              key={comment.id}
              comment={comment}
              postId={comment.post_id}
              currentUserId={currentUserId}
              showPostLink={true}
              isDashboardView={true}
            />
          ))}
          {hasMoreComments && (
            <Button
              variant="outline"
              onClick={() => fetchComments()}
              disabled={fetchingComments}
            >
              Load More
            </Button>
          )}
        </div>
      )}
    </div>
  );
}
