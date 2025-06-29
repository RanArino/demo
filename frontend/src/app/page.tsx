export default function Home() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="text-center">
        <h1 className="text-4xl font-bold text-gray-900 mb-4">
          Knowledge Exploration Platform
        </h1>
        <p className="text-gray-600 mb-8">
          Navigate to <a href="/spaces" className="text-blue-600 hover:underline">/spaces</a> to view the spaces page
        </p>
        <a 
          href="/spaces" 
          className="bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700 transition-colors"
        >
          Go to Spaces
        </a>
      </div>
    </div>
  );
}
