import { Badge } from "@/components/ui/badge";
import { getSession } from "@/lib/session";
import { formatDate } from "@/lib/utils";
import { redirect } from "next/navigation";
import { CommentNode, CommentThread } from "../_components/comment-thread";
import { CreateCommentForm } from "../_components/create-comment-form";

interface PostDetailPageProps {
  params: Promise<{ postId: string }>;
}

interface PostDetail {
  id: string;
  title: string;
  content: string;
  category: string;
  created_at: string;
  updated_at: string;
  author_id: string;
  author_name: string;
}

interface CommentsDetail {
  id: string;
  parent_id: string | null;
  content: string;
  created_at: string;
  updated_at: string;
  author_id: string;
  author_name: string;
}

async function getPostByID(postId: string): Promise<PostDetail | null> {
  try {
    const res = await fetch(`http://localhost:8080/api/posts/${postId}`, {
      cache: "no-store",
    });

    if (!res.ok) {
      if (res.status === 404) {
        return null;
      }
      throw new Error("Failed to fetch post");
    }

    return res.json();
  } catch (error) {
    console.error(error);
    return null;
  }
}

async function getCommentsByPostId(postId: string): Promise<CommentsDetail[]> {
  try {
    const res = await fetch(
      `http://localhost:8080/api/posts/${postId}/comments`,
      { cache: "no-store" },
    );
    if (!res.ok) {
      return [];
    }
    return res.json();
  } catch (error) {
    console.error("Failed to fetch comments", error);
    return [];
  }
}

// algorithm function to convert flat array to nested tree
function buildCommentTree(flatComments: any[]): CommentNode[] {
  const commentMap = new Map<string, CommentNode>();
  const rootComments: CommentNode[] = [];

  // first pass: initialize all nodes with an empty children array
  flatComments.forEach((c) => {
    commentMap.set(c.id, {
      ...c,
      parent_id: c.parent_id,
      children: [],
    });
  });

  // second pass: link children to their parents
  flatComments.forEach((c) => {
    const node = commentMap.get(c.id)!;
    if (node.parent_id) {
      const parentNode = commentMap.get(node.parent_id);
      if (parentNode) {
        parentNode.children.push(node);
      }
    } else {
      // if it has no parent, it is top level comment
      rootComments.push(node);
    }
  });

  return rootComments;
}

export default async function PostDetailPage({ params }: PostDetailPageProps) {
  const user = await getSession();
  if (!user) redirect("sign-in");

  const { postId } = await params;
  const [post, flatComments] = await Promise.all([
    getPostByID(postId),
    getCommentsByPostId(postId),
  ]);

  if (!post) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-20 text-center">
        <h1 className="text-2xl font-bold">Post Not Found</h1>
        <p className="text-muted-foreground mt-2">
          The discussion you are looking for does not exist or has been deleted.
        </p>
      </div>
    );
  }

  // build the tree
  const commentTree = buildCommentTree(flatComments);

  return (
    <div className="flex flex-col mx-auto w-full max-w-3xl px-4 gap-8">
      <div>
        <div className="flex items-center gap-2 mb-4">
          <Badge>{post.category}</Badge>
          <span className="text-sm text-muted-foreground">
            {formatDate(post.created_at)}
          </span>
        </div>

        <h1 className="text-3xl font-bold tracking-tight mb-2">{post.title}</h1>

        <p className="text-muted-foreground border-b pb-6">
          Posted by{" "}
          <span className="font-medium text-foreground">
            {post.author_name}
          </span>
        </p>

        <div className="mt-6 text-base leading-relaxed whitespace-pre-wrap">
          {post.content}
        </div>
      </div>

      {/*comment section*/}
      <div className="border-t pt-8 mt-4">
        <h2 className="text-xl font-bold mb-6">
          Discussion ({flatComments.length})
        </h2>

        {/*top level comment form*/}
        <CreateCommentForm postId={postId} />

        {/*nested comment feed*/}
        <div className="flex flex-col gap-2">
          {commentTree.length === 0 ? (
            <p className="text-muted-foreground text-sm text-center py-8 border rounded-lg bg-muted/10">
              No comments yet. Start the conversation !
            </p>
          ) : (
            commentTree.map((rootComment) => (
              <CommentThread
                comment={rootComment}
                key={rootComment.id}
                postId={postId}
              />
            ))
          )}
        </div>
      </div>
    </div>
  );
}
