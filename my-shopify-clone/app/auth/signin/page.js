"use client"

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import Image from 'next/image';

export default function SignIn() {
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState(''); // Add password state if needed
  const [error, setError] = useState(null);

  const handleSignIn = async (e) => {
    e.preventDefault();

    try {
      const response = await fetch('http://localhost:8080/auth/email', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: new URLSearchParams({
          email: email,
          password: password,  // Only include if you are handling passwords
        }),
      });

      const data = await response.json();
      if (data.token) {
        // Store JWT token in localStorage
        localStorage.setItem('token', data.token);
        // Redirect to the search page
        router.push('/search');
      } else {
        setError(data.error || 'Login failed');
      }
    } catch (err) {
      setError('Something went wrong, please try again.');
    }
  };

  const handleGoogleLogin = async () => {
    // Initialize Google login and obtain the ID token (integrate Google login SDK here)
    const googleIdToken = 'your-google-id-token'; // Replace this with actual Google ID token

    try {
      const response = await fetch('http://localhost:8080/auth/google', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: new URLSearchParams({
          id_token: googleIdToken,
        }),
      });

      const data = await response.json();
      if (data.token) {
        localStorage.setItem('token', data.token); // Store the JWT token
        router.push('/search'); // Redirect to search page
      } else {
        setError(data.error || 'Google login failed');
      }
    } catch (err) {
      setError('Something went wrong with Google login.');
    }
  };

  return (
    <div className="min-h-screen flex">
      {/* Left Section */}
      <div className="flex flex-col justify-center items-center w-full md:w-1/2 p-8">
        <div className="w-full max-w-md">
          {/* Company Logo */}
          <div className="mb-8 text-center">
            <Image src="/img/image.png" alt="Logo" width={60} height={60} />
          </div>
          <h2 className="text-2xl font-bold mb-6">Sign in to your account</h2>
          <p className="text-gray-600 mb-6 text-center">
            Not a member? <Link href="/auth/signup" className="text-indigo-600">Start a 14-day free trial</Link>
          </p>

          {/* Sign In Form */}
          <form onSubmit={handleSignIn} className="space-y-4">
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-700">Email address</label>
              <input
                id="email"
                name="email"
                type="email"
                autoComplete="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full p-2 border border-gray-300 rounded"
              />
            </div>
            <div>
              <label htmlFor="password" className="block text-sm font-medium text-gray-700">Password</label>
              <input
                id="password"
                name="password"
                type="password"
                autoComplete="current-password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full p-2 border border-gray-300 rounded"
              />
            </div>
            <div className="flex justify-between items-center">
              <div className="flex items-center">
                <input id="remember-me" name="remember-me" type="checkbox" className="h-4 w-4" />
                <label htmlFor="remember-me" className="ml-2 block text-sm text-gray-700">Remember me</label>
              </div>
              <Link href="/auth/forgot-password" className="text-sm text-indigo-600">Forgot password?</Link>
            </div>
            {error && <p className="text-red-500">{error}</p>}
            <button type="submit" className="w-full py-2 px-4 bg-indigo-600 text-white rounded hover:bg-indigo-700">
              Sign in
            </button>
          </form>

          {/* OAuth Buttons */}
          <div className="mt-6 text-center">
            <p className="text-sm text-gray-600">Or continue with</p>
            <div className="mt-3 flex justify-center space-x-4">
              <button
                className="px-4 py-2 border border-gray-300 rounded flex items-center space-x-2"
                onClick={handleGoogleLogin}
              >
                <Image src="/img/google-icon.png" alt="Google Icon" width={20} height={20} />
                <span>Google</span>
              </button>
              <button className="px-4 py-2 border border-gray-300 rounded flex items-center space-x-2">
                <Image src="/img/github-icon.png" alt="GitHub Icon" width={20} height={20} />
                <span>GitHub</span>
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Right Section with Image */}
      <div className="hidden md:block md:w-1/2 bg-gray-200 p-25">
        {/* Placeholder for the right image */}
      </div>
    </div>
  );
}
