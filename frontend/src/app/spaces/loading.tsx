export default function SpacesLoading() {
  return (
    <div className="flex h-screen bg-background">
      {/* Sidebar Skeleton */}
      <div className="w-[280px] border-r border-border p-6 space-y-6">
        <div className="h-10 bg-muted animate-pulse rounded-md" />
        <div className="space-y-4">
          <div className="h-8 bg-muted animate-pulse rounded-md w-3/4" />
          <div className="h-8 bg-muted animate-pulse rounded-md w-2/3" />
          <div className="h-8 bg-muted animate-pulse rounded-md w-4/5" />
        </div>
      </div>

      {/* Main Content Skeleton */}
      <div className="flex-1 p-8">
        {/* Header Skeleton */}
        <div className="flex items-center justify-between mb-8">
          <div className="h-8 bg-muted animate-pulse rounded-md w-48" />
          <div className="flex gap-4">
            <div className="h-10 bg-muted animate-pulse rounded-md w-32" />
            <div className="h-10 bg-muted animate-pulse rounded-md w-32" />
          </div>
        </div>

        {/* Grid Skeleton */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {Array.from({ length: 8 }).map((_, i) => (
            <SpaceCardSkeleton key={i} />
          ))}
        </div>
      </div>
    </div>
  );
}

function SpaceCardSkeleton() {
  return (
    <div className="bg-card border border-border rounded-lg p-6 space-y-4">
      {/* Icon/Thumbnail */}
      <div className="h-32 bg-muted animate-pulse rounded-md" />
      
      {/* Title */}
      <div className="h-6 bg-muted animate-pulse rounded-md w-3/4" />
      
      {/* Description */}
      <div className="space-y-2">
        <div className="h-4 bg-muted animate-pulse rounded-md" />
        <div className="h-4 bg-muted animate-pulse rounded-md w-5/6" />
      </div>
      
      {/* Metadata */}
      <div className="flex justify-between pt-2">
        <div className="h-4 bg-muted animate-pulse rounded-md w-20" />
        <div className="h-4 bg-muted animate-pulse rounded-md w-16" />
      </div>
    </div>
  );
}