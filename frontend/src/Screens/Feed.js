import { useEffect, useState } from "react";
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
      } finally {
        setLoading(false);
      }
    };

    fetchPosts();

    try {
      const userDataString = localStorage.getItem("user");
      if (userDataString) {
        const userData = JSON.parse(userDataString);
        setCurrentUser(userData);
      }
    } catch (error) {
      console.error("Error parsing localStorage data:", error);
    }
    
  }, []);

  const handleReactionUpdate = (updatedPost) => {
    setPosts((prevPosts) =>
      prevPosts.map((post) =>
        post.id === updatedPost.id ? updatedPost : post
      )
    );
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

            {post.imageHash && (
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
            )}

            {post.videoHash && (
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
            )}

            <PostReactions
              post={post}
              currentUser={currentUser}
              onReactionUpdate={handleReactionUpdate}
            />
          </div>
        ))}
      </div>
    </div>
  );
};

export default UserFeed;
