"use client";

import { createCommentAction } from "@/app/actions/forum";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { formatDate } from "@/lib/utils";
import { useState } from "react";
import { toast } from "sonner";

export interface CommentNode {
  id: string;
  parent_id: string | null;
  content: string;
  created_at: string;
  author_name: string;
  author_id: string;
  children: CommentNode[];
}

export function CommentThread({
  comment,
  postId,
}: {
  comment: CommentNode;
  postId: string;
}) {
  const [isReplying, setIsReplying] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleReplySubmit(formData: FormData) {
    setIsSubmitting(true);
    const result = await createCommentAction(formData);

    if (!result.success) {
      toast.error(result.error);
    } else {
      toast.success(result.message);
      setIsReplying(false);
    }
    setIsSubmitting(false);
  }

  return (
    <div className="flex flex-col mt-4">
      {/*comment box*/}
      <div className="flex flex-col gap-2 p-4 rounded-lg bg-card border shadow-sm">
        <div className="flex justify-between items-center">
          <span className="font-semibold text-sm">{comment.author_name}</span>
          <span className="text-xs text-muted-foreground">
            {formatDate(comment.created_at)}
          </span>
        </div>

        <p className="text-sm whitespace-pre-wrap mt-1">{comment.content}</p>
        <div className="mt-2">
          <Button
            variant="ghost"
            size="sm"
            className="h-8 px-2 text-muted-foreground hover:text-foreground text-xs"
            onClick={() => setIsReplying(true)}
          >
            {isReplying ? "Cancel" : "Reply"}
          </Button>
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

      {comment.children.length > 0 && (
        <div className="ml-6 pl-4 border-l-2 mt-2 flex flex-col gap-2">
          {comment.children.map((child) => (
            <CommentThread key={child.id} comment={child} postId={postId} />
          ))}
        </div>
      )}
    </div>
  );
}
