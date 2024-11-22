import { useEffect, useState } from 'react';
const Share = ({ postId }) => {
  const [isSharing, setIsSharing] = useState(false);
  const [shareStatus, setShareStatus] = useState(null);
  const [selectedUser, setSelectedUser] = useState(null);
  const [post, setPost] = useState(null); // State to store fetched post details

  // Hardcoded users (you can fetch this from an API if needed)
  const users = [
    { id: '1', name: 'ship', email: 'ship1@gmail.com' },
    { id: '2', name: 'ship2', email: 'ship2@gmail.com' },
    { id: '3', name: 'ship3', email: 'ship3@gmail.com' },
  ];

  // Fetch post details based on postId
  useEffect(() => {
    const fetchPost = async () => {
      try {
        const response = await fetch(`http://localhost:8081/post/${postId}`);
        if (!response.ok) {
          throw new Error('Failed to fetch post');
        }
        const data = await response.json();
        setPost(data); // Update the post state with fetched data
      } catch (error) {
        console.error('Error fetching post:', error);
      }
    };

    if (postId) {
      fetchPost();
    }
  }, [postId]);

  const handleShare = async () => {
    if (!selectedUser) {
      alert('Please select a user to share with.');
      return;
    }

    setIsSharing(true);
    setShareStatus(null);

    // Debugging: Log postId and selectedUser details
    console.log('Attempting to share post with the following details:');
    console.log('Post ID:', postId);
    console.log('User Name:', selectedUser.name); // This will show the name of the selected user

    try {
      const response = await fetch('http://localhost:8081/share', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          postId: postId,
          userName: selectedUser.name,
          
        }),
      });
      console.log('Post ID:', postId);
      console.log('User Name:', selectedUser.name); 
      if (!response.ok) {
        throw new Error('Share failed');
      }

      setShareStatus({ type: 'success', message: 'Post shared successfully!' });
    } catch (error) {
      setShareStatus({ type: 'error', message: 'Failed to share post. Please try again.' });
    } finally {
      setIsSharing(false);
    }
  };

  return (
    <div className="flex flex-col gap-2">
      <h2>Select User to Share With</h2>
      <ul className="mb-4">
        {users.map(user => (
          <li key={user.id} className="flex items-center">
            <input
              type="radio"
              id={user.id}
              name="user"
              value={user.name}
              onChange={() => setSelectedUser(user)} // Set the selected user
              className="mr-2"
            />
            <label htmlFor={user.id}>
              {user.name} (Email: {user.email})
            </label>
          </li>
        ))}
      </ul>

      <button
        onClick={handleShare}
        disabled={isSharing}
        className="inline-flex items-center px-3 py-1 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-4 w-4 mr-2"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z"
          />
        </svg>
        {isSharing ? 'Sharing...' : 'Share'}
      </button>

      {shareStatus && (
        <div
          className={`p-2 rounded-md text-sm ${
            shareStatus.type === 'success'
              ? 'bg-green-50 text-green-700 border border-green-200'
              : 'bg-red-50 text-red-700 border border-red-200'
          }`}
        >
          {shareStatus.message}
        </div>
      )}
    </div>
  );
};

export default Share;
