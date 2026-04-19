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
