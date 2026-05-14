"use server";
import { revalidatePath } from "next/cache";
import { cookies } from "next/headers";
import { CommentNode } from "../forum/_components/comment-thread";
import { Post } from "../forum/_components/post-card";
import { parseBackendError } from "@/lib/utils";
import {
  CommentSchema,
  PostSchema,
  UpdateCommentArgsSchema,
  VoteSchema,
} from "@/schemas/forum";
import { getApiUrl } from "@/lib/api";

export interface ActionState<T = undefined> {
  success: boolean;
  message?: string;
  error?: string;
  data?: T;
}

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
  const cookieStore = await cookies();
  const token = cookieStore.get("session")?.value;

  const headers = new Headers();
  // attach the token
  if (token) {
    headers.append("Authorization", `Bearer ${token}`);
  }

  let cursorQuery = "";
  if (sort == "popular") {
    cursorQuery = `&offset=${pageParam}`;
  } else if (pageParam && pageParam !== 0) {
    cursorQuery = `&cursor=${encodeURIComponent(pageParam as string)}`;
  }

  const url = `http://localhost:8080/api/posts?limit=20&sort=${sort}${cursorQuery}`;

  const response = await fetch(url, { headers, cache: "no-store" });
  if (!response.ok) {
    const errorMessage = await parseBackendError(
      response,
      "Failed to fetch posts",
    );
    throw new Error(errorMessage);
  }
  const data = await response.json();
  return data.posts;
}

export async function fetchSinglePostAction(
  postId: string,
): Promise<PostDetail | null> {
  const cookieStore = await cookies();
  const token = cookieStore.get("session")?.value;

  const headers = new Headers();
  // attach the token
  if (token) {
    headers.append("Authorization", `Bearer ${token}`);
  }
  try {
    const response = await fetch(`http://localhost:8080/api/posts/${postId}`, {
      headers,
      cache: "no-store",
    });

    if (!response.ok) {
      if (response.status === 404) {
        return null;
      }
      throw new Error(`Failed to fetch post: ${response.status}`);
    }
    const data = await response.json();

    return data.post;
  } catch (error) {
    console.error("fetchSinglePostAction Network Error:", error);
    return null;
  }
}

export async function createPostAction(
  formData: FormData,
): Promise<ActionState<Post>> {
  const cookieStore = await cookies();
  const token = cookieStore.get("session")?.value;

  if (!token) {
    return {
      success: false,
      message: "Authentication required",
      error: "Your session has expired. Please log in again",
    };
  }

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

  const { title, content } = validatedFields.data;

  try {
    const response = await fetch("http://localhost:8080/api/posts", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ title, content }),
    });

    if (!response.ok) {
      // parse json response from Go server
      const errorMessage = await parseBackendError(
        response,
        "Failed to create post",
      );
      return {
        success: false,
        message: "failed to create post",
        error: errorMessage,
      };
    }

    const responseData = await response.json();
    revalidatePath("/forum");
    return {
      success: true,
      message: "Post created successfully",
      data: responseData.post,
    };
  } catch (error) {
    // log error to server logs
    console.error("createPostAction Network Error:", error);
    return {
      success: false,
      message: "Connection failure",
      error:
        "Could not reach the server. Please check your internet connection again",
    };
  }
}

export async function updatePostAction(
  postId: string,
  formData: FormData,
): Promise<ActionState> {
  const cookieStore = await cookies();
  const token = cookieStore.get("session")?.value;

  if (!token) {
    return {
      success: false,
      message: "Authentication required",
      error: "Your session has expired. Please log in again",
    };
  }

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

  const { title, content } = validatedFields.data;

  try {
    const response = await fetch(`http://localhost:8080/api/posts/${postId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ title, content }),
    });

    if (!response.ok) {
      const errorMessage = await parseBackendError(
        response,
        "Failed to update post",
      );
      return {
        success: false,
        message: "Backend error",
        error: errorMessage,
      };
    }

    revalidatePath(`/forum/${postId}`);
    revalidatePath(`/forum`);
    return { success: true, message: "Post updated successfully" };
  } catch (error) {
    console.error("updatePostAction Network Error:", error);
    return {
      success: false,
      message: "Connection failure",
      error:
        "Could not reach the server. Please check your internet connection again",
    };
  }
}

export async function deletePostAction(postId: string): Promise<ActionState> {
  const cookieStore = await cookies();
  const token = cookieStore.get("session")?.value;

  if (!token) {
    return {
      success: false,
      message: "Authentication required",
      error: "Your session has expired. Please log in again",
    };
  }

  try {
    const response = await fetch(`http://localhost:8080/api/posts/${postId}`, {
      method: "DELETE",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      const errorMessage = await parseBackendError(
        response,
        "Failed to delete post",
      );
      return {
        success: false,
        message: "Backend error",
        error: errorMessage,
      };
    }

    revalidatePath(`/forum`);
    return { success: true, message: "Post deleted successfully" };
  } catch (error) {
    console.error("deletePostAction Network Error:", error);
    return {
      success: false,
      message: "Connection failure",
      error:
        "Could not reach the server. Please check your internet connection again",
    };
  }
}

