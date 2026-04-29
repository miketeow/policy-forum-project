import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatDate(date: string) {
  return new Date(date).toLocaleDateString("en-US", {
    day: "numeric",
    month: "short",
    year: "numeric",
  });
}

export function getCategoryColor(category: string) {
  switch (category) {
    case "PENDING":
      return "bg-gray-500 hover:bg-gray-600 animate-pulse"; // Pulses while AI is thinking!
    case "INFRASTRUCTURE":
      return "bg-amber-600 hover:bg-amber-700";
    case "ECONOMY":
      return "bg-blue-600 hover:bg-blue-700";
    case "HEALTHCARE":
      return "bg-rose-600 hover:bg-rose-700";
    case "EDUCATION":
      return "bg-purple-600 hover:bg-purple-700";
    case "ENVIRONMENT":
      return "bg-emerald-600 hover:bg-emerald-700";
    case "SAFETY":
      return "bg-orange-600 hover:bg-orange-700";
    default:
      return "bg-slate-600 hover:bg-slate-700";
  }
}
