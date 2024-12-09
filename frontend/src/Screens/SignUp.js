'use client'

import { jsPDF } from 'jspdf'; // Import jsPDF
import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Alert, AlertDescription } from "../Components/ui/alert";
import { Button } from "../Components/ui/button";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "../Components/ui/card";
import { Input } from "../Components/ui/input";
import { Label } from "../Components/ui/label";

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
        generatePDF(data.publicKey, data.privateKey, data.signature);
        // Redirect to /login page
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

  const generatePDF = (publicKey, privateKey, signature) => {
    const doc = new jsPDF();
    doc.setFontSize(3);
    doc.text('Wallet Details', 20, 20);
    doc.text(`Public Key: ${publicKey}`, 20, 40);
    doc.text(`Private Key: ${privateKey}`, 20, 60);
    

    // Save PDF with a custom filename
    doc.save('wallet-details.pdf');
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-animation p-4">
      <Card className="w-full max-w-md bg-white/95 shadow-lg backdrop-blur-sm dark:bg-gray-800/95">
        <CardHeader>
          <CardTitle className="text-center text-3xl font-bold text-[#052a47] dark:text-white">
            Create Account
          </CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name" className="text-sm font-medium text-[#052a47] dark:text-white">
                Name
              </Label>
              <Input
                id="name"
                name="name"
                type="text"
                placeholder="Enter your name"
                required
                value={formData.name}
                onChange={handleChange}
                className="border-gray-300 focus:border-[#4dbf38] focus:ring-[#4dbf38] dark:border-gray-600 dark:bg-gray-700 dark:text-white"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="phone" className="text-sm font-medium text-[#052a47] dark:text-white">
                Phone
              </Label>
              <Input
                id="phone"
                name="phone"
                type="text"
                placeholder="Enter your phone number"
                required
                value={formData.phone}
                onChange={handleChange}
                className="border-gray-300 focus:border-[#4dbf38] focus:ring-[#4dbf38] dark:border-gray-600 dark:bg-gray-700 dark:text-white"
              />
            </div>

            {error && (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            <Button type="submit" className="w-full bg-[#052a47] hover:bg-[#03192b] text-white font-semibold py-2 px-4 rounded-md transition duration-300 ease-in-out transform hover:-translate-y-1 hover:shadow-lg">
              Register
            </Button>
          </form>
        </CardContent>
        <CardFooter className="flex flex-col space-y-2">
          <p className="text-sm text-gray-600 dark:text-gray-300">
            Already have an account?{' '}
            <Link to="/login" className="text-[#4dbf38] hover:text-[#3da029] hover:underline transition duration-300 ease-in-out">
              Log In
            </Link>
          </p>
        </CardFooter>
      </Card>
    </div>
  );
};

export default RegistrationPage;
