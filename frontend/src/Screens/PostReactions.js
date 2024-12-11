import React, { useState } from "react";
import { ThumbsUp, Heart, Laugh, Angry, Frown } from 'lucide-react';

const reactions = {
  like: { icon: ThumbsUp, color: "text-[#052a47]", label: "Like" },
  love: { icon: Heart, color: "text-red-500", label: "Love" },
  laugh: { icon: Laugh, color: "text-yellow-500", label: "Haha" },
  angry: { icon: Angry, color: "text-orange-500", label: "Angry" },
  sad: { icon: Frown, color: "text-purple-500", label: "Sad" },
};

const PostReactions = ({ post, currentUser, onReactionUpdate }) => {
  const [showReactionPicker, setShowReactionPicker] = useState(false);
  const [error, setError] = useState("");

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

  const currentReaction = post.reactions?.[currentUser?.publicKey];
  const totalReactions = post.reactionCount || 0;

  return (
    <div className="relative">
      {error && <div className="text-red-500 text-sm mb-2">{error}</div>}

      <button
        onClick={() => setShowReactionPicker(!showReactionPicker)}
        className="flex items-center text-gray-500 hover:text-[#4dbf38]"
      >
        {currentReaction ? (
          <>
            {React.createElement(reactions[currentReaction].icon, {
              className: `h-5 w-5 mr-1 ${reactions[currentReaction].color}`,
            })}
            <span>{reactions[currentReaction].label}</span>
          </>
        ) : (
          <>
            <ThumbsUp className="h-5 w-5 mr-1" />
            <span>React</span>
          </>
        )}
      </button>

      {totalReactions > 0 && (
        <span className="text-sm text-gray-500 ml-2">{totalReactions} Reactions</span>
      )}

      {showReactionPicker && (
        <div className="absolute bottom-full left-0 mb-2 bg-white rounded-lg shadow-md p-2 flex space-x-2">
          {Object.entries(reactions).map(([type, { icon: Icon, color, label }]) => (
            <button
              key={type}
              onClick={() => handleReaction(type)}
              className="p-2 rounded-full hover:bg-gray-100 transition-colors duration-200"
              title={label}
            >
              <Icon className={`h-6 w-6 ${color}`} />
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

export default PostReactions;

