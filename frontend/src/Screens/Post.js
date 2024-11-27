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
  const [image, setImage] = useState(null);
  const [video, setVideo] = useState(null);
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
        setCurrentUser(null);
      }

      if (walletDataString) {
        const walletData = JSON.parse(walletDataString);
        setWallet(walletData);
      }
    } catch (error) {
      console.error("Error parsing localStorage data:", error);
      setCurrentUser(null);
      setWallet(null);
      localStorage.clear();
    }
  }, []);

  const handleLogout = () => {
    localStorage.clear();
    setCurrentUser(null);
    setWallet(null);
    navigate("/");
  };

  const handleImageChange = (e) => {
    setImage(e.target.files[0]);
    setVideo(null); // Clear video if an image is selected
  };

  const handleVideoChange = (e) => {
    setVideo(e.target.files[0]);
    setImage(null); // Clear image if a video is selected
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!content.trim() && !image && !video) {
      setStatus("error");
      setErrorMessage("Post must contain text, an image, or a video.");
      return;
    }

    if (!wallet?.publicKey) {
      setStatus("error");
      setErrorMessage("Please login to create a post.");
      return;
    }

    if (!currentUser) {
      setStatus("error");
      setErrorMessage("User information is missing.");
      return;
    }

    setStatus("loading");

    const formData = new FormData();
    formData.append("user.name",currentUser.name)
    formData.append("content", content.trim());
    formData.append("wallet.publicKey", wallet.publicKey);
    if (image) formData.append("photo", image); // Match backend field name
    if (video) formData.append("video", video); // Match backend field name

    try {
      const response = await fetch("http://localhost:8081/post", {
        method: "POST",
        body: formData,
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || `Error: ${response.status}`);
      }

      const result = await response.json();
      console.log("Post created:", result);

      setContent("");
      setImage(null);
      setVideo(null);
      setStatus("success");

      setTimeout(() => navigate("/feed"), 2000);
    } catch (error) {
      console.error("Post creation failed:", error);
      setStatus("error");
      setErrorMessage(error.message || "Failed to create post. Please try again.");
    }
  };

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
        <textarea
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="What's on your mind?"
          className="w-full p-4 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none min-h-[120px]"
          maxLength={1000}
        />
        <div className="flex gap-4">
          <input
            type="file"
            accept="image/*"
            onChange={handleImageChange}
            disabled={!!video}
          />
          <input
            type="file"
            accept="video/*"
            onChange={handleVideoChange}
            disabled={!!image}
          />
        </div>
        <button
          type="submit"
          className="bg-blue-500 text-white px-6 py-2 rounded-lg hover:bg-blue-600 flex items-center gap-2"
          disabled={status === "loading"}
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
      </form>

      {status === "success" && (
        <Alert className="bg-green-50 border-green-200">
          <CheckCircle2 className="h-4 w-4 text-green-500" />
          <AlertTitle>Success!</AlertTitle>
          <AlertDescription>Your post was created successfully.</AlertDescription>
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