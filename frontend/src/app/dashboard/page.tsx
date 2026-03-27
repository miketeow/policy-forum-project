import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

import { cookies } from "next/headers";
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
  const cookieStore = await cookies();
  const token = cookieStore.get("session")?.value;

  const res = await fetch("http://localhost:8080/api/users/me", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    // don't store in cache, we want fresh data everytime they load the page
    cache: "no-store",
  });

  if (!res.ok) {
    redirect("/sign-in");
  }

  const user: UserProfile = await res.json();

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
