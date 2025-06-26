// frontend/lib/api.ts (Updated)
import { auth } from "./firebase";

// IMPORTANT: Make sure this points to your Go backend's URL
const API_BASE_URL = "http://localhost:8080"; 

export const authenticatedFetch = async (endpoint: string, options: RequestInit = {}) => {
  const user = auth.currentUser;
  if (!user) {
    throw new Error("User is not authenticated");
  }
  const token = await user.getIdToken();

  const headers = new Headers(options.headers || {});
  headers.append("Authorization", `Bearer ${token}`);

  // The browser automatically sets the correct Content-Type with boundary for FormData.
  // Manually setting it to 'application/json' for FormData will break the request.
  if (!(options.body instanceof FormData)) {
    headers.append("Content-Type", "application/json");
  }

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    // Try to parse error as text, as backend might not send JSON on error
    const errorText = await response.text();
    throw new Error(errorText || response.statusText || "An unknown error occurred");
  }

  // Handle responses that might not have a body or are not JSON
  const contentType = response.headers.get("content-type");
  if (contentType && contentType.indexOf("application/json") !== -1) {
    return response.json();
  }
  return response.text();
};