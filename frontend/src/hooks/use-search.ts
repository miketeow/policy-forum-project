import { useQuery } from "@tanstack/react-query";
import { useDebounce } from "./use-debounce";

export interface SearchResult {
  result_type: "post" | "comment";
  unique_id: string;
  url_id: string;
  title: string;
  content: string;
  category: string;
  author_name: string;
  created_at: string;
}

export function useSearch(query: string) {
  // wait 300ms after the user stop typing before changing this value
  const debouncedQuery = useDebounce(query, 300);

  return useQuery({
    queryKey: ["search", debouncedQuery],
    queryFn: async ({ signal }) => {
      if (!debouncedQuery) return [];

      // pass signal to fetch so React Queries can abort stale requests
      const res = await fetch(
        `http://localhost:8080/api/search?q=${encodeURIComponent(debouncedQuery)}`,
        { signal },
      );

      if (!res.ok) {
        throw new Error("Search Failed");
      }
      const data = await res.json();
      return data.result as SearchResult[];
    },
    enabled: debouncedQuery.length > 0,
  });
}
