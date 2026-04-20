"use client";

import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { formatDate } from "@/lib/utils";

export interface Post {
  id: string;
  title: string;
  content: string;
  category: string;
  created_at: string;
  author_name: string;
}

export function PostCard({ post }: { post: Post }) {
  return (
    <Card className="transition-all hover:shadow-md">
      <CardHeader>
        <div className="flex items-center justify-between mb-2">
          <Badge variant="secondary" className="text-xs">
            {post.category}
          </Badge>

          <span className="text-xs text-muted-foreground">
            {formatDate(post.created_at)}
          </span>
        </div>

        <CardTitle className="text-xl leading-tight">{post.title}</CardTitle>
        <CardDescription className="mt-2">
          Posted by{" "}
          <span className="font-medium text-foreground">
            {post.author_name}
          </span>
        </CardDescription>
      </CardHeader>
    </Card>
  );
}
