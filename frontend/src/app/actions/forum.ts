"use server";
import { revalidatePath } from "next/cache";
import { CommentNode } from "../forum/_components/comment-thread";
import { Post } from "../forum/_components/post-card";
import {
  CommentSchema,
  PostSchema,
  UpdateCommentArgsSchema,
  VoteSchema,
} from "@/schemas/forum";
import { ApiError, handleActionError } from "@/lib/api-utils";
import { ActionState, fetchAPI, requireAuth } from "@/lib/api";

export interface PostDetail {
  id: string;
  title: string;
  content: string;
  category: string;
  created_at: string;
  updated_at: string;
  author_id: string;
  author_name: string;
  score: number;
  user_vote: number;
  summary: string;
}

export async function fetchPostAction(
  pageParam: number | string = 0,
  sort: string = "desc",
): Promise<Post[]> {
  const cursorQuery =
    sort === "popular"
      ? `&offset=${pageParam}`
      : pageParam && pageParam !== 0
        ? `&cursor=${encodeURIComponent(pageParam as string)}`
        : "";

  const data = await fetchAPI<{ posts: Post[] }>(
    `/api/posts?limit=20&sort=${sort}${cursorQuery}`,
  );
  return data.posts;
}

export async function fetchSinglePostAction(
  postId: string,
): Promise<PostDetail | null> {
  try {
    const data = await fetchAPI<{ post: PostDetail }>(`/api/posts/${postId}`);
    return data.post;
  } catch (error: unknown) {
    if (error instanceof ApiError) {
      if (error.status === 404) return null;
      console.error(`[fetchSinglePostAction] API Error:`, error.message);
    } else if (error instanceof Error) {
      console.error(`[fetchSinglePostAction] Network Error:`, error.message);
    } else {
      console.error(`[fetchSinglePostAction] Unknown Error:`, error);
    }
    return null;
  }
}

export async function createPostAction(
  formData: FormData,
): Promise<ActionState<Post>> {
  const authError = await requireAuth<Post>();
  if (authError) return authError;

  const rawData = {
    title: formData.get("title")?.toString().trim(),
    content: formData.get("content")?.toString().trim(),
  };

  const validatedFields = PostSchema.safeParse(rawData);

  if (!validatedFields.success) {
    return {
      success: false,
      error: validatedFields.error.issues[0].message,
    };
  }

  try {
    const responseData = await fetchAPI<{ post: Post }>(`/api/posts`, {
      method: "POST",
      body: JSON.stringify(validatedFields.data),
    });

    revalidatePath("/forum");
    return {
      success: true,
      message: "Post created successfully",
      data: responseData.post,
    };
  } catch (error: unknown) {
    return handleActionError<Post>(error, "createPostAction");
  }
}

export async function updatePostAction(
  postId: string,
  formData: FormData,
): Promise<ActionState> {
  const authError = await requireAuth();
  if (authError) return authError;

  const rawData = {
    title: formData.get("title")?.toString().trim(),
    content: formData.get("content")?.toString().trim(),
  };

  const validatedFields = PostSchema.safeParse(rawData);

  if (!validatedFields.success) {
    return {
      success: false,
      error: validatedFields.error.issues[0].message,
    };
  }

  try {
    await fetchAPI(`/api/posts/${postId}`, {
      method: "PUT",
      body: JSON.stringify(validatedFields.data),
    });

    revalidatePath(`/forum/${postId}`);
    revalidatePath(`/forum`);
    return { success: true, message: "Post updated successfully" };
  } catch (error) {
    return handleActionError(error, "updatePostAction");
  }
}

export async function deletePostAction(postId: string): Promise<ActionState> {
  const authError = await requireAuth();
  if (authError) return authError;

  try {
    await fetchAPI(`/api/posts/${postId}`, {
      method: "DELETE",
    });

    revalidatePath(`/forum`);
    return { success: true, message: "Post deleted successfully" };
  } catch (error) {
    return handleActionError(error, "deletePostAction");
  }
}

export async function createCommentAction(
  formData: FormData,
): Promise<ActionState> {
  const authError = await requireAuth();
  if (authError) return authError;

  const rawData = {
    postId: formData.get("postId")?.toString().trim(),
    content: formData.get("content")?.toString().trim(),
    parentId: formData.get("parentId")?.toString().trim() || undefined,
  };

  const validatedFields = CommentSchema.safeParse(rawData);

  if (!validatedFields.success) {
    return {
      success: false,
      error: validatedFields.error.issues[0].message,
    };
  }

  const { postId, content, parentId } = validatedFields.data;

  try {
    await fetchAPI(`/api/posts/${postId}/comments`, {
      method: "POST",

      body: JSON.stringify({
        content,
        parentId,
      }),
    });

    revalidatePath(`/forum/${postId}`);

    return { success: true, message: "Post comment successfully" };
  } catch (error) {
    return handleActionError(error, "createCommentAction");
  }
}

