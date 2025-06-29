
export default async function UsersModal({ params }: { params: Promise<{ spaceId: string }> }) {
  const { spaceId } = await params;
  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
      <div className="bg-white p-8 rounded-lg">
        <h2 className="text-xl font-bold">Users for Space {spaceId}</h2>
        <p>This is a placeholder for the users modal.</p>
        {/* You would fetch and display user data here */}
      </div>
    </div>
  );
}
