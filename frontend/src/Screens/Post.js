import {
  AlertCircle,
  CheckCircle2,
  Loader2,
  LogOut,
  Send,
} from "lucide-react";
import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import "./Userpost.css";

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
  const [content, setContent] = useState("");
  const [status, setStatus] = useState("idle");
  const [errorMessage, setErrorMessage] = useState("");
  const [wallet, setWallet] = useState(null);
  const [currentUser, setCurrentUser] = useState(null);
  const navigate = useNavigate();

  useEffect(() => {
    try {
      const userDataString = localStorage.getItem("user");
      const walletDataString = localStorage.getItem("userWallet");

      if (userDataString) {
        const userData = JSON.parse(userDataString);
        setCurrentUser(userData);
        console.log("Retrieved user data:", userData);
      } else {
        console.log("No user data found in localStorage");
        setCurrentUser(null);
      }

      if (walletDataString) {
        const walletData = JSON.parse(walletDataString);
        if (walletData && walletData.publicKey) {
          setWallet(walletData);
        }
      }
    } catch (error) {
      console.error("Error parsing data from localStorage:", error);
      setCurrentUser(null);
      setWallet(null);
      localStorage.removeItem("user");
      localStorage.removeItem("userWallet");
    }
  }, []);

  const handleLogout = () => {
    // Clear localStorage
    localStorage.removeItem("user");
    localStorage.removeItem("userWallet");
    
    // Clear state
    setCurrentUser(null);
    setWallet(null);
    
    // Redirect to home page
    navigate("/");
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!content.trim()) {
      setStatus("error");
      setErrorMessage("Post content cannot be empty");
      return;
    }

    if (!wallet || !wallet.publicKey) {
      setStatus("error");
      setErrorMessage("Please login to create a post");
      return;
    }

    if (!currentUser) {
      setStatus("error");
      setErrorMessage("User information is missing");
      return;
    }

    setStatus("loading");

    try {
      const postData = {
        content: content.trim(),
        wallet: {
          publicKey: wallet.publicKey,
        },
        user: {
          name: currentUser.name || "Anonymous",
          email: currentUser.email || "Unknown",
        },
        timestamp: new Date().toISOString(),
      };

      const response = await fetch("http://localhost:8081/post", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(postData),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || `Error: ${response.status}`);
      }

      const result = await response.json();
      console.log("Post created:", result);

      setContent("");
      setStatus("success");

      // Navigate to /feed after 2 seconds
      setTimeout(() => {
        navigate("/feed");
      }, 2000);
    } catch (error) {
      console.error("Post creation failed:", error);
      setStatus("error");
      setErrorMessage(error.message || "Failed to create post. Please try again.");
    }
  };

  const handleKeyDown = (e) => {
    if ((e.ctrlKey || e.metaKey) && e.key === "Enter") {
      handleSubmit(e);
    }
  };

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
      <div className="flex justify-end">
        <button
          onClick={handleLogout}
          className="flex items-center gap-2 px-4 py-2 text-red-600 hover:text-red-700 transition-colors"
        >
          <LogOut className="h-4 w-4" />
          Logout
        </button>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="relative">
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="What's on your mind? (Press Ctrl + Enter to post)"
            className="w-full p-4 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none min-h-[120px]"
            disabled={status === "loading"}
            maxLength={1000}
          />
          <div className="absolute bottom-3 right-3 text-gray-400 text-sm">
            {content.length} / 1000
          </div>
        </div>

        <div className="flex justify-end">
          <button
            type="submit"
            disabled={!content.trim() || status === "loading"}
            className="bg-blue-500 text-white px-6 py-2 rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 transition-colors"
          >
            {status === "loading" ? (
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

      {status === "success" && (
        <Alert className="bg-green-50 border-green-200">
          <CheckCircle2 className="h-4 w-4 text-green-500" />
          <AlertTitle>Success!</AlertTitle>
          <AlertDescription>Your post was created successfully</AlertDescription>
        </Alert>
      )}

      {status === "error" && (
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