import React, { useEffect, useState } from 'react';
import './UserFeed.css';

const UserFeed = () => {
    const [posts, setPosts] = useState([]);
    const [users, setUsers] = useState([]);
    const [loading, setLoading] = useState(true);
    const [showSharePopup, setShowSharePopup] = useState(false);
    const [selectedPost, setSelectedPost] = useState(null);

    useEffect(() => {
        const fetchPosts = async () => {
            try {
                const response = await fetch('http://localhost:8081/feed');
                
                if (!response.ok) {
                    const errorDetails = await response.text();
                    throw new Error(`Failed to fetch posts: ${response.status} ${response.statusText}. Details: ${errorDetails}`);
                }
    
                const data = await response.json();
                setPosts(data);
            } catch (error) {
                console.error('Error fetching posts:', error.message);
                console.error('Error details:', error);
            } finally {
                setLoading(false);
            }
        };

        fetchPosts();
    }, []);

    const fetchUsers = async () => {
        try {
            const response = await fetch('http://localhost:8081/users');
            const data = await response.json();
            setUsers(data);
        } catch (error) {
            console.error('Error fetching users:', error);
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
            <h1>User Feed</h1>
            <div className="posts-container">
                {posts.map((post) => (
                    <div key={post.id} className="post-card">
                        <h3>{post.user.name}</h3>
                        <p>{post.content}</p>
                        <div className="reactions">
                            <span>{`${post.reactionCount} ${post.reactionCount === 1 ? 'reaction' : 'reactions'}`}</span>
                            <button className="reaction-button">👍 Like</button>
                            <button className="share-button" onClick={() => handleShareClick(post)}>Share</button>
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
                    <button onClick={() => setShowSharePopup(false)}>Close</button>
                </div>
            )}
        </div>
    );
};

export default UserFeed;