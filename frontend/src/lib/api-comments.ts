import { CommentsDetail } from "@/app/forum/_components/comment-section";

export async function fetchComments({
  pageParam = 0,
  postId,
  parentId,
  sort,
}: {
  pageParam?: string | number;
  postId: string;
  parentId: string | null;
  sort?: "desc" | "asc";
}): Promise<CommentsDetail[]> {
  // base url
  let url = `http://localhost:8080/api/posts/${postId}/comments?limit=5`;

  // append cursor if exist
  if (pageParam) url += `&cursor=${encodeURIComponent(pageParam as string)}`;

  // append parentId if exist
  if (parentId) url += `&parentId=${parentId}`;

  // append sort
  if (sort) url += `&sort=${sort}`;

  const res = await fetch(url);
  if (!res.ok) throw new Error("Failed to fetch comments");

  return res.json();
}
