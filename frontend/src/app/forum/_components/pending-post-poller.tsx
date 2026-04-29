"use client";

import { checkPostCategoryAction } from "@/app/actions/forum";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

export function PendingPostPoller({ postId }: { postId: string }) {
  const router = useRouter();

  useEffect(() => {
    const interval = setInterval(async () => {
      const currentCategory = await checkPostCategoryAction(postId);

      if (currentCategory && currentCategory != "PENDING") {
        router.refresh();
      }
    }, 2000);
    return () => clearInterval(interval);
  }, [postId, router]);

  return null;
}
