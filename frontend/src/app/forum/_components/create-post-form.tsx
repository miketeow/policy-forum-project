"use client";

import { createPostAction } from "@/app/actions/forum";
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field";
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
  InputGroupTextarea,
} from "@/components/ui/input-group";
import { useActionState, useEffect } from "react";
import { toast } from "sonner";

const initialState = {
  success: false,
  message: "",
  error: "",
};
export function CreatePostForm() {
  const [state, formAction, isPending] = useActionState(
    createPostAction,
    initialState,
  );

  useEffect(() => {
    if (state.error) {
      toast.error(state.error);
    } else if (state.success) {
      toast.success(state.message);
    }
  }, [state]);

  return (
    <div className="rounded-lg p-6 shadow-sm bg-card text-card-foreground">
      <form action={formAction}>
        <FieldGroup className="w-full">
          <Field>
            <FieldLabel className="text-lg font-semibold mb-2">
              Title
            </FieldLabel>
            <InputGroup>
              <InputGroupInput
                name="title"
                disabled={isPending}
                placeholder="What is it about ?"
                required
              />
            </InputGroup>
          </Field>
          <Field>
            <FieldLabel className="text-lg font-semibold mb-2">
              Start a new discussion
            </FieldLabel>
            <InputGroup>
              <InputGroupTextarea
                id="block-end-textarea"
                name="content"
                placeholder="What are your thoughts on current policies?..."
                className="min-h-25 resize-y"
                disabled={isPending}
                required
              />
              <InputGroupAddon align="block-end">
                <InputGroupButton
                  className="ml-auto"
                  variant="default"
                  size="sm"
                  type="submit"
                  disabled={isPending}
                >
                  {isPending ? "Posting..." : "Post Discussion"}
                </InputGroupButton>
              </InputGroupAddon>
            </InputGroup>
          </Field>
        </FieldGroup>
      </form>
    </div>
  );
}
