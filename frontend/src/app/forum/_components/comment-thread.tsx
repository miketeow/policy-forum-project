"use client";

import {
  createCommentAction,
  deleteCommentAction,
  updateCommentAction,
} from "@/app/actions/forum";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Textarea } from "@/components/ui/textarea";
import { fetchComments } from "@/lib/api-comments";
import { formatDate } from "@/lib/utils";
import { useInfiniteQuery, useQueryClient } from "@tanstack/react-query";
import { MoreVertical } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

export interface CommentNode {
  id: string;
  parent_id: string | null;
  content: string;
  created_at: string;
  author_name: string;
  author_id: string;
  reply_count: number;
}

export function CommentThread({
  comment,
  postId,
  currentUserId,
}: {
  comment: CommentNode;
  postId: string;
  currentUserId: string;
}) {
  const [isReplying, setIsReplying] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showReply, setShowReply] = useState(false);

  const [isEditing, setIsEditing] = useState(false);
  const [editContent, setEditContent] = useState(comment.content);
  const [isPending, setIsPending] = useState(false);

  const queryClient = useQueryClient();
  const isOwner = currentUserId === comment.author_id;

  const { data, status, fetchNextPage, hasNextPage, isFetchingNextPage } =
    useInfiniteQuery({
      queryKey: ["comments", postId, comment.id],
      queryFn: ({ pageParam }) =>
        fetchComments({ pageParam, postId, parentId: comment.id }),
      initialPageParam: 0 as string | number,
      getNextPageParam: (lastPage) => {
        if (!lastPage || lastPage.length < 5) return undefined;
        const cursor = lastPage[lastPage.length - 1].created_at;
        console.log("[REACT TRACER] Passing Cursor to Next Page:", cursor); // ADD THIS
        return cursor;
      },
      // don't fetch from Go backend until user click "show replies"
      enabled: showReply,
    });

  async function handleReplySubmit(formData: FormData) {
    setIsSubmitting(true);
    const result = await createCommentAction(formData);

    if (!result.success) {
      toast.error(result.error);
    } else {
      toast.success(result.message);
      setIsReplying(false);
      queryClient.invalidateQueries({
        queryKey: ["comments", postId, comment.id],
      });
      queryClient.invalidateQueries({
        queryKey: ["comments", postId, comment.parent_id || "root"],
      });
      setShowReply(true);
    }
    setIsSubmitting(false);
  }

  async function handleEditSubmit() {
    setIsPending(true);
    const res = await updateCommentAction(comment.id, editContent, postId);
    if (res.success) {
      toast.success(res.message);
      setIsEditing(false);
      queryClient.invalidateQueries({
        queryKey: ["comments", postId, comment.parent_id || "root"],
      });
    } else {
      toast.error(res.error);
    }
  }

  async function handleDelete() {
    if (!confirm("Are you sure want to delete this comment?")) return;
    const res = await deleteCommentAction(comment.id, postId);
    if (res.success) {
      toast.success(res.message);
      queryClient.invalidateQueries({
        queryKey: ["comments", postId, comment.parent_id || "root"],
      });
    } else {
      toast.error(res.error);
    }
  }

  return (
    <div className="flex flex-col mt-4">
      {/*comment box*/}
      <div className="flex flex-col gap-2 p-4 rounded-lg bg-card border shadow-sm">
        <div className="flex justify-between items-center">
          <div>
            <span className="font-semibold text-sm">{comment.author_name}</span>
            <span className="text-xs text-muted-foreground">
              {formatDate(comment.created_at)}
            </span>
          </div>

          {isOwner && (
            <DropdownMenu modal={false}>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon" className="size-8">
                  <MoreVertical size={16} />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => setIsEditing(true)}>
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={handleDelete}
                  className="text-destructive"
                >
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>

        {/*content area: swap between text and form*/}
        {isEditing ? (
          <div className="flex flex-col gap-2 mt-2">
            <Textarea
              value={editContent}
              onChange={(e) => setEditContent(e.target.value)}
              className="text-sm min-h-20"
            />
            <div className="flex justify-end gap-2">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setIsEditing(false)}
              >
                Cancel
              </Button>
              <Button size="sm" onClick={handleEditSubmit} disabled={isPending}>
                {isPending ? "Saving..." : "Save"}
              </Button>
            </div>
          </div>
        ) : (
          <p className="text-sm whitespace-pre-wrap mt-1">{comment.content}</p>
        )}

        <div className="mt-2 flex gap-4">
          <Button
            variant="ghost"
            size="sm"
            className="h-8 px-2 text-muted-foreground hover:text-foreground text-xs"
            onClick={() => setIsReplying(!isReplying)}
          >
            {isReplying ? "Cancel" : "Reply"}
          </Button>
          {comment.reply_count > 0 && (
            <Button
              variant="ghost"
              size="sm"
              className="h-8 px-2 text-xs text-blue-500 "
              onClick={() => setShowReply(!showReply)}
            >
              {showReply
                ? "Hide Replies"
                : `Show Replies (${comment.reply_count})`}
            </Button>
          )}
        </div>
      </div>

      {/*inline reply form*/}
      {isReplying && (
        <form
          action={handleReplySubmit}
          className="mt-3 ml-6 flex flex-col gap-2"
        >
          <input type="hidden" name="postId" value={postId} />
          <input type="hidden" name="parentId" value={comment.id} />

          <Textarea
            name="content"
            placeholder={`Replying to ${comment.author_name}...`}
            className="min-h-20 text-sm"
          />
          <div className="flex justify-end">
            <Button type="submit" size="sm">
              {isSubmitting ? "Posting..." : "Post Reply"}
            </Button>
          </div>
        </form>
      )}

      {showReply && (
        <div className="ml-6 pl-4 border-l-2 mt-2 flex flex-col gap-2">
          {status === "pending" && (
            <p className="text-xs text-muted-foreground py-2">
              Loading replies
            </p>
          )}
          {status === "success" &&
            data.pages.map((page, i) => (
              <div key={i} className="flex flex-col gap-2">
                {page.map((reply: CommentNode) => (
                  <CommentThread
                    comment={reply}
                    postId={postId}
                    key={reply.id}
                    currentUserId={currentUserId}
                  />
                ))}
              </div>
            ))}

          {hasNextPage && (
            <Button
              variant="link"
              size="sm"
              onClick={() => fetchNextPage()}
              disabled={isFetchingNextPage}
              className="self-start text-xs text-muted-foreground"
            >
              {isFetchingNextPage ? "Loading..." : "Load more replies"}
            </Button>
          )}
        </div>
      )}
    </div>
  );
}
