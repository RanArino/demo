import { ProfileSetupForm } from '@/features/user/ProfileSetupForm';
import { auth, currentUser } from '@clerk/nextjs/server';
import { redirect } from 'next/navigation';

export default async function ProfileSetupPage() {
  // Critical security check: Ensure user is authenticated
  const { userId } = await auth();
  
  if (!userId) {
    redirect('/sign-in');
  }


  // Get user email from Clerk
  const user = await currentUser();
  const userEmail = user?.primaryEmailAddress?.emailAddress;

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <ProfileSetupForm userEmail={userEmail} />
    </div>
  );
}
