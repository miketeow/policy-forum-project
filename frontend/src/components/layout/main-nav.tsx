"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { cn } from "@/lib/utils";
import { navLinks } from "@/lib/constants";

export default function MainNav() {
  const pathname = usePathname();
  return (
    <div className="mr-4 hidden md:flex">
      <Link href="/" className="mr-6 flex items-center space-x-2">
        <span className="font-serif text-2xl font-bold tracking-tight">
          Public Policy Forum
        </span>
      </Link>

      <nav className="flex items-center gap-6 text-sm font-medium font-grotesk">
        {navLinks.map((link) => {
          const isActive = pathname === link.href;
          return (
            <Link
              key={link.name}
              href={link.href}
              className={cn(
                "hover:text-foreground/80 transition-colors",
                isActive ? "text-foreground" : "text-foreground/60",
              )}
            >
              {link.name}
            </Link>
          );
        })}
      </nav>
    </div>
  );
}
