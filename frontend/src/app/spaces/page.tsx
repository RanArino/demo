export default function SpacesPage() {
  const spaces = [
    { id: 1, name: 'Knowledge Base', description: 'A space for all your documents and notes.' },
    { id: 2, name: 'Project Alpha', description: 'Collaboration space for Project Alpha.' },
    { id: 3, name: 'Personal Journal', description: 'Your private space for thoughts and ideas.' },
    { id: 4, name: 'Team Meetings', description: 'Archive of all team meeting recordings and notes.' },
  ];

  return (
    <div className="min-h-screen bg-gray-50 text-gray-800">
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-4xl font-bold text-gray-900 mb-8">
          My Spaces
        </h1>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {spaces.map((space) => (
            <div key={space.id} className="bg-white p-6 rounded-lg shadow-md hover:shadow-lg transition-shadow duration-300">
              <h2 className="text-2xl font-semibold text-gray-800 mb-2">{space.name}</h2>
              <p className="text-gray-600">{space.description}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
