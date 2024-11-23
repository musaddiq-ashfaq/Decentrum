import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import './LoginPage.css';

const LoginPage = () => {
  const [publicKey, setPublicKey] = useState('');
  const [signature, setSignature] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError(''); // Clear any previous error messages
  
    try {
      console.log('Public key:', publicKey);
      console.log('Signature:', signature);
  
      const response = await fetch('http://localhost:8081/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ publicKey, signature }),
      });
  
      if (response.ok) {
        const data = await response.json(); // Parse the response from the backend
  
        console.log('Response data:', data);
  
        if (data) {
          // Store complete user data
          localStorage.setItem('user', JSON.stringify(data));
          
          // Store wallet information separately
          const walletData = {
            publicKey: data.publicKey,
            privateKey: data.privateKey || null // Include privateKey if provided
          };
          localStorage.setItem('userWallet', JSON.stringify(walletData));
  
          // Log both stored items for verification
          console.log('User data stored in localStorage:', JSON.parse(localStorage.getItem('user')));
          console.log('Wallet data stored in localStorage:', JSON.parse(localStorage.getItem('userWallet')));
  
          // Navigate to the feed page
          navigate('/feed');
        } else {
          console.error('User data missing in response:', data);
          setError('Login failed. Invalid user data returned.');
        }
      } else if (response.status === 401) {
        setError('Invalid credentials. Please check your details and try again.');
      } else {
        const errorData = await response.json();
        setError(errorData.message || 'Unexpected error occurred during login.');
      }
    } catch (error) {
      console.error('Error during login:', error);
      setError('Failed to log in. Please try again later.');
    }
  };

  return (
    <div className="container">
      <div className="form-container">
        <h1 className="header">Decentrum</h1>
        <form onSubmit={handleSubmit}>
          <div>
            <label htmlFor="publicKey" className="label">Public Key</label>
            <input
              id="publicKey"
              type="text"
              value={publicKey}
              onChange={(e) => setPublicKey(e.target.value)}
              placeholder="Enter your Public Key"
              required
              className="input"
            />
          </div>
          <div>
            <label htmlFor="signature" className="label">Signature</label>
            <input
              id="signature"
              type="text"
              value={signature}
              onChange={(e) => setSignature(e.target.value)}
              placeholder="Enter your Signature"
              required
              className="input"
            />
          </div>

          {error && <p className="error">{error}</p>}
          <button
            type="submit"
            className="submit-button"
          >
            Log In
          </button>
        </form>
        <div className="forgot-password">
          <a href="/forgot-password" className="link">Forgot Password?</a>
        </div>
        <p className="signup-text">
          Don't have an account? <a href="/signup" className="link">Sign up</a>
        </p>
      </div>
    </div>
  );
};

export default LoginPage;