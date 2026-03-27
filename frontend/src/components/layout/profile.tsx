"use client";
import { User } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { deleteSession } from "@/app/actions/auth";

interface UserData {
  name: string;
  image?: string;
  email: string;
}
export function Profile({ user }: { user: UserData }) {
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  const router = useRouter();

  const handleLogout = async () => {
    setIsLoggingOut(true);
    try {
      await deleteSession();

      router.push("/");
    } catch (error) {
      console.error("Network error during logout", error);
    } finally {
      setIsLoggingOut(false);
    }
  };
  return (
    <DropdownMenu>
      <DropdownMenuTrigger>
        <Avatar>
          <AvatarImage src={user?.image} alt="user" />
          <AvatarFallback>
            <User />
          </AvatarFallback>
        </Avatar>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="rounded-sm w-full">
        <DropdownMenuLabel>Hi, {user.name}</DropdownMenuLabel>

        <DropdownMenuSeparator />
        <DropdownMenuItem
          onClick={handleLogout}
          disabled={isLoggingOut}
          className="text-destructive cursor-pointer focus:text-destructive focus:bg-destructive/10"
        >
          {isLoggingOut ? "Logging out..." : "Logout"}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
