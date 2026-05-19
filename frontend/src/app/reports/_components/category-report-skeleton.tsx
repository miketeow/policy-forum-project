import { Card, CardHeader, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

export function CategoryReportSkeleton() {
  return (
    <Card className="w-full max-w-4xl shadow-sm border-slate-200">
      <CardHeader className="pb-4">
        <div className="flex justify-between items-start">
          <div className="space-y-2">
            <Skeleton className="h-8 w-64" />
            <Skeleton className="h-4 w-48" />
          </div>
          <Skeleton className="h-6 w-24 rounded-full" />
        </div>
      </CardHeader>

      <CardContent className="space-y-6">
        <div className="space-y-2">
          <Skeleton className="h-4 w-32 mb-4" />
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-3/4" />
        </div>

        <div className="space-y-2 pt-4 border-t">
          <Skeleton className="h-4 w-32 mb-4" />
          <Skeleton className="h-16 w-full rounded-md" />
        </div>
      </CardContent>
    </Card>
  );
}
