import { getSession } from "@/lib/session";
import { redirect } from "next/navigation";
import { CreatePostForm } from "./_components/create-post-form";

interface UserProfile {
  id: string;
  name: string;
  email: string;
  kyc_status: string;
  created_at: string;
  updated_at: string;
}
export default async function Forum() {
  const user: UserProfile = await getSession();

  if (!user) {
    redirect("/sign-in");
  }

  return (
    <div className="mx-auto w-full max-w-3xl px-4 flex flex-col gap-8">
      {/*page header*/}
      <div>
        <h1 className="text-3xl font-bold tracking-tight">
          Public Policy Forum
        </h1>
        <p className="text-muted-foreground mt-2">
          Share your ideas, discuss local policies, and engage with the
          communities
        </p>
      </div>

      {/*create post*/}
      <CreatePostForm />

      {/*the feed*/}
      <div className="flex flex-col gap-4">
        <h2 className="text-xl font-semibold pb-2">Recent Discussions</h2>

        <div className="pb-4 border rounded-md opacity-50 flex items-center justify-center h-32 bg-muted/50">
          <p className="text-muted-foreground text-sm">
            Post feed will be here...
          </p>
        </div>
      </div>
    </div>
  );
}
