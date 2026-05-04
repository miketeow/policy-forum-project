"use server";

import { cookies } from "next/headers";
import { Post } from "../forum/_components/post-card";
import { CommentNode } from "../forum/_components/comment-thread";

// helper function to handle secure fetch
async function secureFetch(endpoint: string, pageParam: number | string = 0) {
  const cookieStore = await cookies();
  const token = cookieStore.get("session")?.value;

  const headers = new Headers();
  // attach the token
  if (token) {
    headers.append("Authorization", `Bearer ${token}`);
  }

  let url = `http://localhost:8080/api/users/me/${endpoint}?limit=10`;

  if (pageParam && pageParam !== 0) {
    url += `&cursor=${encodeURIComponent(pageParam as string)}`;
  }

  const res = await fetch(url, { headers, cache: "no-store" });
  if (!res.ok) {
    throw new Error(`Failed to fetch ${endpoint}`);
  }
  return res.json();
}

export async function fetchUserPostsAction(
  pageParam: number | string = 0,
): Promise<Post[]> {
  const data = await secureFetch("posts", pageParam);
  return data.posts;
}
export async function fetchUserCommentsAction(
  pageParam: number | string = 0,
): Promise<CommentNode[]> {
  const data = await secureFetch("comments", pageParam);
  return data.comments;
}
export async function fetchUserUpvotedPostsAction(
  pageParam: number | string = 0,
): Promise<Post[]> {
  const data = await secureFetch("upvoted/posts", pageParam);
  return data.posts;
}
export async function fetchUserUpvotedCommentsAction(
  pageParam: number | string = 0,
): Promise<CommentNode[]> {
  const data = await secureFetch("upvoted/comments", pageParam);
  return data.comments;
}
