"use client"
import { useRouter } from 'next/navigation';
import Link from 'next/link'
// import { useRouter } from 'next/navigation';
import Image from 'next/image';
// import Link from 'next/link';

export default function SignIn() {
  const router = useRouter();

  const handleSignIn = (e) => {
    e.preventDefault();
    // Add your authentication logic here (Google/GitHub OAuth or email/password login)
    router.push('/search'); // After successful login, redirect to the search page
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
            <button type="submit" className="w-full py-2 px-4 bg-indigo-600 text-white rounded hover:bg-indigo-700">
              Sign in
            </button>
          </form>

          {/* OAuth Buttons */}
          <div className="mt-6 text-center">
            <p className="text-sm text-gray-600">Or continue with</p>
            <div className="mt-3 flex justify-center space-x-4">
              <button className="px-4 py-2 border border-gray-300 rounded flex items-center space-x-2">
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
        {/* <Image className="h-2/3 w-2/3 rounded-lg place-self-center shadow-lg"
          src="/img/image6.jpg" // Use the uploaded image here
          alt="Side Image"
          width={100} height={100}
          // layout="fill"
          // objectFit="cover"
        /> */}
      </div>
    </div>
  );
}
