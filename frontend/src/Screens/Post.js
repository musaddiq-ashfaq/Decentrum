import React, { useEffect, useState } from "react";
import "./Userpost.css"; // Import the CSS file for styling

const UserPost = () => {
  const [posts, setPosts] = useState([]);
  const [newPost, setNewPost] = useState("");
  const [loading, setLoading] = useState(true);
  const [currentUser, setCurrentUser] = useState(null);

  useEffect(() => {
    try {
      const userDataString = localStorage.getItem("user");
      if (userDataString) {
        const userData = JSON.parse(userDataString);
        setCurrentUser(userData);
      } else {
        console.log("No user data found in localStorage");
        setCurrentUser(null);
      }
    } catch (error) {
      console.error("Error parsing user data from localStorage:", error);
      setCurrentUser(null);
      // Optionally clear the invalid data
      localStorage.removeItem("user");
    }
  }, []);

  // Rest of your useEffect for fetching posts remains the same
  useEffect(() => {
    const fetchPosts = async () => {
      try {
        const response = await fetch("http://localhost:8081/post");
        const data = await response.json();
        setPosts(data);
      } catch (error) {
        console.error("Error fetching posts:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchPosts();
  }, []);

  const handlePostSubmit = async (e) => {
    e.preventDefault();
    if (!newPost.trim() || !currentUser) return; // Check if user exists

    try {
      const response = await fetch("http://localhost:8081/post", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          user: {
            name: currentUser.name,
            email: currentUser.email,
          },
          content: newPost,
          reactions: {
            likes: "0",
          },
          reactionCount: 0,
        }),
      });

      if (response.ok) {
        const createdPost = await response.json();
        setPosts([createdPost, ...posts]);
        setNewPost("");
      } else {
        console.error("Failed to create post", response.status);
      }
    } catch (error) {
      console.error("Error creating post:", error);
    }
  };

  const handleReaction = async (postId, reaction) => {
    const post = posts.find((p) => p.id === postId);
    const isAlreadyReacted =
      post.reactions && post.reactions.includes(reaction);
    let updatedReactions;

    if (isAlreadyReacted) {
      // Remove the reaction
      updatedReactions = post.reactions.filter((r) => r !== reaction);
    } else {
      // Add the reaction
      updatedReactions = [...(post.reactions || []), reaction];
    }

    try {
      const response = await fetch(`http://localhost:8081/reactions`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          postId,
          reaction: isAlreadyReacted ? null : reaction,
        }),
      });

      if (response.ok) {
        const updatedPost = await response.json();

        // Update the local posts state with the new reactions and counts
        setPosts(
          posts.map((post) => {
            if (post.id === updatedPost.id) {
              return {
                ...post,
                reactions: updatedReactions,
                reactionCount: updatedReactions.length, // Update the count based on the new reactions
              };
            }
            return post;
          })
        );
      } else {
        console.error("Failed to react to post", response.status);
      }
    } catch (error) {
      console.error("Error reacting to post:", error);
    }
  };

  const handleShare = async (postId) => {
    try {
      const response = await fetch(`http://localhost:8081/share`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ postId }),
      });

      if (response.ok) {
        const sharedPost = await response.json();
        setPosts([sharedPost, ...posts]); // Add shared post to the top of the feed
      } else {
        console.error("Failed to share post", response.status);
      }
    } catch (error) {
      console.error("Error sharing post:", error);
    }
  };

  if (loading) {
    return <div>Loading posts...</div>;
  }

  return (
    <div className="user-feed-container">
      {currentUser ? (
        <>
          <h1>Create Post</h1>
          <form onSubmit={handlePostSubmit} className="new-post-form">
            <textarea
              value={newPost}
              onChange={(e) => setNewPost(e.target.value)}
              placeholder="What's on your mind?"
              required
            />
            <button type="submit" className="submit-button">
              Post
            </button>
          </form>
        </>
      ) : (
        <p> Please login to create post</p>
      )}
    </div>
  );
};

export default UserPost;