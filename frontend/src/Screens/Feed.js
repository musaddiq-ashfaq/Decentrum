import React, { useEffect, useState } from "react";
import Navbar from "./Navbar"; // Import Navbar
import PostReactions from "./PostReactions";
import "./UserFeed.css"; // Import the CSS file for styling

const UserFeed = () => {
  const [posts, setPosts] = useState([]);
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showSharePopup, setShowSharePopup] = useState(false);
  const [selectedPost, setSelectedPost] = useState(null);
  const [currentUser, setCurrentUser] = useState(null); // Add currentUser state

  useEffect(() => {
    // Fetch posts
    const fetchPosts = async () => {
      try {
        const response = await fetch("http://localhost:8081/feed");

        if (!response.ok) {
          const errorDetails = await response.text();
          throw new Error(
            `Failed to fetch posts: ${response.status} ${response.statusText}. Details: ${errorDetails}`
          );
        }

        const data = await response.json();
        setPosts(data);
      } catch (error) {
        console.error("Error fetching posts:", error.message);
        console.error("Error details:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchPosts();

    // Retrieve current user
    try {
      const userDataString = localStorage.getItem("user");
      if (userDataString) {
        const userData = JSON.parse(userDataString);
        setCurrentUser(userData);
        console.log("Retrieved user data:", userData);
      } else {
        setCurrentUser(null);
      }
    } catch (error) {
      console.error("Error parsing localStorage data:", error);
      setCurrentUser(null);
      localStorage.clear(); // Clear localStorage if invalid data is found
    }
  }, []);

  const fetchUsers = async () => {
    try {
      const response = await fetch("http://localhost:8081/users");
      const data = await response.json();
      setUsers(data);
    } catch (error) {
      console.error("Error fetching users:", error);
    }
  };

  const handleShareClick = (post) => {
    setSelectedPost(post);
    setShowSharePopup(true);
    fetchUsers();
  };

  const handleShareWithUser = (user) => {
    alert(`Post shared with ${user.name}`);
    setShowSharePopup(false);
  };

  if (loading) {
    return <div>Loading posts...</div>;
  }

  return (
    <div className="user-feed-container">
      <Navbar />

      <h1>User Feed</h1>
      <div className="posts-container">
        {posts.map((post) => (
          <div key={post.id} className="post-card">
            <h3>{post.user?.name || "Anonymous User"}</h3>
            <p>{post.content}</p>

            {/* Check if the post contains an image */}
            {post.imageHash && (
              <div className="post-media">
                <img
                  src={`http://localhost:8080/ipfs/${post.imageHash}`}
                  alt="Post content"
                  style={{
                    width: "100%",
                    maxHeight: "400px",
                    objectFit: "contain",
                    borderRadius: "10px",
                  }}
                />
              </div>
            )}

            {/* Check if the post contains a video */}
            {post.videoHash && (
              <div className="post-media">
                <video
                  controls
                  style={{
                    width: "100%",
                    maxHeight: "400px",
                    objectFit: "contain",
                    borderRadius: "10px",
                  }}
                >
                  <source
                    src={`http://localhost:8080/ipfs/${post.videoHash}`}
                    type="video/mp4"
                  />
                  Your browser does not support the video tag.
                </video>
              </div>
            )}

            {/* Post Reactions Component */}
            <PostReactions
              post={post}
              currentUser={currentUser}
              onReactionUpdate={(updatedPost) => {
                // Update the posts state with the new post data
                setPosts(
                  posts.map((p) => (p.id === updatedPost.id ? updatedPost : p))
                );
              }}
            />

            <div className="reactions">
              <button
                className="share-button"
                onClick={() => handleShareClick(post)}
              >
                Share
              </button>
            </div>
          </div>
        ))}
      </div>

      {showSharePopup && (
        <div className="share-popup">
          <h3>Share with:</h3>
          <div className="user-list">
            {users.map((user) => (
              <div key={user.publicKey} className="user-item">
                <span>{user.name}</span>
                <button onClick={() => handleShareWithUser(user)}>Share</button>
              </div>
            ))}
          </div>
          <button
            className="close-button"
            onClick={() => setShowSharePopup(false)}
          >
            Close
          </button>
        </div>
      )}
    </div>
  );
};

export default UserFeed;