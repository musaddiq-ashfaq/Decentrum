import React, { useEffect, useState } from 'react';
import './UserFeed.css'; // Import the CSS file for styling

const UserFeed = () => {
    const [posts, setPosts] = useState([]);
    const [loading, setLoading] = useState(true);

    // Fetch posts from the API when the component mounts
    useEffect(() => {
        const fetchPosts = async () => {
            try {
                const response = await fetch('http://localhost:8081/feed');
                
                if (!response.ok) {
                    // Throw an error if the response status is not ok (status code 4xx or 5xx)
                    const errorDetails = await response.text(); // Get the error response body
                    throw new Error(`Failed to fetch posts: ${response.status} ${response.statusText}. Details: ${errorDetails}`);
                }
    
                const data = await response.json();
                setPosts(data);
            } catch (error) {
                console.error('Error fetching posts:', error.message);  // Print the error message
                console.error('Error details:', error);                  // Print the full error object for more details
            } finally {
                setLoading(false);
            }
        };
    
        fetchPosts();
    }, []);
    

    if (loading) {
        return <div>Loading posts...</div>;
    }

    return (
        <div className="user-feed-container">
            <h1>User Feed</h1>
            <div className="posts-container">
                {posts.map((post) => (
                    <div key={post.id} className="post-card">
                        <h3>{post.user.name}</h3>
                        <p>{post.content}</p>
                        <div className="reactions">
                            <span>{`${post.reactionCount} ${post.reactionCount === 1 ? 'reaction' : 'reactions'}`}</span>
                            <button className="reaction-button">👍 Like</button>
                            <button className="share-button">Share</button>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default UserFeed;