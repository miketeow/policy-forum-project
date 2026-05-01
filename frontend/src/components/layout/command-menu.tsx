"use client";

import { useRouter } from "next/navigation";
import { useSearch as useSearchContext } from "../providers/search-provider";
import {
  Command,
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "../ui/command";
import { useState } from "react";
import { useSearch as useSearchQuery } from "@/hooks/use-search";
import { formatDate } from "@/lib/utils";

export function CommandMenu() {
  const router = useRouter();
  const { isOpen, setIsOpen } = useSearchContext();
  const [inputValue, setInputValue] = useState("");
  const [prevIsOpen, setPrevIsOpen] = useState(isOpen);

  if (isOpen !== prevIsOpen) {
    setPrevIsOpen(isOpen);
    if (!isOpen) {
      setInputValue("");
    }
  }
  const { data: results, isLoading } = useSearchQuery(inputValue);
  const handleSelect = (url: string) => {
    setInputValue("");
    setIsOpen(false);
    router.push(url);
  };

  return (
    <CommandDialog open={isOpen} onOpenChange={setIsOpen}>
      <Command className="w-full" shouldFilter={false}>
        <CommandInput
          placeholder="Search posts or comments..."
          value={inputValue}
          onValueChange={setInputValue}
        />
        <CommandList>
          {isLoading && <CommandEmpty>Searching...</CommandEmpty>}
          {!isLoading && inputValue && results?.length === 0 && (
            <CommandEmpty>No result found.</CommandEmpty>
          )}
          {results && results.length > 0 && (
            <CommandGroup heading="Results">
              {results.map((item) => (
                <CommandItem
                  key={`${item.result_type} - ${item.unique_id}`}
                  value={`${item.result_type} - ${item.unique_id}`}
                  onSelect={() => handleSelect(`/forum/${item.url_id}`)}
                  className="py-3"
                >
                  <div className="flex flex-col gap-1.5 w-full">
                    {/*title and badge*/}
                    <div className="flex items-center justify-between">
                      <span className="font-semibold text-sm truncate pr-4">
                        {item.result_type === "post" ? (
                          item.title
                        ) : (
                          <span className="text-muted-foreground font-normal">
                            Re: {item.title}
                          </span>
                        )}
                      </span>

                      <div className="flex items-center gap-2 shrink-0">
                        <span className="text-[10px] uppercase tracking-wider font-semibold text-primary bg-primary/10 px-1.5 py-0.5 rounded">
                          {item.category}
                        </span>
                        <span className="text-[10px] uppercase tracking-wider font-bold bg-muted px-1.5 py-0.5 rounded text-muted-foreground border">
                          {item.result_type}
                        </span>
                      </div>
                    </div>
                    {/*metadata*/}
                    <div className="flex items-center text-xs text-muted-foreground gap-2">
                      <span className="font-medium text-foreground/80">
                        @{item.author_name}
                      </span>
                      <span>•</span>
                      <span>{formatDate(item.created_at)}</span>
                    </div>
                    {/*content*/}
                    <span className="text-sm text-muted-foreground line-clamp-2 leading-snug border-l-2 border-primary/20 pl-2 mt-0.5">
                      {item.content}
                    </span>
                  </div>
                </CommandItem>
              ))}
            </CommandGroup>
          )}
        </CommandList>
      </Command>
    </CommandDialog>
  );
}
