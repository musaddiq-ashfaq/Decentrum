import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import Modal from "./Modal";
import Navbar from "./Navbar";
import "./Userpost.css";

const UserPost = () => {
  const [posts, setPosts] = useState([]);
  const [newPost, setNewPost] = useState("");
  const [newImage, setNewImage] = useState(null);
  const [loading, setLoading] = useState(true);
  const [currentUser, setCurrentUser] = useState(null);
  const [errorMessage, setErrorMessage] = useState("");
  const [successMessage, setSuccessMessage] = useState("");
  const [showModal, setShowModal] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    try {
      const userDataString = localStorage.getItem("user");
      if (userDataString) {
        const userData = JSON.parse(userDataString);
        setCurrentUser(userData);
        console.log("Retrieved user data:", userData);
      } else {
        console.log("No user data found in localStorage");
        setCurrentUser(null);
      }
    } catch (error) {
      console.error("Error parsing user data from localStorage:", error);
      setCurrentUser(null);
      localStorage.removeItem("user");
    }
  }, []);

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
    setErrorMessage("");
    setSuccessMessage("");

    if (!newPost.trim() && !newImage) {
      setErrorMessage("Please input text or an image to post.");
      return;
    }

    if (!currentUser || !currentUser.publicKey) {
      setErrorMessage("User not authenticated. Please login again.");
      return;
    }

    try {
      // Create the request body according to the backend's expected format
      const postData = {
        content: newPost.trim(),
        publicKey: currentUser.publicKey
      };

      const response = await fetch("http://localhost:8081/post", {
        method: "POST",
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(postData),
      });
      
      if (!response.ok) {
        const errorData = await response.text();
        throw new Error(`Server responded with ${response.status}: ${errorData}`);
      }

      const createdPost = await response.json();
      setPosts([createdPost, ...posts]);
      setNewPost("");
      setNewImage(null);
      setSuccessMessage("Post created successfully!");
      setShowModal(true);
    } catch (error) {
      console.error("Error creating post:", error);
      setErrorMessage(`Failed to create post: ${error.message}`);
    }
  };

  const handleModalClose = () => {
    setShowModal(false);
    // navigate("/feed");
  };

  const handleReaction = async (postId, reaction) => {
    if (!currentUser) {
      setErrorMessage("Please login to react to posts");
      return;
    }

    const post = posts.find((p) => p.id === postId);
    const isAlreadyReacted = post.reactions && post.reactions.includes(reaction);
    let updatedReactions;

    if (isAlreadyReacted) {
      updatedReactions = post.reactions.filter((r) => r !== reaction);
    } else {
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
          publicKey: currentUser.publicKey
        }),
      });

      if (response.ok) {
        const updatedPost = await response.json();
        setPosts(
          posts.map((post) => {
            if (post.id === updatedPost.id) {
              return {
                ...post,
                reactions: updatedReactions,
                reactionCount: updatedReactions.length,
              };
            }
            return post;
          })
        );
      } else {
        throw new Error(`Failed to react to post: ${response.statusText}`);
      }
    } catch (error) {
      console.error("Error reacting to post:", error);
      setErrorMessage(error.message);
    }
  };

  if (loading) {
    return <div className="loading-spinner">Loading posts...</div>;
  }

  return (
    <div className="user-post-container">
      <Navbar />
      {currentUser ? (
        <>
          <h1>Create Post</h1>
          <form onSubmit={handlePostSubmit} className="new-post-form">
            <textarea
              value={newPost}
              onChange={(e) => setNewPost(e.target.value)}
              placeholder="What's on your mind?"
              className="post-textarea"
            />
            <button type="submit" className="submit-button">
              Post
            </button>
          </form>
          {errorMessage && <p className="error-message">{errorMessage}</p>}
        </>
      ) : (
        <p className="login-prompt">Please login to create a post</p>
      )}
      <div className="posts-container">
        {posts.map((post) => (
          <div key={post.id} className="post-card">
            <p className="post-content">{post.content}</p>
            <div className="post-footer">
              <div className="reactions">
                <button
                  className="reaction-button"
                  onClick={() => handleReaction(post.id, "like")}
                >
                  👍 Like
                </button>
                <button
                  className="reaction-button"
                  onClick={() => handleReaction(post.id, "love")}
                >
                  ❤️ Love
                </button>
                <span className="reaction-count">
                  {`${post.reactionCount || 0} ${
                    post.reactionCount === 1 ? "reaction" : "reactions"
                  }`}
                </span>
              </div>
            </div>
          </div>
        ))}
      </div>
      {showModal && (
        <Modal message={successMessage} onClose={handleModalClose} />
      )}
    </div>
  );
};

export default UserPost;