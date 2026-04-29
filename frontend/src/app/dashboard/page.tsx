import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

import { getSession } from "@/lib/session";
import { formatDate } from "@/lib/utils";
import { CalendarDays, Mail, ShieldCheck } from "lucide-react";

import { redirect } from "next/navigation";
import { UserPostList } from "./_components/user-post-list";
import { UserCommentList } from "./_components/user-comment-list";
import { UserUpvotedList } from "./_components/user-upvoted-list";

interface UserProfile {
  id: string;
  name: string;
  email: string;
  kyc_status: string;
  created_at: string;
  updated_at: string;
}
export default async function Dashboard() {
  const user: UserProfile = await getSession();

  if (!user) {
    redirect("/sign-in");
  }

  const initials = user.name
    ?.split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase()
    .substring(0, 2);

  return (
    <div className="max-w-4xl mx-auto px-4 py-8 flex flex-col gap-8">
      {/*profile header*/}
      <Card className="border-none shadow-md bg-linear-to-br from-card to-muted/50">
        <CardContent className="p-6 sm:p-8 flex flex-col sm:flex-row items-center sm:items-start gap-6">
          <Avatar className="size-24 border-4 border-background shadow-sm">
            <AvatarFallback className="text-3xl bg-primary text-primary-foreground font-semibold">
              {initials}
            </AvatarFallback>
          </Avatar>

          <div className="flex flex-col items-center sm:items-start flex-1 text-center sm:text-left gap-2">
            <div className="flex items-center gap-3 flex-col sm:flex-row">
              <h1 className="text-3xl font-bold tracking-tight">{user.name}</h1>
              <Badge
                variant={
                  user.kyc_status === "VERIFIED" ? "default" : "secondary"
                }
                className="flex items-center gap-1"
              >
                <ShieldCheck className="size-3" />
                {user.kyc_status}
              </Badge>
            </div>

            <div className="flex flex-col sm:flex-row gap-2 sm:gap-6 text-muted-foreground text-sm mt-1">
              <div className="flex items-center gap-2 justify-center sm:justify-start">
                <Mail className="size-4" />
                <span>{user.email}</span>
              </div>

              <div className="items-center flex gap-2 justify-center sm:justify-start">
                <CalendarDays className="size-4" />
                <span>Joined {formatDate(user.created_at)}</span>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/*dashboard tabs*/}
      <Tabs defaultValue="posts" className="w-full">
        <TabsList className="grid w-full grid-cols-3 max-w-md mx-auto sm:mx-0">
          <TabsTrigger value="posts">My Posts</TabsTrigger>
          <TabsTrigger value="comments">My Comments</TabsTrigger>
          <TabsTrigger value="upvoted">Upvoted</TabsTrigger>
        </TabsList>

        {/*tab 1 my posts*/}
        <TabsContent value="posts" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle>Discussions</CardTitle>
              <CardDescription>
                Posts and discussions you have started
              </CardDescription>
            </CardHeader>
            <CardContent>
              <UserPostList />
            </CardContent>
          </Card>
        </TabsContent>

        {/*tab 2 my comments*/}
        <TabsContent value="comments" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle>My Comments</CardTitle>
              <CardDescription>Your replies across the forum</CardDescription>
            </CardHeader>
            <CardContent>
              <UserCommentList currentUserId={user.id} />
            </CardContent>
          </Card>
        </TabsContent>
        {/*tab 1 my posts*/}
        <TabsContent value="upvoted" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle>Upvoted Content</CardTitle>
              <CardDescription>
                Posts and comments you found valuable
              </CardDescription>
            </CardHeader>
            <CardContent>
              <UserUpvotedList currentUserId={user.id} />
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
