import React, { useEffect, useState } from 'react';
import './Userpost.css'; // Import the CSS file for styling

const UserPost = () => {
    const [posts, setPosts] = useState([]);
    const [newPost, setNewPost] = useState('');
    const [loading, setLoading] = useState(true);

    // Fetch posts from the API when the component mounts
    useEffect(() => {
        const fetchPosts = async () => {
            try {
                const response = await fetch('http://localhost:8081/post'); // Replace with your actual endpoint
                const data = await response.json();
                setPosts(data);
            } catch (error) {
                console.error('Error fetching posts:', error);
            } finally {
                setLoading(false);
            }
        };

        fetchPosts();
    }, []);

    const handlePostChange = (e) => {
        setNewPost(e.target.value);
    };

    const handlePostSubmit = async (e) => {
        e.preventDefault();
        if (!newPost.trim()) return; // Prevent empty submissions

        try {
            const response = await fetch('http://localhost:8081/post', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    user: {
                      name: "user2",  // Replace with dynamic user info if needed
                      email: "user2@gmail.com"
                    },
                    content: newPost,
                    reactions: {
                      likes: "0" // Default value, can be updated dynamically
                    },
                    reactionCount: 0
                  }),
            });

            if (response.ok) {
                const createdPost = await response.json();
                setPosts([createdPost, ...posts]); // Add new post to the top of the feed
                setNewPost(''); // Clear the input field
            } else {
                console.error('Failed to create post', response.status);
                const errorResponse = await response.json();
                console.error('Failed to create post', response.status, errorResponse);
            }
        } catch (error) {
            console.error('Error creating post:', error);
        }
    };

    const handleReaction = async (postId, reaction) => {
        const post = posts.find(p => p.id === postId);
        const isAlreadyReacted = post.reactions && post.reactions.includes(reaction);
        let updatedReactions;

        if (isAlreadyReacted) {
            // Remove the reaction
            updatedReactions = post.reactions.filter(r => r !== reaction);
        } else {
            // Add the reaction
            updatedReactions = [...(post.reactions || []), reaction];
        }

        try {
            const response = await fetch(`http://localhost:8081/reactions`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    postId,
                    reaction: isAlreadyReacted ? null : reaction,
                }),
            });

            if (response.ok) {
                const updatedPost = await response.json();

                // Update the local posts state with the new reactions and counts
                setPosts(posts.map(post => {
                    if (post.id === updatedPost.id) {
                        return {
                            ...post,
                            reactions: updatedReactions,
                            reactionCount: updatedReactions.length // Update the count based on the new reactions
                        };
                    }
                    return post;
                }));
            } else {
                console.error('Failed to react to post', response.status);
            }
        } catch (error) {
            console.error('Error reacting to post:', error);
        }
    };

    const handleShare = async (postId) => {
        try {
            const response = await fetch(`http://localhost:8081/share`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ postId }),
            });

            if (response.ok) {
                const sharedPost = await response.json();
                setPosts([sharedPost, ...posts]); // Add shared post to the top of the feed
            } else {
                console.error('Failed to share post', response.status);
            }
        } catch (error) {
            console.error('Error sharing post:', error);
        }
    };

    if (loading) {
        return <div>Loading posts...</div>;
    }

    return (
        <div className="user-feed-container">
            <h1>Create Post</h1>
            <form onSubmit={handlePostSubmit} className="new-post-form">
                <textarea
                    value={newPost}
                    onChange={handlePostChange}
                    placeholder="What's on your mind?"
                    required
                />
                <button type="submit" className="submit-button">Post</button>
            </form>
            <div className="posts-container">
                {posts.map((post) => (
                    <div key={post.id} className="post-card">
                        <p>{post.content}</p>
                        <div className="reactions">
                            <span className="reaction-button" onClick={() => handleReaction(post.id, 'like')}>👍</span>
                            <div className="reaction-options">
                                <button className="reaction-button" onClick={() => handleReaction(post.id, 'like')}>👍 Like</button>
                                <button className="reaction-button" onClick={() => handleReaction(post.id, 'love')}>❤️ Love</button>
                                {/* Add more reaction options here if needed */}
                            </div>
                            <span>{post.reactions ? post.reactions.join(', ') : ''}</span>
                            <span className="reaction-count">
                                {` (${post.reactionCount || 0} ${post.reactionCount === 1 ? 'reaction' : 'reactions'})`}
                            </span>
                        </div>
                        <button onClick={() => handleShare(post.id)} className="share-button">Share</button>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default UserPost;
