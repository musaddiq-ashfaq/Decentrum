import { jsPDF } from 'jspdf'; // Import jsPDF
import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import './RegistrationPage.css';

const RegistrationPage = () => {
  const [formData, setFormData] = useState({
    name: '',
    phone: '',
  });

  const [wallet, setWallet] = useState({
    publicKey: '',
    privateKey: '',
    signature: '',
  });

  const [error, setError] = useState('');

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
      phone: formData.phone,
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
        const data = await response.json();
        setWallet({
          publicKey: data.publicKey,
          privateKey: data.privateKey,
          signature: data.signature,
        });
        // Generate PDF after successful registration
        generatePDF(data.publicKey, data.privateKey,data.signature);
        // Redirect to /info page for collecting additional details
        window.location.href = '/login';
      } else {
        const errorData = await response.json();
        setError(errorData.error || 'Failed to register user.');
      }
    } catch (error) {
      console.error('Error:', error);
      setError('Error occurred while registering.');
    }
  };

  const generatePDF = (publicKey, privateKey,signature) => {
    const doc = new jsPDF();
    doc.setFontSize(6);
    doc.text('Wallet Details', 20, 20);
    doc.setFontSize(6);
    doc.text(`Public Key: ${publicKey}`, 20, 40);
    doc.text(`Private Key: ${privateKey}`, 20, 60);
    doc.text(`Signature: ${signature}`, 20, 80);

    // Save PDF with a custom filename
    doc.save('wallet-details.pdf');
  };

  return (
    <div className="container">
      <div className="form-container">
        <h1 className="header">Create Account</h1>
        <form onSubmit={handleSubmit}>
          <div>
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
          </div>
          
          <div>
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
          </div>
         
         
          <button type="submit" className="submit-button">Register</button>
        </form>

        
        {error && <p className="error-message">{error}</p>}

        <p className="login-link">
          Already have an account?{' '}
          <Link to="/login" className="login-link-text">Log In</Link>
        </p>
      </div>
    </div>
  );
};

export default RegistrationPage;
