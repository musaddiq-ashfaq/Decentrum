import { useEffect, useState } from "react";
import Navbar from "./Navbar";
import PostReactions from "./PostReactions";
import { User, MessageSquare, Share2 } from 'lucide-react';

const UserFeed = () => {
  const [posts, setPosts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [currentUser, setCurrentUser] = useState(null);

  useEffect(() => {
    const fetchPosts = async () => {
      try {
        const response = await fetch("http://localhost:8081/feed");
        if (!response.ok) {
          throw new Error(`Failed to fetch posts: ${response.status} ${response.statusText}`);
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
    return <div className="flex justify-center items-center h-screen bg-gradient-animation">
      <div className="animate-spin rounded-full h-32 w-32 border-t-2 border-b-2 border-[#4dbf38]"></div>
    </div>;
  }

  return (
    <div className="bg-gradient-animation min-h-screen">
      <Navbar />
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-4xl font-bold text-center mb-10 text-[#052a47] text-shadow">User Feed</h1>
        <div className="max-w-2xl mx-auto space-y-8">
          {posts.map((post) => (
            <div key={post.id} className="bg-white rounded-lg shadow-md overflow-hidden transition-all duration-300 hover:shadow-xl">
              <div className="p-6">
                <div className="flex items-center mb-4">
                  <User className="h-10 w-10 text-[#052a47] mr-3" />
                  <div>
                    <h3 className="font-semibold text-lg text-[#052a47]">{post.user?.name || "Anonymous User"}</h3>
                    <span className="text-sm text-gray-500">{new Date(post.createdAt).toLocaleDateString()}</span>
                  </div>
                </div>
                <p className="text-gray-700 mb-4">{post.content}</p>
                {post.imageHash && (
                  <img
                    src={`http://localhost:8080/ipfs/${post.imageHash}`}
                    alt="Post content"
                    className="w-full h-48 object-cover mb-4 rounded"
                  />
                )}
                {post.videoHash && (
                  <video controls className="w-full h-48 object-cover mb-4 rounded">
                    <source
                      src={`http://localhost:8080/ipfs/${post.videoHash}`}
                      type="video/mp4"
                    />
                    Your browser does not support the video tag.
                  </video>
                )}
                <div className="flex justify-between items-center">
                  <PostReactions
                    post={post}
                    currentUser={currentUser}
                    onReactionUpdate={handleReactionUpdate}
                  />
                  <div className="flex space-x-2">
                    <button className="flex items-center text-gray-500 hover:text-[#4dbf38]">
                      <MessageSquare className="h-5 w-5 mr-1" />
                      <span>Comment</span>
                    </button>
                    <button className="flex items-center text-gray-500 hover:text-[#4dbf38]">
                      <Share2 className="h-5 w-5 mr-1" />
                      <span>Share</span>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default UserFeed;

