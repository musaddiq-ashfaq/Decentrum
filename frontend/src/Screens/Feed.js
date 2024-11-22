import React, { useEffect, useState } from "react";
// import Navbar from "./Navbar"; // Import Navbar
import "./UserFeed.css"; // Import the CSS file for styling

const UserFeed = () => {
  const [posts, setPosts] = useState([]);
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showSharePopup, setShowSharePopup] = useState(false);
  const [selectedPost, setSelectedPost] = useState(null);

  useEffect(() => {
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
      {/* Add the Navbar at the top */}
      {/* <Navbar /> */}

      <h1>User Feed</h1>
      <div className="posts-container">
        {posts.map((post) => (
          <div key={post.id} className="post-card">
            <h3>{post.user.name}</h3>
            <p>{post.content}</p>

            {/* Check if the post contains an image */}
            {post.imageHash && (
              <div className="post-image">
                {/* Assuming the image URL is constructed from the IPFS hash */}
                <img
                  src={`http://localhost:8080/ipfs/${post.imageHash}`}
                  alt="Post content"
                  style={{ width: "100%", maxHeight: "400px", objectFit: "contain" }}
                />
              </div>
            )}

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
              <div key={user.email} className="user-item">
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