export async function updateCommentAction(
  commentId: string,
  content: string,
  postId: string,
): Promise<ActionState> {
  const authError = await requireAuth();
  if (authError) return authError;

  const validatedFields = UpdateCommentArgsSchema.safeParse({
    commentId: commentId,
    postId: postId,
    content: content.trim(),
  });

  if (!validatedFields.success) {
    return {
      success: false,
      error: validatedFields.error.issues[0].message,
    };
  }

  try {
    await fetchAPI(`/api/comments/${validatedFields.data.commentId}`, {
      method: "PUT",
      body: JSON.stringify({ content: validatedFields.data.content }),
    });

    revalidatePath(`/forum/${validatedFields.data.postId}`);
    return { success: true, message: "Comment updated" };
  } catch (error) {
    return handleActionError(error, "updateCommentAction");
  }
}

export async function deleteCommentAction(
  commentId: string,
  postId: string,
): Promise<ActionState> {
  const authError = await requireAuth();
  if (authError) return authError;

  try {
    await fetchAPI(`/api/comments/${commentId}`, { method: "DELETE" });
    revalidatePath(`/forum/${postId}`);
    return { success: true, message: "Comment deleted successfully" };
  } catch (error) {
    return handleActionError(error, "deleteCommentAction");
  }
}

export async function votePostAction(
  postId: string,
  vote: 1 | -1,
): Promise<ActionState> {
  const authError = await requireAuth();
  if (authError) return authError;

  const validatedFields = VoteSchema.safeParse({ vote: vote });
  if (!validatedFields.success) {
    return {
      success: false,
      error: validatedFields.error.issues[0].message,
    };
  }

  try {
    await fetchAPI(`/api/posts/${postId}/vote`, {
      method: "POST",
      body: JSON.stringify({ vote: validatedFields.data.vote }),
    });
    return { success: true, message: "Post voted successfully" };
  } catch (error) {
    return handleActionError(error, "votePostAction");
  }
}

export async function voteCommentAction(
  commentId: string,
  vote: 1 | -1,
): Promise<ActionState> {
  const authError = await requireAuth();
  if (authError) return authError;

  const validatedFields = VoteSchema.safeParse({ vote: vote });

  if (!validatedFields.success) {
    return {
      success: false,
      error: validatedFields.error.issues[0].message,
    };
  }
  try {
    await fetchAPI(`/api/comments/${commentId}/vote`, {
      method: "POST",
      body: JSON.stringify({ vote: validatedFields.data.vote }),
    });

    return { success: true, message: "Comment voted successfully" };
  } catch (error) {
    return handleActionError(error, "voteCommentAction");
  }
}

export async function fetchCommentsAction(
  postId: string,
  parentId: string | null,
  pageParam: number | string = 0,
  sort: string = "desc",
): Promise<CommentNode[]> {
  // base url
  let url = `/api/posts/${postId}/comments?limit=5`;

  // dynamic pagination
  if (sort === "popular") url += `&offset=${pageParam}`;
  else if (pageParam && pageParam !== 0)
    url += `&cursor=${encodeURIComponent(pageParam as string)}`;

  if (parentId) url += `&parentId=${parentId}`;
  if (sort) url += `&sort=${sort}`;

  try {
    const data = await fetchAPI<{ comments: CommentNode[] }>(url);
    return data.comments;
  } catch (error: unknown) {
    if (error instanceof ApiError) {
      if (error.status === 404) return [];
      console.error(`[fetchSinglePostAction] API Error:`, error.message);
    } else if (error instanceof Error) {
      console.error(`[fetchSinglePostAction] Network Error:`, error.message);
    } else {
      console.error(`[fetchSinglePostAction] Unknown Error:`, error);
    }
    return [];
  }
}

export async function checkPostCategoryAction(
  postId: string,
): Promise<string | null> {
  try {
    const data = await fetchAPI<{ post: PostDetail }>(`/api/posts/${postId}`);
    return data.post.category;
  } catch {
    return null;
  }
}

export async function triggerAiSummaryAction(
  postId: string,
): Promise<ActionState> {
  const authError = await requireAuth();
  if (authError) return authError;
  try {
    await fetchAPI(`/api/posts/${postId}/summary`, {
      method: "POST",
    });
    return {
      success: true,
      message: "Summary generation queue successfully",
    };
  } catch (error) {
    return handleActionError(error, "triggerAiSummaryAction");
  }
}
