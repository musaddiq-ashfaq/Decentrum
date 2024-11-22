import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import './LoginPage.css';

const LoginPage = () => {
  const [publicKey, setPublicKey] = useState('');
  const [signature, setSignature] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate();

  // const handleSubmit = async (e) => {
  //   e.preventDefault();
  //   try {
  //     console.log('Public key:', publicKey);  // Log public key input
  //     console.log('Signature:', signature);  // Log signature input
      
  //     // Send login request to the backend
  //     const response = await fetch('http://localhost:8081/login', {
  //       method: 'POST',
  //       headers: { 'Content-Type': 'application/json' },
  //       body: JSON.stringify({ publicKey, signature }),
  //     });

  //     // Check if the response is successful
  //     if (response.ok) {
  //       const data = await response.json();  // Parse the response data
        
  //       console.log('Response data:', data);  // Log the response from the backend

  //       // Ensure the user data exists in the response
  //       if (data) {
  //         // Save the user data to localStorage
  //         localStorage.setItem('user', JSON.stringify(data));
          
  //         // Verify the data stored in localStorage
  //         console.log('User data stored in localStorage:', JSON.parse(localStorage.getItem('user')));
          
  //         // Redirect to the feed page on successful login
  //         navigate('/feed');
  //       } else {
  //         // If no user data in the response, log an error
  //         console.error('User data not found in response:', data);
  //         setError('Login failed. No user data returned.');
  //       }
  //     } else if (response.status === 401) {
  //       setError('Wrong credentials. Please try again.');
  //     } else {
  //       // Handle other error responses
  //       const errorData = await response.json();
  //       alert(errorData.message || 'Login failed');
  //     }
  //   } catch (error) {
  //     console.error('Error during login:', error);
  //     setError('Failed to log in.');
  //   }
  // };
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
  
        if (data && data.publicKey) {
          // Store the user's wallet data in localStorage
          localStorage.setItem('userWallet', JSON.stringify({
            publicKey: data.publicKey,
            privateKey: data.privateKey || null, // Include privateKey if provided by the backend
          }));
  
          console.log('User wallet stored in localStorage:', JSON.parse(localStorage.getItem('userWallet')));
  
          // Navigate to the feed page
          navigate('/feed');
        } else {
          console.error('User wallet data missing in response:', data);
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
