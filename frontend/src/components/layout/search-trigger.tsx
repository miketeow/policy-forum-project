"use client";

import { useSearch } from "../providers/search-provider";
import { Button } from "../ui/button";

import { Kbd, KbdGroup } from "../ui/kbd";

export function SearchTrigger() {
  const { setIsOpen } = useSearch();

  return (
    <Button
      variant="outline"
      className="w-full justify-start text-sm text-muted-foreground sm:pr-12 md:w-40 lg:w-64"
      onClick={() => setIsOpen(true)}
    >
      <span className="hidden lg:inline-flex">Search documentation...</span>
      <span className="lg:hidden inline-flex">Search...</span>
      <KbdGroup>
        <Kbd>⌘</Kbd>
        <span>+</span>
        <Kbd>K</Kbd>
      </KbdGroup>
    </Button>
  );
}
