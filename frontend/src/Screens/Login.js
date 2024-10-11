import React, { useState } from "react";
import { useNavigate } from "react-router-dom";

const LoginPage = () => {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError(""); // Clear any existing errors

    try {
      const response = await fetch("http://localhost:8081/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });

      const data = await response.json();

      if (response.ok) {
        if (data.user) {
          // Store user data in localStorage
          localStorage.setItem("user", JSON.stringify(data.user));
          navigate("/feed");
        } else {
          setError("Login successful but user data is missing");
        }
      } else {
        setError(data.message || "Wrong credentials. Please try again.");
      }
    } catch (error) {
      console.error("Error during login:", error);
      setError("Failed to log in. Please check your connection and try again.");
    }
  };

  return (
    <div className="container">
      <div className="form-container">
        <h1 className="header">Facebook</h1>
        <form onSubmit={handleSubmit}>
          <div>
            <label htmlFor="email" className="label">
              Email or Phone
            </label>
            <input
              id="email"
              type="text"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="Enter your email or phone"
              required
              className="input"
            />
          </div>
          <div>
            <label htmlFor="password" className="label">
              Password
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter your password"
              required
              className="input"
            />
          </div>
          {error && <p className="error">{error}</p>}
          <button type="submit" className="submit-button">
            Log In
          </button>
        </form>
        <div className="forgot-password">
          <a href="/forgot-password" className="link">
            Forgot Password?
          </a>
        </div>
        <p className="signup-text">
          Don't have an account?{" "}
          <a href="/signup" className="link">
            Sign up
          </a>
        </p>
      </div>
    </div>
  );
};

export default LoginPage;
