import Link from 'next/link'

export default function Home() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-8">
      <h1 className="text-4xl font-bold mb-8">Welcome to Video Search App</h1>

      <div className="flex gap-4">
        <Link href="/auth/signin" className="btn">Sign In</Link>
        <Link href="/auth/signup" className="btn">Sign Up</Link>
      </div>
    </div>
  );
}
