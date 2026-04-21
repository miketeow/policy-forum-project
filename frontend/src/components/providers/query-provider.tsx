"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useState } from "react";

export function ReactQueryProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  // use useState to make sure the QueryClient is initialized once per session,
  // so that React keep using the same client everytime the layout rerender

  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            // data is considered as fresh for 1 minute
            // after 1 minute, the vault will silently refetch it in the background
            staleTime: 60 * 1000,
            // don't automatically retry failing request in development, it is annoying
            retry: false,
          },
        },
      }),
  );

  return (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}
