import axios from 'axios';

// Create a configured axios instance
// Base URL points to the Go Backend (localhost:8080)
// Since we are running creating a separate frontend dev server, we'll need CORS handling on backend
// OR execute requests via Next.js API routes proxy. 
// For now, let's point directly to localhost:8080 and assume CORS is enabled or we run locally.
// If CORS is an issue, we'll need to enable it in Gin.

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add a request interceptor to attach JWT token
api.interceptors.request.use(
  (config) => {
    // Check if running in browser
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Add a response interceptor to handle 401s (optional logout)
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      if (typeof window !== 'undefined') {
        // localStorage.removeItem('token');
        // window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

export default api;
