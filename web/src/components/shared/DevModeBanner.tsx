export default function DevModeBanner() {
  if (process.env.NEXT_PUBLIC_DEV_BYPASS_AUTH !== "true") return null;

  return (
    <div className="bg-yellow-100 border-b border-yellow-300 px-4 py-2 text-center text-sm text-yellow-800">
      Dev Mode: Logged in as dev@initium.local
    </div>
  );
}
