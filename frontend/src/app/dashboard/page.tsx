import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

import { getSession } from "@/lib/session";

import { redirect } from "next/navigation";

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

  return (
    <div className="p-10 max-w-4xl mx-auto py-20">
      <h1 className="text-4xl font-mono text-gray-900">
        Welcome, {user.name}!
      </h1>

      <Card className="mt-5">
        <CardHeader>
          <CardTitle className="text-sm font-mono text-gray-500 uppercase tracking-wider">
            Account Details
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="mt-4 text-gray-500 font-mono text-ellipsis">
            Email: {user.email}
          </p>
          <p className="mt-1 text-sm text-gray-500 font-mono text-ellipsis">
            ID: {user.id}
          </p>
          <p className="mt-1 text-sm text-gray-500 font-mono text-ellipsis">
            KYC Status: {user.kyc_status}
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
