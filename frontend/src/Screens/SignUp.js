import React, { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import "./RegistrationPage.css";

const RegistrationPage = () => {
  const [formData, setFormData] = useState({
    name: "",
    email: "",
    password: "",
    phone: "",
  });
  const [error, setError] = useState("");
  const navigate = useNavigate();

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData({ ...formData, [name]: value });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError(""); // Clear any existing errors

    try {
      const response = await fetch("http://localhost:8081/signup", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(formData),
      });

      const data = await response.json();

      if (response.ok) {
        // Store user data in localStorage if available
        if (data.user) {
          localStorage.setItem("user", JSON.stringify(data.user));
        }
        navigate("/info"); // Use navigate instead of window.location
      } else {
        setError(data.message || "Failed to register user. Please try again.");
      }
    } catch (error) {
      console.error("Error:", error);
      setError("An error occurred during registration. Please try again.");
    }
  };

  return (
    <div className="container">
      <div className="form-container">
        <h1 className="header">Create Account</h1>
        {error && <p className="error">{error}</p>}
        <form onSubmit={handleSubmit}>
          {
            <div>
              <label htmlFor="name" className="label">
                Name
              </label>
              <input
                id="name"
                name="name"
                type="text"
                placeholder="Enter your name"
                required
                className="input"
                value={formData.name}
                onChange={handleChange}
              />
            </div>
          }
          <div>
            <label htmlFor="email" className="label">
              Email
            </label>
            <input
              id="email"
              name="email"
              type="email"
              placeholder="Enter your email"
              required
              className="input"
              value={formData.email}
              onChange={handleChange}
            />
          </div>
          {
            <div>
              <label htmlFor="phone" className="label">
                Phone
              </label>
              <input
                id="phone"
                name="phone"
                type="text"
                placeholder="Enter your phone number"
                required
                className="input"
                value={formData.phone}
                onChange={handleChange}
              />
            </div>
          }
          <div>
            <label htmlFor="password" className="label">
              Password
            </label>
            <input
              id="password"
              name="password"
              type="password"
              placeholder="Enter your password"
              required
              className="input"
              value={formData.password}
              onChange={handleChange}
            />
          </div>
          {/* <div>
            <label htmlFor="documents" className="label">Upload KYC Documents</label>
            <input
              id="documents"
              name="documents"
              type="file"
              accept=".jpg,.jpeg,.png,.pdf"
              onChange={handleFileChange}
              className="input"
              required
            />
          </div> */}
          <button type="submit" className="submit-button">
            Register
          </button>
        </form>
        <p className="login-link">
          Already have an account?{" "}
          <Link to="/login" className="login-link-text">
            Log In
          </Link>
        </p>
      </div>
    </div>
  );
};

export default RegistrationPage;
