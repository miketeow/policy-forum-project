"use client";

import { voteCommentAction } from "@/app/actions/forum";
import { Button } from "@/components/ui/button";
import { useQueryClient } from "@tanstack/react-query";
import { ArrowBigDown, ArrowBigUp } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

interface CommentVoteButtonProps {
  commentId: string;
  initialScore: number;
  initialUserVote: number;
}

export function CommentVoteButton({
  commentId,
  initialScore,
  initialUserVote,
}: CommentVoteButtonProps) {
  const queryClient = useQueryClient();
  const [score, setScore] = useState(initialScore);
  const [userVote, setUserVote] = useState(initialUserVote);
  const [isVoting, setIsVoting] = useState(false); // prevent spam clicking
  // 2. THE FIX: Track the "previous" props
  const [prevInitialScore, setPrevInitialScore] = useState(initialScore);
  const [prevInitialVote, setPrevInitialVote] = useState(initialUserVote);

  // 3. THE FIX: Update during render if the server data changed!
  if (
    initialScore !== prevInitialScore ||
    initialUserVote !== prevInitialVote
  ) {
    setPrevInitialScore(initialScore);
    setPrevInitialVote(initialUserVote);

    setScore(initialScore);
    setUserVote(initialUserVote);
  }

  const handleVote = async (voteValue: 1 | -1) => {
    if (isVoting) return;
    setIsVoting(true);

    // calculate delta
    let delta = 0;
    let newVote: 1 | -1 | 0 = voteValue;

    if (userVote === voteValue) {
      // toggling off
      delta = -voteValue;
      newVote = 0;
    } else if (userVote === 0) {
      // first time voting
      delta = voteValue;
    } else {
      // flipping vote
      delta = voteValue * 2;
    }

    // optimistic update
    setScore((prev) => prev + delta);
    setUserVote(newVote);

    // send request to Go backend
    const res = await voteCommentAction(commentId, voteValue);

    // if fails, revert the update
    if (!res.success) {
      toast.error(res.message);
      // undo the math
      setScore((prev) => prev - delta);
      setUserVote(userVote); // set to original value
    } else {
      queryClient.invalidateQueries({ queryKey: ["comments"] });
    }

    setIsVoting(false);
  };

  return (
    <div className="flex items-center gap-1">
      <Button
        variant="ghost"
        size="icon"
        className={`size-6 rounded-full ${userVote === 1 ? "text-orange-500 bg-orange-500/10" : "text-muted-foreground hover:text-orange-500 hover:bg-orange-500/10"}`}
        onClick={() => handleVote(1)}
        disabled={isVoting}
      >
        <ArrowBigUp
          className={`size-4 ${userVote === 1 ? "fill-orange-500" : ""}`}
        />
      </Button>

      <span
        className={`text-xs font-bold min-w-3 text-center ${userVote === 1 ? "text-orange-500" : userVote === -1 ? "text-blue-500" : ""}`}
      >
        {score}
      </span>
      <Button
        variant="ghost"
        size="icon"
        className={`size-6 rounded-full ${userVote === -1 ? "text-blue-500 bg-blue-500/10" : "text-muted-foreground"}`}
        onClick={() => handleVote(-1)}
        disabled={isVoting}
      >
        <ArrowBigDown
          className={`size-4 ${userVote === -1 ? "fill-blue-500" : ""}`}
        />
      </Button>
    </div>
  );
}
