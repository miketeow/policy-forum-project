import Link from "next/link";
import {
  checkReportPendingAction,
  getLatestReportAction,
} from "../actions/report";
import { GenerateReportButton } from "./_components/generate-report-button";
import { cn } from "@/lib/utils";
import { ReportView } from "./_components/report-view";

const CATEGORIES = [
  "ECONOMY",
  "INFRASTRUCTURE",
  "HEALTHCARE",
  "EDUCATION",
  "ENVIRONMENT",
  "SAFETY",
  "OTHER",
];

export default async function ReportPage(props: {
  searchParams: Promise<{ category?: string }>;
}) {
  const searchParams = await props.searchParams;

  // url state management
  const activeCategory = searchParams.category?.toUpperCase() || "ECONOMY";

  // server side data fetching
  const [reportData, isPending] = await Promise.all([
    getLatestReportAction(activeCategory),
    checkReportPendingAction(activeCategory),
  ]);

  return (
    <div className="container py-8 max-w-6xl mx-auto space-y-8">
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">
            Civic Intelligence
          </h1>
          <p className="text-slate-500 mt-1">
            AI-aggregated consensus from public forum data
          </p>
        </div>

        {/*interactive client component*/}
        <GenerateReportButton category={activeCategory} disabled={isPending} />
      </div>

      {/*url driven tabs*/}
      <div className="flex overflow-x-auto pb-2 border-b border-slate-200 space-x-6 no-scrollbar">
        {CATEGORIES.map((cat) => {
          const isActive = cat === activeCategory;
          return (
            <Link
              key={cat}
              href={`/reports?category=${cat}`}
              className={cn(
                "whitespace-nowrap pb-3 text-sm font-medium transition-colors border-b-2",
                isActive
                  ? "border-indigo-600 text-indigo-600"
                  : "border-transparent text-slate-500 hover:text-slate-600 hover:border-slate-300",
              )}
            >
              {cat}
            </Link>
          );
        })}
      </div>

      {/*main content area*/}
      <ReportView
        category={activeCategory}
        isPending={isPending}
        reportData={reportData}
      />
    </div>
  );
}
