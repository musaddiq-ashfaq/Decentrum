import React, { useState } from "react";
import {
  ThumbsUp,
  Heart,
  Laugh,
  Angry,
  Frown,
} from "lucide-react";
import "./PostReactions.css"; // Optional for additional styling

const reactions = {
  like: { icon: ThumbsUp, color: "text-blue-500", label: "Like" },
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
    <div className="post-reactions-container">
      {error && <div className="error-message">{error}</div>}

      <button
        onClick={() => setShowReactionPicker(!showReactionPicker)}
        className="reaction-button"
      >
        {currentReaction ? (
          <>
            {React.createElement(reactions[currentReaction].icon, {
              className: reactions[currentReaction].color,
            })}
            {reactions[currentReaction].label}
          </>
        ) : (
          <>
            <ThumbsUp className="text-gray-500" />
            React
          </>
        )}
      </button>

      {totalReactions > 0 && (
        <span className="total-reactions">{totalReactions} Reactions</span>
      )}

      {showReactionPicker && (
        <div className="reaction-picker">
          {Object.entries(reactions).map(
            ([type, { icon: Icon, color, label }]) => (
              <button
                key={type}
                onClick={() => handleReaction(type)}
                className="reaction-option"
                title={label}
              >
                <Icon className={color} />
                {label}
              </button>
            )
          )}
        </div>
      )}
    </div>
  );
};

export default PostReactions;
