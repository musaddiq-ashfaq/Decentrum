import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { AlertCircle, CheckCircle2, Loader2, LogOut, Send, Image, Video } from 'lucide-react';
import Navbar from "./Navbar";
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
    formData.append("user.name", currentUser.name)
    formData.append("content", content.trim());
    formData.append("wallet.publicKey", wallet.publicKey);
    if (image) formData.append("photo", image);
    if (video) formData.append("video", video);

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
    <div className="bg-gradient-animation min-h-screen flex flex-col">
      <Navbar />
      <div className="flex-grow flex items-center justify-center p-4">
        <div className="max-w-2xl w-full bg-white rounded-xl shadow-lg p-6 space-y-6">
          <div className="flex justify-between items-center">
            <h2 className="text-3xl font-bold text-[#052a47]">Create Post</h2>
            <button
              onClick={handleLogout}
              className="flex items-center gap-2 px-4 py-2 text-red-600 hover:text-red-700 transition-colors rounded-full hover:bg-red-50"
            >
              <LogOut className="h-5 w-5" />
              Logout
            </button>
          </div>

          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="relative">
              <textarea
                value={content}
                onChange={(e) => setContent(e.target.value)}
                placeholder="What's on your mind?"
                className="w-full p-4 border-2 border-[#4dbf38] rounded-lg focus:ring-2 focus:ring-[#80d12a] focus:border-transparent resize-none min-h-[150px] transition-all duration-300 ease-in-out"
                maxLength={1000}
              />
              <span className="absolute bottom-3 right-3 text-sm text-gray-400">
                {content.length}/1000
              </span>
            </div>
            <div className="flex gap-4">
              <label className="flex-1 flex items-center justify-center p-4 border-2 border-dashed border-[#4dbf38] rounded-lg cursor-pointer hover:bg-[#f0fdf4] transition-colors duration-300">
                <input
                  type="file"
                  accept="image/*"
                  onChange={handleImageChange}
                  disabled={!!video}
                  className="hidden"
                />
                <Image className="h-6 w-6 mr-2 text-[#4dbf38]" />
                <span className="text-[#052a47]">{image ? 'Change Image' : 'Add Image'}</span>
              </label>
              <label className="flex-1 flex items-center justify-center p-4 border-2 border-dashed border-[#4dbf38] rounded-lg cursor-pointer hover:bg-[#f0fdf4] transition-colors duration-300">
                <input
                  type="file"
                  accept="video/*"
                  onChange={handleVideoChange}
                  disabled={!!image}
                  className="hidden"
                />
                <Video className="h-6 w-6 mr-2 text-[#4dbf38]" />
                <span className="text-[#052a47]">{video ? 'Change Video' : 'Add Video'}</span>
              </label>
            </div>
            <button
              type="submit"
              className="w-full bg-[#4dbf38] text-white px-6 py-3 rounded-lg hover:bg-[#80d12a] flex items-center justify-center gap-2 transition-colors duration-300"
              disabled={status === "loading"}
            >
              {status === "loading" ? (
                <>
                  <Loader2 className="h-5 w-5 animate-spin" />
                  Posting...
                </>
              ) : (
                <>
                  <Send className="h-5 w-5" />
                  Post
                </>
              )}
            </button>
          </form>

          {status === "success" && (
            <Alert className="bg-green-50 border-green-200">
              <CheckCircle2 className="h-5 w-5 text-green-500" />
              <AlertTitle>Success!</AlertTitle>
              <AlertDescription>Your post was created successfully.</AlertDescription>
            </Alert>
          )}

          {status === "error" && (
            <Alert className="bg-red-50 border-red-200">
              <AlertCircle className="h-5 w-5 text-red-500" />
              <AlertTitle>Error</AlertTitle>
              <AlertDescription>{errorMessage}</AlertDescription>
            </Alert>
          )}
        </div>
      </div>
    </div>
  );
};

export default CreatePost;

