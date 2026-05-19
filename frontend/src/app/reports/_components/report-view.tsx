"use client";
import { CategoryReportData } from "@/app/actions/report";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { CategoryReportCard } from "./category-report-card";
import { FileText } from "lucide-react";
import { CategoryReportSkeleton } from "./category-report-skeleton";

interface ReportViewProps {
  category: string;
  isPending: boolean;
  reportData: {
    id: string;
    category: string;
    report: CategoryReportData;
    generated_at: string;
  } | null;
}

export function ReportView({
  category,
  isPending,
  reportData,
}: ReportViewProps) {
  const router = useRouter();

  useEffect(() => {
    if (!isPending) return;

    const intervalId = setInterval(() => {
      router.refresh();
    }, 3000);

    return () => clearInterval(intervalId);
  }, [isPending, router]);

  return (
    <div className="pt-4">
      {isPending ? (
        // skeleton
        <CategoryReportSkeleton />
      ) : reportData ? (
        <CategoryReportCard
          category={reportData.category}
          report={reportData.report}
          generatedAt={reportData.generated_at}
        />
      ) : (
        <div className="flex flex-col items-center justify-center p-16 text-center border rounded-lg bg-slate-50/50 border-dashed border-slate-300">
          <div className="bg-white p-4 rounded-full shadow-sm mb-4">
            <FileText className="size-8 text-slate-400" />
          </div>
          <h3 className="text-lg font-semibold text-slate-800">
            No Intelligence Brief
          </h3>
          <p className="text-slate-500 mt-2 max-w-sm">
            `We have not generated a consensus report for ${category} recently`
          </p>
        </div>
      )}
    </div>
  );
}
