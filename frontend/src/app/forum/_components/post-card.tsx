"use client";

import { Badge } from "@/components/ui/badge";
import { Card, CardHeader, CardTitle } from "@/components/ui/card";
import { formatDate } from "@/lib/utils";
import Link from "next/link";
import { VoteButton } from "./vote-button";

export interface Post {
  id: string;
  title: string;
  content: string;
  category: string;
  created_at: string;
  author_name: string;
  author_id: string;
  score: number;
  user_vote: number;
}

export function PostCard({ post }: { post: Post }) {
  return (
    <Card className="transition-all hover:shadow-lg cursor-pointer relative group">
      <CardHeader>
        <div className="flex items-center justify-between mb-2">
          <Badge variant="secondary" className="text-xs">
            {post.category}
          </Badge>

          <span className="text-xs text-muted-foreground">
            {formatDate(post.created_at)}
          </span>
        </div>

        <CardTitle className="text-xl leading-tight">
          <Link href={`/forum/${post.id}`}>
            <span className="absolute inset-0" aria-hidden="true" />
            <span className="group-hover:text-primary transition-colors">
              {post.title}
            </span>
          </Link>
        </CardTitle>
        <div className="mt-4 flex items-center justify-between relative z-10">
          <div className="text-sm text-muted-foreground">
            Posted by{" "}
            <span className="font-medium text-foreground">
              {post.author_name}
            </span>
          </div>

          <VoteButton
            postId={post.id}
            initialScore={post.score}
            initialUserVote={post.user_vote}
          />
        </div>
      </CardHeader>
    </Card>
  );
}
