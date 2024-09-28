// pages/account.js

import { getSession } from 'next-auth/react'

export default function Account({ session }) {
  if (!session) {
    return <p>Please sign in to access your account</p>
  }

  return (
    <div className="min-h-screen flex items-center justify-center">
      <div>
        <h1 className="text-xl font-semibold mb-4">Account Details</h1>
        <p>Email: {session.user.email}</p>
      </div>
    </div>
  )
}

export async function getServerSideProps(context) {
  const session = await getSession(context)
  return {
    props: { session },
  }
}