export async function createCommentAction(
  formData: FormData,
): Promise<ActionState> {
  const cookieStore = await cookies();
  const token = cookieStore.get("session")?.value;
  if (!token) {
    return {
      success: false,
      message: "Authentication required",
      error: "Your session has expired. Please log in again",
    };
  }

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
    const response = await fetch(
      `http://localhost:8080/api/posts/${postId}/comments`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          content,
          parentId: parentId,
        }),
      },
    );

    if (!response.ok) {
      const errorMessage = await parseBackendError(
        response,
        "Failed to create comment",
      );
      return {
        success: false,
        message: "Failed to create comment",
        error: errorMessage,
      };
    }

    revalidatePath(`/forum/${postId}`);

    return { success: true, message: "Post comment successfully" };
  } catch (error) {
    console.error("createCommentAction Network Error:", error);
    return {
      success: false,
      message: "Connection failure",
      error:
        "Could not reach the server. Please check your internet connection again",
    };
  }
}

export async function updateCommentAction(
  commentId: string,
  content: string,
  postId: string,
): Promise<ActionState> {
  const cookieStore = await cookies();
  const token = cookieStore.get("session")?.value;
  if (!token) {
    return {
      success: false,
      message: "Authentication required",
      error: "Your session has expired. Please log in again",
    };
  }

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

  const safeData = validatedFields.data;

  try {
    const response = await fetch(
      `http://localhost:8080/api/comments/${safeData.commentId}`,
      {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ content: safeData.content }),
      },
    );

    if (!response.ok) {
      const errorMessage = await parseBackendError(
        response,
        "Failed to update comment",
      );
      return {
        success: false,
        message: "Failed to update comment",
        error: errorMessage,
      };
    }
    revalidatePath(`/forum/${safeData.postId}`);
    return { success: true, message: "Comment updated" };
  } catch (error) {
    console.error("updateCommentAction Network Error:", error);
    return {
      success: false,
      message: "Connection failure",
      error:
        "Could not reach the server. Please check your internet connection again",
    };
  }
}

export async function deleteCommentAction(
  commentId: string,
  postId: string,
): Promise<ActionState> {
  const cookieStore = await cookies();
  const token = cookieStore.get("session")?.value;
  if (!token) {
    return {
      success: false,
      message: "Authentication required",
      error: "Your session has expired. Please log in again",
    };
  }

  try {
    const response = await fetch(
      `http://localhost:8080/api/comments/${commentId}`,
      {
        method: "DELETE",
        headers: { Authorization: `Bearer ${token}` },
      },
    );
    if (!response.ok) {
      const errorMessage = await parseBackendError(
        response,
        "Failed to delete comment",
      );
      return {
        success: false,
        message: "Failed to delete comment",
        error: errorMessage,
      };
    }
    revalidatePath(`/forum/${postId}`);
    return { success: true, message: "Comment deleted successfully" };
  } catch (error) {
    console.error("deleteCommentAction Network Error:", error);
    return {
      success: false,
      message: "Connection failure",
      error:
        "Could not reach the server. Please check your internet connection again",
    };
  }
}

export async function votePostAction(
  postId: string,
  vote: 1 | -1,
): Promise<ActionState> {
  const cookieStore = await cookies();
  const token = cookieStore.get("session")?.value;
  if (!token) {
    return {
      success: false,
      message: "Authentication required",
      error: "Your session has expired. Please log in again",
    };
  }

  const validatedFields = VoteSchema.safeParse({ vote: vote });

  if (!validatedFields.success) {
    return {
      success: false,
      error: validatedFields.error.issues[0].message,
    };
  }

  const safeData = validatedFields.data;

  try {
    const response = await fetch(
      `http://localhost:8080/api/posts/${postId}/vote`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ vote: safeData.vote }),
      },
    );

    if (!response.ok) {
      const errorMessage = await parseBackendError(
        response,
        "Failed to vote post",
      );
      return {
        success: false,
        message: "Failed to vote post",
        error: errorMessage,
      };
    }
    return { success: true, message: "Post voted successfully" };
  } catch (error) {
    console.error("votePostAction Network Error:", error);
    return {
      success: false,
      message: "Connection failure",
      error:
        "Could not reach the server. Please check your internet connection again",
    };
  }
}

