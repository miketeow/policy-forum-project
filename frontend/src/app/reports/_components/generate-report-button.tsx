"use client";

import { triggerReportGenerationAction } from "@/app/actions/report";
import { Button } from "@/components/ui/button";
import { Loader2, Sparkle } from "lucide-react";
import { useRouter } from "next/navigation";
import { useTransition } from "react";
import { toast } from "sonner";

interface GenerateReportButtonProps {
  category: string;
  disabled?: boolean;
}

export function GenerateReportButton({
  category,
  disabled = false,
}: GenerateReportButtonProps) {
  const [isPending, startTransition] = useTransition();
  const router = useRouter();

  const handleGenerate = () => {
    startTransition(async () => {
      const result = await triggerReportGenerationAction(category);

      if (result.success) {
        toast.success("Generation Queued", {
          description: result.message,
        });

        router.refresh();
      } else {
        toast.error("Generation Failed", {
          description: result.error,
        });
      }
    });
  };

  const isButtonLocked = isPending || disabled;

  return (
    <Button
      onClick={handleGenerate}
      disabled={isButtonLocked}
      className="bg-indigo-600 hover:bg-indigo-700 text-white shadow-sm"
    >
      {isPending ? (
        <Loader2 className="mr-2 size-4 animate-spin" />
      ) : (
        <Sparkle className="mr-2 size-4" />
      )}
      {
        disabled
          ? "Analyzing Data..." // DB says a job is running
          : isPending
            ? "Queuing..." // React says we are sending the request
            : "Queue New Generation" // Idle state
      }
    </Button>
  );
}
