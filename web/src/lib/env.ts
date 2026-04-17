const raw = process.env.API_URL;

if (!raw && process.env.NODE_ENV === "production") {
  throw new Error("API_URL environment variable is required in production");
}

export const API_URL = raw ?? "http://localhost:8000";
