import { useRouter } from 'next/navigation';

export default function SignUp() {
  const router = useRouter();

  const handleSignUp = (e) => {
    e.preventDefault();
    // Add sign-up logic here
    // After sign-up, redirect to the search page
    router.push('/search');
  };

  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-8">
      <h1 className="text-2xl font-bold mb-8">Sign Up</h1>

      <form className="flex flex-col gap-4" onSubmit={handleSignUp}>
        <input
          type="text"
          placeholder="Name"
          className="p-2 border rounded"
          required
        />
        <input
          type="email"
          placeholder="Email"
          className="p-2 border rounded"
          required
        />
        <input
          type="password"
          placeholder="Password"
          className="p-2 border rounded"
          required
        />
        <button type="submit" className="btn">Sign Up</button>
      </form>

      <p className="mt-4">
        Already have an account? <Link href="/auth/signin">Sign In</Link>
      </p>
    </div>
  );
}
