import { Badge } from "@/components/ui/badge";
import { getSession } from "@/lib/session";
import { formatDate, getCategoryColor } from "@/lib/utils";
import { CreateCommentForm } from "../_components/create-comment-form";
import { BreadcrumbNav } from "../_components/breadcumb-nav";
import { CommentSection } from "../_components/comment-section";
import { PostAction } from "../_components/post-actions";
import { VoteButton } from "../_components/vote-button";
import { PendingPostPoller } from "../_components/pending-post-poller";
import { fetchSinglePostAction } from "@/app/actions/forum";
import { ArrowBigDown, ArrowBigUp } from "lucide-react";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { AiSummary } from "../_components/ai-summary";

interface PostDetailPageProps {
  params: Promise<{ postId: string }>;
  searchParams: Promise<{ sort?: "desc" | "asc" }>;
}

export default async function PostDetailPage({
  params,
  searchParams,
}: PostDetailPageProps) {
  const user = await getSession();
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

  const isOwner = user?.id === post.author_id;
  const isLoggedIn = !!user;

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
          {user ? (
            <VoteButton
              postId={post.id}
              initialScore={post.score}
              initialUserVote={post.user_vote}
            />
          ) : (
            <div className="flex items-center gap-2">
              <div className="flex items-center gap-1 bg-muted/30 rounded-full px-1 py-1 opacity-70">
                <Button
                  variant="ghost"
                  size="icon"
                  className="size-8 rounded-full text-muted-foreground"
                  disabled
                >
                  <ArrowBigUp />
                </Button>
                <span className="text-sm font-bold min-w-5 text-center">
                  {post.score}
                </span>
                <Button
                  variant="ghost"
                  size="icon"
                  className="size-8 rounded-full text-muted-foreground"
                  disabled
                >
                  <ArrowBigDown />
                </Button>
              </div>
              <span className="text-xs text-muted-foreground font-medium">
                Log in to vote
              </span>
            </div>
          )}
        </div>
      </div>

      <div className="mt-6">
        <AiSummary
          postId={post.id}
          summary={post.summary}
          isLoggedIn={isLoggedIn}
        />
      </div>

      {/*comment section*/}
      <div className="border-t pt-8 mt-4">
        <h2 className="text-xl font-bold mb-6">Discussion</h2>

        {user ? (
          <CreateCommentForm postId={postId} />
        ) : (
          <div className="flex flex-col items-center justify-center p-6 mb-8 border-2 border-dashed rounded-lg bg-muted/10 text-center">
            <h3 className="font-medium mb-1">Join the Conversation</h3>
            <p className="text-muted-foreground text-sm mb-4">
              You must be logged in to share your thoughts on this policy.
            </p>
            <Button asChild variant="outline" size="sm">
              <Link href="/sign-in">Sign In to Comment</Link>
            </Button>
          </div>
        )}

        <CommentSection
          postId={postId}
          initialSort={sort}
          currentUserId={user?.id}
        />
      </div>
    </div>
  );
}
