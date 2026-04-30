"use client";

import { useSearch } from "../providers/search-provider";
import {
  Command,
  CommandDialog,
  CommandInput,
  CommandList,
} from "../ui/command";

export function CommandMenu() {
  const { isOpen, setIsOpen } = useSearch();

  return (
    <CommandDialog open={isOpen} onOpenChange={setIsOpen}>
      <Command className="w-full">
        <CommandInput placeholder="Search forum" />
        <CommandList></CommandList>
      </Command>
    </CommandDialog>
  );
}
