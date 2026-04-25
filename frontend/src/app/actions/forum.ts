"use server";
import { revalidatePath } from "next/cache";
import { cookies } from "next/headers";

export interface ActionState {
  success: boolean;
  message: string;
  error: string;
}

export async function createPostAction(
  prevState: ActionState,
  formData: FormData,
): Promise<ActionState> {
  const title = formData.get("title");
  const content = formData.get("content");

  if (!title || !content) {
    return {
      success: false,
      message: "Please try again",
      error: "Title and content are required",
    };
  }

  const cookieStore = await cookies();
  const token = cookieStore.get("session")?.value;

  if (!token) {
    return {
      success: false,
      message: "Please log in",
      error: "Unauthorized. Please log in.",
    };
  }

  try {
    // server to server
    const response = await fetch("http://localhost:8080/api/posts", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        title,
        content,
      }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      return {
        success: false,
        message: "Backend error",
        error: `${errorText}`,
      };
    }

    revalidatePath("/forum");
    return { success: true, message: "Create post successfully", error: "" };
  } catch (error) {
    return {
      success: false,
      message: "Failed to connect to the server",
      error: `${error}`,
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
      message: "Please log in",
      error: "Unauthorized",
    };
  }

  const postId = formData.get("postId") as string;
  const content = formData.get("content") as string;
  const parentId = formData.get("parentId") as string | null;

  if (!content || content.trim() === "") {
    return {
      success: false,
      message: "Comment cannot be empty",
      error: "Content cannot be empty",
    };
  }

  try {
    const res = await fetch(
      `http://localhost:8080/api/posts/${postId}/comments`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          content,
          ...(parentId ? { parent_id: parentId } : {}),
        }),
      },
    );

    if (!res.ok) {
      const errorText = await res.text();
      let parsedError = "Failed to create comment";
      try {
        const errData = JSON.parse(errorText);
        parsedError = errData.error || errData.message || errorText;
      } catch (e) {
        parsedError = errorText;
      }

      return {
        success: false,
        message: "Failed to create comment",
        error: parsedError,
      };
    }

    revalidatePath(`/forum/${postId}`);

    return { success: true, message: "Post comment successfully", error: "" };
  } catch (error) {
    console.error("Action error: ", error);
    return {
      success: false,
      message: "Failed to connect to the server",
      error: `${error}`,
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
      message: "Unauthorized",
      error: "Unauthorized",
    };
  }

  try {
    const res = await fetch(`http://localhost:8080/api/comments/${commentId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ content }),
    });

    if (!res.ok) throw new Error("Failed to update comment");
    revalidatePath(`/forum/${postId}`);
    return { success: true, message: "Comment updated", error: "" };
  } catch (error) {
    return { success: false, message: "Error", error: `${error}` };
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
      message: "Unauthorized",
      error: "Unauthorized",
    };
  }

  try {
    const res = await fetch(`http://localhost:8080/api/comments/${commentId}`, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) throw new Error("Failed to delete comment");
    revalidatePath(`/forum/${postId}`);
    return { success: true, message: "Comment deleted", error: "" };
  } catch (error) {
    return { success: false, message: "Error", error: `${error}` };
  }
}
