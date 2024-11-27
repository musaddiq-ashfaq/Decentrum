import { Angry, Frown, Heart, Laugh, ThumbsUp } from "lucide-react";
import React, { useState } from "react";
import { Alert, AlertDescription } from "../Components/Alert";

const PostReactions = ({ post, currentUser, onReactionUpdate }) => {
  console.log("Current User in PostReactions:", currentUser);
  const [showReactionPicker, setShowReactionPicker] = useState(false);
  const [error, setError] = useState("");

  const reactions = {
    like: { icon: ThumbsUp, color: "text-blue-500", label: "Like" },
    love: { icon: Heart, color: "text-red-500", label: "Love" },
    laugh: { icon: Laugh, color: "text-yellow-500", label: "Haha" },
    angry: { icon: Angry, color: "text-orange-500", label: "Angry" },
    sad: { icon: Frown, color: "text-purple-500", label: "Sad" },
  };

  const handleReaction = async (reactionType) => {
    if (!currentUser?.publicKey) {
      setError("Please login to react to posts");
      return;
    }

    try {
      const response = await fetch(
        `http://localhost:8081/post/${post.id}/react`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            userPublicKey: currentUser.publicKey,
            reactionType: reactionType,
          }),
        }
      );

      if (!response.ok) {
        throw new Error("Failed to add reaction");
      }

      const updatedPost = await response.json();
      onReactionUpdate(updatedPost);
      setShowReactionPicker(false);
    } catch (err) {
      setError("Failed to add reaction. Please try again.");
    }
  };

  const getUserReaction = () => {
    return post.reactions?.[currentUser?.publicKey];
  };

  const currentReaction = getUserReaction();
  const totalReactions = post.reactionCount || 0;

  return (
    <div className="relative">
      {error && (
        <Alert variant="destructive" className="mb-2">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="flex items-center gap-2">
        <button
          onClick={() => setShowReactionPicker(!showReactionPicker)}
          className="flex items-center gap-1 px-3 py-1 rounded-md hover:bg-gray-100 transition-colors"
        >
          {currentReaction ? (
            <>
              {React.createElement(reactions[currentReaction].icon, {
                className: `w-5 h-5 ${reactions[currentReaction].color}`,
              })}
              <span className={reactions[currentReaction].color}>
                {reactions[currentReaction].label}
              </span>
            </>
          ) : (
            <>
              <ThumbsUp className="w-5 h-5 text-gray-500" />
              <span className="text-gray-500">React</span>
            </>
          )}
        </button>

        {totalReactions > 0 && (
          <span className="text-sm text-gray-500">
            {totalReactions} {totalReactions === 1 ? "reaction" : "reactions"}
          </span>
        )}
      </div>

      {showReactionPicker && (
        <div className="absolute bottom-full left-0 mb-2 p-2 bg-white rounded-lg shadow-lg border flex gap-2 animate-in slide-in-from-bottom-2">
          {Object.entries(reactions).map(
            ([type, { icon: Icon, color, label }]) => (
              <button
                key={type}
                onClick={() => handleReaction(type)}
                className="group flex flex-col items-center p-2 rounded-lg hover:bg-gray-100 transition-colors"
                title={label}
              >
                <Icon
                  className={`w-6 h-6 ${color} group-hover:scale-125 transition-transform`}
                />
              </button>
            )
          )}
        </div>
      )}
    </div>
  );
};

export default PostReactions;