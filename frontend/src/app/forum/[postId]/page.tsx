import { Badge } from "@/components/ui/badge";
import { getSession } from "@/lib/session";
import { formatDate, getCategoryColor } from "@/lib/utils";
import { redirect } from "next/navigation";
import { CreateCommentForm } from "../_components/create-comment-form";
import { BreadcrumbNav } from "../_components/breadcumb-nav";
import { CommentSection } from "../_components/comment-section";
import { PostAction } from "../_components/post-actions";
import { VoteButton } from "../_components/vote-button";
import { PendingPostPoller } from "../_components/pending-post-poller";
import { fetchSinglePostAction } from "@/app/actions/forum";

interface PostDetailPageProps {
  params: Promise<{ postId: string }>;
  searchParams: Promise<{ sort?: "desc" | "asc" }>;
}

export default async function PostDetailPage({
  params,
  searchParams,
}: PostDetailPageProps) {
  const user = await getSession();
  if (!user) redirect("sign-in");
  const { postId } = await params;
  const { sort: sortQuery } = await searchParams;

  const sort = sortQuery === "asc" ? "asc" : "desc";
  const post = await fetchSinglePostAction(postId);

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

  const breadcrumbs = [
    { label: "Forum", href: "/forum" },
    { label: post.title },
  ];

  const isOwner = user.id === post.author_id;

  return (
    <div className="flex flex-col mx-auto w-full max-w-3xl px-4 gap-8 py-8">
      {post.category === "PENDING" && <PendingPostPoller postId={post.id} />}
      <BreadcrumbNav items={breadcrumbs} />
      <div>
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <Badge
              className={`text-xs text-white border-none transition-colors ${getCategoryColor(post.category)}`}
            >
              {post.category}
            </Badge>
            <span className="text-sm text-muted-foreground">
              {formatDate(post.created_at)}
            </span>
          </div>

          {isOwner && (
            <PostAction
              postId={post.id}
              initialTitle={post.title}
              initialContent={post.content}
            />
          )}
        </div>

        <h1 className="text-3xl font-bold tracking-tight mb-2">{post.title}</h1>

        <p className="text-muted-foreground">
          Posted by{" "}
          <span className="font-medium text-foreground">
            {post.author_name}
          </span>
        </p>

        <div className="mt-6 text-base leading-relaxed whitespace-pre-wrap">
          {post.content}
        </div>

        <div className="flex items-center mt-6 border-b pb-6">
          <VoteButton
            postId={post.id}
            initialScore={post.score}
            initialUserVote={post.user_vote}
          />
        </div>
      </div>

      {/*comment section*/}
      <div className="border-t pt-8 mt-4">
        <h2 className="text-xl font-bold mb-6">Discussion</h2>

        <CreateCommentForm postId={postId} />

        <CommentSection
          postId={postId}
          initialSort={sort}
          currentUserId={user.id}
        />
      </div>
    </div>
  );
}
