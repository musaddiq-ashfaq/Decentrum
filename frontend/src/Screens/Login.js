import { Database } from 'lucide-react';
import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom'; // React Router hook for navigation
import { Alert, AlertDescription } from "../Components/ui/alert";
import { Button } from "../Components/ui/button";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "../Components/ui/card";
import { Input } from "../Components/ui/input";
import { Label } from "../Components/ui/label";

const LoginPage = () => {
  const [publicKey, setPublicKey] = useState('');
  const [privateKey, setPrivateKey] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    try {
      console.log('Public key:', publicKey);
      console.log('Private key:', privateKey);

      const response = await fetch('http://localhost:8081/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ publicKey, privateKey}),
      });

      if (response.ok) {
        const data = await response.json();

        console.log('Response data:', data);

        if (data) {
          localStorage.setItem('user', JSON.stringify(data));

          const walletData = {
            publicKey: data.publicKey,
            privateKey: data.privateKey || null
          };
          localStorage.setItem('userWallet', JSON.stringify(walletData));

          console.log('User data stored in localStorage:', JSON.parse(localStorage.getItem('user') || '{}'));
          console.log('Wallet data stored in localStorage:', JSON.parse(localStorage.getItem('userWallet') || '{}'));

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
    <div className="min-h-screen flex items-center justify-center bg-gradient-animation p-4">
      <Card className="w-full max-w-md bg-white/95 shadow-lg backdrop-blur-sm dark:bg-gray-800/95">
        <CardHeader className="space-y-1">
          <div className="flex items-center justify-center space-x-2">
            <Database className="h-8 w-8 text-[#4dbf38]" />
            <CardTitle className="text-3xl font-bold text-center text-[#052a47] dark:text-white">Decentrum</CardTitle>
          </div>
          <p className="text-sm text-center text-gray-600 dark:text-gray-300">Decentralized. Secure. Connected.</p>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="publicKey" className="text-sm font-medium text-[#052a47] dark:text-white">Public Key</Label>
              <Input
                id="publicKey"
                type="text"
                value={publicKey}
                onChange={(e) => setPublicKey(e.target.value)}
                placeholder="Enter your Public Key"
                required
                className="border-gray-300 focus:border-[#4dbf38] focus:ring-[#4dbf38] dark:border-gray-600 dark:bg-gray-700 dark:text-white"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="signature" className="text-sm font-medium text-[#052a47] dark:text-white">Private Key</Label>
              <Input
                id="signature"
                type="password"
                value={privateKey}
                onChange={(e) => setPrivateKey(e.target.value)}
                placeholder="Enter your Signature"
                required
                className="border-gray-300 focus:border-[#4dbf38] focus:ring-[#4dbf38] dark:border-gray-600 dark:bg-gray-700 dark:text-white"
              />
            </div>
            {error && (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}
            <Button type="submit" className="w-full bg-[#052a47] hover:bg-[#03192b] text-white font-semibold py-2 px-4 rounded-md transition duration-300 ease-in-out transform hover:-translate-y-1 hover:shadow-lg">
              Log In
            </Button>
          </form>
        </CardContent>
        <CardFooter className="flex flex-col space-y-2">
          <Link to="/forgot-password" className="text-sm text-[#4dbf38] hover:text-[#3da029] hover:underline transition duration-300 ease-in-out">
            Forgot Password?
          </Link>
          <p className="text-sm text-gray-600 dark:text-gray-300">
            Don't have an account?{' '}
            <Link to="/signup" className="text-[#4dbf38] hover:text-[#3da029] hover:underline transition duration-300 ease-in-out">
              Sign up
            </Link>
          </p>
        </CardFooter>
      </Card>
    </div>
  );
};

export default LoginPage;
