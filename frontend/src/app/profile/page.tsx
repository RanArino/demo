import { ProfileView } from '@/features/user/ProfileView';
import { auth } from '@clerk/nextjs/server';
import { redirect } from 'next/navigation';
import { checkUserStatus } from '@/api/actions/userActions';

export default async function ProfilePage() {
  // Critical security check: Ensure user is authenticated
  const { userId } = await auth();
  
  if (!userId) {
    redirect('/sign-in');
  }

  // Check user status and redirect if profile incomplete
  const statusResult = await checkUserStatus();
  if (statusResult.success && !statusResult.profileCompleted) {
    redirect('/profile/setup');
  }

  return (
    <div className="min-h-screen bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-4xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">My Profile</h1>
          <p className="mt-2 text-gray-600">
            Manage your account information and preferences
          </p>
        </div>
        
        <ProfileView />
      </div>
    </div>
  );
}