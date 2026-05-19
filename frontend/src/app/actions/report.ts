import { ActionState, fetchAPI, requireAuth } from "@/lib/api";
import { ApiError, handleActionError } from "@/lib/api-utils";

export interface CategoryReportData {
  trend_summary: string;
  overall_sentiment: "POSITIVE" | "NEGATIVE" | "DIVIDED";
  actionable_insight: string;
  key_themes: string[];
}

export interface CategoryReportResponse {
  id: string;
  category: string;
  report: CategoryReportData;
  generated_at: string;
}

// Query: fetch the latest report
export async function getLatestReportAction(
  category: string,
): Promise<CategoryReportResponse | null> {
  try {
    const data = await fetchAPI<CategoryReportResponse>(
      `/api/reports/${category}`,
    );
    return data;
  } catch (error: unknown) {
    if (error instanceof ApiError) {
      if (error.status === 404) return null;
      console.error(`[getLatestReportAction] API Error:`, error.message);
    } else if (error instanceof Error) {
      console.error(`[getLatestReportAction] Network Error:`, error.message);
    } else {
      console.error(`[getLatestReportAction] Unknown Error:`, error);
    }
    return null;
  }
}

export async function triggerReportGenerationAction(
  category: string,
): Promise<ActionState> {
  const authError = await requireAuth();
  if (authError) return authError;

  try {
    await fetchAPI(`/api/reports/${category}/generate`, { method: "POST" });
    return {
      success: true,
      message: `${category} intelligence generation queued successfully`,
    };
  } catch (error) {
    return handleActionError(error, "triggeReportGenerationAction");
  }
}

export async function checkReportPendingAction(
  category: string,
): Promise<boolean> {
  try {
    const data = await fetchAPI<{ is_pending: boolean }>(
      `/api/reports/${category}/status`,
    );
    return data.is_pending;
  } catch {
    return false;
  }
}
