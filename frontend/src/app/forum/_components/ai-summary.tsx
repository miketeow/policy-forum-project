"use client";

import { triggerAiSummaryAction } from "@/app/actions/forum";
import { Button } from "@/components/ui/button";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { Skeleton } from "@/components/ui/skeleton";
import { getApiUrl } from "@/lib/api-utils";
import {
  ChevronDown,
  ChevronUp,
  LogIn,
  RefreshCw,
  Sparkles,
} from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useState, useTransition } from "react";
import { toast } from "sonner";

interface AiSummaryProps {
  postId: string;
  summary: string | null;
  isLoggedIn: boolean;
}

export function AiSummary({ postId, summary, isLoggedIn }: AiSummaryProps) {
  const router = useRouter();
  // state for collapsible
  const [isOpen, setIsOpen] = useState(false);
  // state for ai worker
  const [status, setStatus] = useState<
    "LOADING" | "NONE" | "FAILED" | "COMPLETED"
  >("LOADING");
  const [isPending, startTransition] = useTransition();

  useEffect(() => {
    // if already have summary or failed previously, do nothing
    if (summary || status != "LOADING") return;

    // open the Server-Sent Events stream
    const sse = new EventSource(getApiUrl(`/api/posts/${postId}/stream`));

    // listen for the Go server
    sse.onmessage = (event) => {
      const data = JSON.parse(event.data);

      if (data.status == "COMPLETED") {
        sse.close();
        setStatus("COMPLETED");
        router.refresh();
      } else if (data.status == "FAILED") {
        sse.close();
        setStatus("FAILED");
      } else if (data.status == "NONE") {
        sse.close();
        setStatus("NONE");
      }
    };
    // if network drop, close it to prevent infinite loop
    sse.onerror = () => {
      sse.close();
      setStatus("FAILED");
    };

    // cleanup function
    return () => {
      sse.close();
    };
  }, [status, postId, router, summary]);

  const handleManualTrigger = () => {
    startTransition(async () => {
      setStatus("LOADING");
      const result = await triggerAiSummaryAction(postId);

      if (!result.success) {
        console.error(result.error);
        setStatus("FAILED");
        toast.error(result.message);
      }
    });
  };

  const isGhostJob = status == "COMPLETED" && !summary;
  const needsManualTrigger =
    status === "NONE" || status === "FAILED" || isGhostJob;

  return (
    <Collapsible
      open={isOpen}
      onOpenChange={setIsOpen}
      className="bg-blue-500/10 border border-blue-500/20 rounded-lg p-5 w-full"
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 font-semibold text-blue-600 dark:text-blue-400">
          <Sparkles className="size-4" />
          <h3 className="text-sm uppercase tracking-wider">
            AI Executive Summary
          </h3>
          {!summary && status === "LOADING" && (
            <div className="size-3 rounded-full border-2 border-blue-500 border-t-transparent animate-spin ml-2" />
          )}
        </div>

        <CollapsibleTrigger asChild>
          <Button
            variant="ghost"
            size="sm"
            className="size-8 p-0 text-blue-600 hover:bg-blue-500/20"
          >
            {isOpen ? (
              <ChevronUp className="size-4" />
            ) : (
              <ChevronDown className="size-4" />
            )}
            <span className="sr-only">Toggle AI Summary</span>
          </Button>
        </CollapsibleTrigger>
      </div>

      <CollapsibleContent>
        {/*state 1 success*/}
        {summary && (
          <p className="text-sm leading-relaxed whitespace-pre-wrap text-foreground">
            {summary}
          </p>
        )}

        {/*state 2 fallback error ui*/}
        {!summary && needsManualTrigger && (
          <div className="flex flex-col items-center justify-center p-6 border border-dashed rounded-lg border-blue-500/30 bg-blue-500/5">
            <p className="text-sm text-muted-foreground mb-4 text-center">
              {status === "FAILED"
                ? "The AI summary generation failed."
                : isGhostJob
                  ? "The previous summary attempt corrupted. Please try again"
                  : "No AI summary exists for this post."}
            </p>
            {isLoggedIn ? (
              <Button variant="outline" size="sm" onClick={handleManualTrigger}>
                <RefreshCw
                  className={`size-4 mr-2 ${isPending ? "animate-spin" : ""}`}
                />
                {isPending ? "Generating..." : "Generate Summary"}
              </Button>
            ) : (
              <Button
                variant="secondary"
                size="sm"
                onClick={() => router.push("/sign-in")}
              >
                <LogIn className="size-4 mr-2" />
                Sign in to generate summary
              </Button>
            )}
          </div>
        )}

        {/*state 3 skeleton loading ui*/}
        {!summary && !needsManualTrigger && (
          <div className="space-y-3">
            <Skeleton className="h-4 w-full bg-blue-500/20" />
            <Skeleton className="h-4 w-[95%] bg-blue-500/20" />
            <Skeleton className="h-4 w-[90%] bg-blue-500/20" />
            <Skeleton className="h-4 w-[75%] bg-blue-500/20" />
          </div>
        )}
      </CollapsibleContent>
    </Collapsible>
  );
}
