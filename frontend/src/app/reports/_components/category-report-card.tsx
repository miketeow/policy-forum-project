import { CategoryReportData } from "@/app/actions/report";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";

interface CategoryReportCardProps {
  category: string;
  report: CategoryReportData;
  generatedAt: string;
}

export function CategoryReportCard({
  category,
  report,
  generatedAt,
}: CategoryReportCardProps) {
  const sentimentColor =
    report.overall_sentiment === "POSITIVE"
      ? "bg-emerald-100 text-emerald-800 hover:bg-emerald-100 border-emerald-200"
      : report.overall_sentiment === "NEGATIVE"
        ? "bg-rose-100 text-rose-800 hover:bg-rose-100 border-rose-200"
        : "bg-amber-100 text-amber-800 hover:bg-amber-100 border-amber-200";

  return (
    <Card className="w-full max-w-4xl shadow-sm border-slate-200">
      <CardHeader className="pb-4">
        <div className="flex justify-between items-start">
          <div>
            <CardTitle className="text-2xl font-bold tracking-tight">
              {category} Intelligence Brief
            </CardTitle>
            <CardDescription className="mt-1">
              Generated: {new Date(generatedAt).toLocaleString()}
            </CardDescription>
          </div>
          <Badge
            variant="outline"
            className={`text-xs px-3 py-1 uppercase tracking-wider font-semibold ${sentimentColor}`}
          >
            {report.overall_sentiment}
          </Badge>
        </div>
      </CardHeader>

      <CardContent className="space-y-6">
        <section>
          <h3 className="text-sm font-semibold text-slate-500 uppercase tracking-wider mb-2 ">
            Executive Summary
          </h3>
          <p className="text-slate-800 leading-relaxed">
            {report.trend_summary}
          </p>
        </section>

        <Separator />

        <section>
          <h3 className="text-sm font-semibold text-slate-500 uppercase tracking-wider mb-2 ">
            Actionable Insight
          </h3>
          <p className="text-indigo-950 italic font-medium">
            {report.actionable_insight}
          </p>
        </section>
        <section>
          <h3 className="text-sm font-semibold text-slate-500 uppercase tracking-wider mb-3 ">
            Recurring Themes
          </h3>
          <div className="flex flex-wrap gap-2">
            {report.key_themes.map((theme, idx) => (
              <Badge
                key={idx}
                variant="secondary"
                className="px-3 py-1 font-medium bg-slate-100 text-slate-700 hover:bg-slate-200"
              >
                {theme}
              </Badge>
            ))}
          </div>
        </section>
      </CardContent>
    </Card>
  );
}
