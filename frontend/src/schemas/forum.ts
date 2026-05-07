import { z } from "zod";

const Categories = [
  "INFRASTRUCTURE",
  "ECONOMY",
  "HEALTHCARE",
  "EDUCATION",
  "ENVIRONMENT",
  "SAFETY",
  "OTHER",
] as const;

export const PostSchema = z.object({
  title: z
    .string({ error: () => "Title is required" })
    .min(5, { error: "Title must be at least 5 characters" })
    .max(300, { error: "Title cannot exceed 300 characters" }),

  content: z
    .string({ error: () => "Content is required" })
    .min(10, { error: "Content must be at least 10 characters" })
    .max(40000, { error: "Content is too long" }),

  category: z.enum(Categories).optional(),
});

const CommentContentRule = z
  .string({ error: () => "Comment is required" })
  .min(1, { error: "Comment cannot be empty" })
  .max(10000, { error: "Comment is too long" });

export const CommentSchema = z.object({
  content: CommentContentRule,
  postId: z.string({ error: () => "Post ID is required" }),
  // Top-level UUID validator in v4
  parentId: z.uuid({ error: "Invalid Parent ID format" }).optional(),
});

export const UpdateCommentArgsSchema = z.object({
  commentId: z.uuid({ error: "Invalid Comment ID format" }),
  postId: z.uuid({ error: "Invalid Post ID format" }),
  content: CommentContentRule,
});
export const VoteSchema = z.object({
  // Zod v4 array syntax for literals - much cleaner than z.union!
  vote: z.literal([1, -1], { error: "Vote must be strictly 1 or -1" }),
});
