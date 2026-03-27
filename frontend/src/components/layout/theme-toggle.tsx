"use client";

import { Moon, Sun } from "lucide-react";
import { useTheme } from "next-themes";
import { Toggle } from "../ui/toggle";

export const ThemeToggle = () => {
  const { theme, setTheme } = useTheme();
  return (
    <Toggle
      onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
      variant="ghost"
      className="hover:text-foreground hover:bg-zinc-200 dark:hover:bg-zinc-800"
    >
      <Sun className="h-[1.2rem] w-[1.2rem] scale-100 rotate-0 transition-all dark:scale-0 dark:-rotate-90" />
      <Moon className="absolute h-[1.2rem] w-[1.2rem] scale-0 rotate-90 transition-all dark:scale-100 dark:rotate-0" />
      <span className="sr-only">Toggle theme</span>
    </Toggle>
  );
};
