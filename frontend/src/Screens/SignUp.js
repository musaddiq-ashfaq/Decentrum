import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import './RegistrationPage.css';

const RegistrationPage = () => {
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    password: '',
    phone: '',
    // documents: null,
  });

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData({ ...formData, [name]: value });
  };

  const handleFileChange = (e) => {
    setFormData({ ...formData, documents: e.target.files });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    const user = {
      name: formData.name,
      email: formData.email,
      phone: formData.phone,
      password: formData.password,
    };

    try {
      const response = await fetch('http://localhost:8081/signup', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(user),
      });

      if (response.ok) {
        // Redirect to /info page for collecting additional details
        window.location.href = '/info';
      } else {
        console.error('Failed to register user.');
      }
    } catch (error) {
      console.error('Error:', error);
    }
  };


  return (
    <div className="container">
      <div className="form-container">
        <h1 className="header">Create Account</h1>
        <form onSubmit={handleSubmit}>
          {<div>
            <label htmlFor="name" className="label">Name</label>
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
          </div>}
          <div>
            <label htmlFor="email" className="label">Email</label>
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
          {<div>
            <label htmlFor="phone" className="label">Phone</label>
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
          </div>}
          <div>
            <label htmlFor="password" className="label">Password</label>
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
          <button type="submit" className="submit-button">Register</button>
        </form>
        <p className="login-link">
          Already have an account?{' '}
          <Link to="/login" className="login-link-text">Log In</Link>
        </p>
      </div>
    </div>
  );
};

export default RegistrationPage;