export function SpacesSkelton() {
  return (
    <div className="flex h-screen">
      <aside className="w-full md:w-1/4 bg-gray-100 p-4">
        <h2 className="text-lg font-semibold mb-4">Filters</h2>
        <div className="mb-4">
          <h3 className="font-medium mb-2">Keywords</h3>
          <div className="space-y-2">
            <div className="h-4 bg-gray-200 rounded w-2/3"></div>
            <div className="h-4 bg-gray-200 rounded w-1/2"></div>
            <div className="h-4 bg-gray-200 rounded w-3/4"></div>
          </div>
        </div>
      </aside>
      <main className="flex-1 p-4 w-full md:w-3/4">
        <nav className="mb-4 flex justify-between items-center">
          <h1 className="text-2xl font-bold">Spaces</h1>
          <div className="h-10 bg-gray-200 rounded w-1/3"></div>
          <div className="flex space-x-2">
            <div className="h-10 bg-gray-200 rounded w-24"></div>
            <div className="h-10 bg-gray-200 rounded w-24"></div>
            <div className="h-10 bg-gray-200 rounded w-24"></div>
          </div>
          <div className="h-10 bg-gray-200 rounded w-32"></div>
        </nav>
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          <div className="h-48 bg-gray-200 rounded"></div>
          <div className="h-48 bg-gray-200 rounded"></div>
          <div className="h-48 bg-gray-200 rounded"></div>
          <div className="h-48 bg-gray-200 rounded"></div>
        </div>
      </main>
    </div>
  );
}