export async function voteCommentAction(
  commentId: string,
  vote: 1 | -1,
): Promise<ActionState> {
  const cookieStore = await cookies();
  const token = cookieStore.get("session")?.value;
  if (!token) {
    return {
      success: false,
      message: "Authentication required",
      error: "Your session has expired. Please log in again",
    };
  }

  const validatedFields = VoteSchema.safeParse({ vote: vote });

  if (!validatedFields.success) {
    return {
      success: false,
      error: validatedFields.error.issues[0].message,
    };
  }

  const safeData = validatedFields.data;

  try {
    const response = await fetch(
      `http://localhost:8080/api/comments/${commentId}/vote`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(safeData.vote),
      },
    );

    if (!response.ok) {
      const errorMessage = await parseBackendError(
        response,
        "Failed to vote comment",
      );
      return {
        success: false,
        message: "Failed to vote comment",
        error: errorMessage,
      };
    }

    return { success: true, message: "Comment voted successfully" };
  } catch (error) {
    console.error("voteCommentAction Network Error:", error);
    return {
      success: false,
      message: "Connection failure",
      error:
        "Could not reach the server. Please check your internet connection again",
    };
  }
}

export async function fetchCommentsAction(
  postId: string,
  parentId: string | null,
  pageParam: number | string = 0,
  sort: string = "desc",
): Promise<CommentNode[]> {
  const cookieStore = await cookies();
  const token = cookieStore.get("session")?.value;

  const headers = new Headers();
  // attach the token
  if (token) {
    headers.append("Authorization", `Bearer ${token}`);
  }
  // base url
  let url = `http://localhost:8080/api/posts/${postId}/comments?limit=5`;

  // dynamic pagination
  if (sort === "popular") {
    url += `&offset=${pageParam}`;
  } else if (pageParam && pageParam !== 0) {
    url += `&cursor=${encodeURIComponent(pageParam as string)}`;
  }

  // append parentId if exist
  if (parentId) url += `&parentId=${parentId}`;

  // append sort
  if (sort) url += `&sort=${sort}`;

  const res = await fetch(url, { headers, cache: "no-store" });
  if (!res.ok) {
    const errorMessage = await parseBackendError(
      res,
      "Failed to fetch comments",
    );
    throw new Error(errorMessage);
  }
  const data = await res.json();
  return data.comments;
}

export async function checkPostCategoryAction(
  postId: string,
): Promise<string | null> {
  try {
    const cookieStore = await cookies();
    const token = cookieStore.get("session")?.value;

    const headers = new Headers();
    // attach the token
    if (token) {
      headers.append("Authorization", `Bearer ${token}`);
    }
    const res = await fetch(`http://localhost:8080/api/posts/${postId}`, {
      headers,
      cache: "no-store",
    });
    if (!res.ok) {
      return null;
    }

    const data = await res.json();
    return data.post.category;
  } catch (error) {
    console.error("voteCommentAction checkPostCategoryAction Error:", error);
    return null;
  }
}

export async function triggerAiSummaryAction(
  postId: string,
): Promise<ActionState> {
  const cookieStore = await cookies();
  const token = cookieStore.get("session")?.value;
  if (!token) {
    return {
      success: false,
      message: "Authentication required",
      error: "Your session has expired. Please log in again",
    };
  }
  try {
    const response = await fetch(getApiUrl(`/api/posts/${postId}/summary`), {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      cache: "no-store",
    });

    if (!response.ok) {
      const errorMessage = await parseBackendError(
        response,
        "Failed to trigger AI summary",
      );
      return {
        success: false,
        message: "Failed to queue summary",
        error: errorMessage,
      };
    }

    return {
      success: true,
      message: "Summary generation queue successfully",
    };
  } catch (error) {
    console.error("triggerAiSummaryAction Network Error:", error);
    return {
      success: false,
      message: "Connection failure",
      error:
        "Could not reach the server. Please check your internet connection again",
    };
  }
}
