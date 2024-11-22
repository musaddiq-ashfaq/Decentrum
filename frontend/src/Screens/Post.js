import React, { useState, useEffect } from 'react';
import { AlertCircle, CheckCircle2, Loader2, Send } from 'lucide-react';

// Alert Components
const Alert = ({ children, className, ...props }) => (
  <div
    role="alert"
    className={`relative w-full rounded-lg border p-4 ${className}`}
    {...props}
  >
    {children}
  </div>
);

const AlertTitle = ({ children, className, ...props }) => (
  <h5 className={`mb-1 font-medium leading-none tracking-tight ${className}`} {...props}>
    {children}
  </h5>
);

const AlertDescription = ({ children, className, ...props }) => (
  <div className={`text-sm ${className}`} {...props}>
    {children}
  </div>
);

const CreatePost = () => {
  const [content, setContent] = useState('');
  const [status, setStatus] = useState('idle');
  const [errorMessage, setErrorMessage] = useState('');
  const [wallet, setWallet] = useState(null);

  useEffect(() => {
    // Check for wallet data on component mount
    const storedWallet = localStorage.getItem('userWallet');
    if (storedWallet) {
      try {
        const parsedWallet = JSON.parse(storedWallet);
        if (parsedWallet && parsedWallet.publicKey) {
          setWallet(parsedWallet);
        }
      } catch (error) {
        console.error('Error parsing wallet data:', error);
        setStatus('error');
        setErrorMessage('Invalid wallet data. Please login again.');
      }
    }
  }, []);

  const handleSubmit = async (e) => {
    e.preventDefault();

    // Basic validation
    if (!content.trim()) {
      setStatus('error');
      setErrorMessage('Post content cannot be empty');
      return;
    }

    if (!wallet || !wallet.publicKey) {
      setStatus('error');
      setErrorMessage('Please login to create a post');
      return;
    }

    setStatus('loading');

    try {
      const postData = {
        content: content.trim(),
        wallet: {
          publicKey: wallet.publicKey,
          privateKey: wallet.privateKey // Include if needed by your backend
        },
        timestamp: new Date().toISOString(),
      };

      console.log('Sending post data:', postData); // Debug log

      const response = await fetch('http://localhost:8081/post', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(postData),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || `Error: ${response.status}`);
      }

      const result = await response.json();
      console.log('Post created:', result);

      setContent('');
      setStatus('success');

      setTimeout(() => {
        setStatus('idle');
      }, 3000);
    } catch (error) {
      console.error('Post creation failed:', error);
      setStatus('error');
      setErrorMessage(error.message || 'Failed to create post. Please try again.');
    }
  };

  const handleKeyDown = (e) => {
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
      handleSubmit(e);
    }
  };

  // Show login required message if no wallet is found
  if (!wallet) {
    return (
      <div className="max-w-2xl mx-auto p-4">
        <Alert className="bg-yellow-50 border-yellow-200">
          <AlertCircle className="h-4 w-4 text-yellow-500" />
          <AlertTitle>Login Required</AlertTitle>
          <AlertDescription>Please login to create posts</AlertDescription>
        </Alert>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto p-4 space-y-4">
      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="relative">
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="What's on your mind? (Press Ctrl + Enter to post)"
            className="w-full p-4 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none min-h-[120px]"
            disabled={status === 'loading'}
          />
          <div className="absolute bottom-3 right-3 text-gray-400 text-sm">
            {content.length} / 1000
          </div>
        </div>

        <div className="flex justify-end">
          <button
            type="submit"
            disabled={!content.trim() || status === 'loading'}
            className="bg-blue-500 text-white px-6 py-2 rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 transition-colors"
          >
            {status === 'loading' ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                Posting...
              </>
            ) : (
              <>
                <Send className="h-4 w-4" />
                Post
              </>
            )}
          </button>
        </div>
      </form>

      {status === 'success' && (
        <Alert className="bg-green-50 border-green-200">
          <CheckCircle2 className="h-4 w-4 text-green-500" />
          <AlertTitle>Success!</AlertTitle>
          <AlertDescription>Your post was created successfully</AlertDescription>
        </Alert>
      )}

      {status === 'error' && (
        <Alert className="bg-red-50 border-red-200">
          <AlertCircle className="h-4 w-4 text-red-500" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{errorMessage}</AlertDescription>
        </Alert>
      )}
    </div>
  );
};

export default CreatePost;