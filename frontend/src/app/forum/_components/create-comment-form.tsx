"use client";

import { createCommentAction } from "@/app/actions/forum";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { useRef, useState } from "react";
import { toast } from "sonner";

export function CreateCommentForm({ postId }: { postId: string }) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const formRef = useRef<HTMLFormElement>(null);

  async function handleSubmit(formData: FormData) {
    setIsSubmitting(true);

    const result = await createCommentAction(formData);

    if (!result.success) {
      toast.error(result.error);
    } else {
      toast.success(result.message);
      formRef.current?.reset(); // clear the textarea after success
    }

    setIsSubmitting(false);
  }

  return (
    <form
      ref={formRef}
      action={handleSubmit}
      className="flex flex-col gap-4 mb-8 bg-muted/30 p-4 rounded-lg border"
    >
      <input type="hidden" name="postId" value={postId} />
      <Textarea
        name="content"
        placeholder="Share your thoughts on this policy..."
        className="min-h-25"
        required
      />
      <div className="flex justify-end">
        <Button type="submit">
          {isSubmitting ? "Posting..." : "Post Comment"}
        </Button>
      </div>
    </form>
  );
}
