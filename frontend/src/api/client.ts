import axios from "axios";
import { useAuthStore } from "../store/authStore";

const API_URL =
  (import.meta.env.VITE_API_URL as string) || "http://localhost:8000";

const client = axios.create({
  baseURL: API_URL,
  headers: {
    "Content-Type": "application/json",
  },
});

// Helper function to redirect to login
function redirectToLogin(): void {
  if (
    typeof window !== "undefined" &&
    !window.location.pathname.startsWith("/auth")
  ) {
    window.location.href = "/auth/login";
  }
}

// Request interceptor - add auth token
client.interceptors.request.use(
  async (config) => {
    const { accessToken } = useAuthStore.getState();
    if (accessToken) {
      config.headers.Authorization = `Bearer ${accessToken}`;
    }

    return config;
  },
  (error) => {
    return Promise.reject(error);
  },
);

// Response interceptor - handle auth errors
client.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    // Handle 401 Unauthorized errors
    if (error.response?.status === 401) {
      const { clearAuth } = useAuthStore.getState();
      clearAuth();
      redirectToLogin();
    }

    console.error("API Error:", error);
    return Promise.reject(error);
  },
);

export default client;
