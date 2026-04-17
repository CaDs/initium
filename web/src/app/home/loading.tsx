export default function HomeLoading() {
  return (
    <div
      className="max-w-2xl mx-auto px-6 py-12 animate-pulse"
      aria-label="Loading"
      aria-busy="true"
    >
      {/* Title skeleton */}
      <div className="h-9 bg-muted/30 rounded-md w-64 mb-2" />
      {/* Subtitle skeleton */}
      <div className="h-5 bg-muted/20 rounded-md w-80 mb-8" />

      {/* Profile card skeleton */}
      <div className="bg-card border border-border rounded-lg p-6 space-y-3">
        <div className="h-6 bg-muted/30 rounded-md w-24 mb-3" />
        <div className="grid grid-cols-[100px_1fr] gap-2">
          <div className="h-4 bg-muted/20 rounded w-12" />
          <div className="h-4 bg-muted/20 rounded w-40" />
          <div className="h-4 bg-muted/20 rounded w-12" />
          <div className="h-4 bg-muted/20 rounded w-28" />
          <div className="h-4 bg-muted/20 rounded w-16" />
          <div className="h-4 bg-muted/20 rounded w-56" />
        </div>
      </div>
    </div>
  );
}
