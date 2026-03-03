import axios from 'axios';

const API_BASE = process.env.REACT_APP_API_URL || '';

const api = axios.create({
  baseURL: API_BASE,
  headers: { 'Content-Type': 'application/json' },
});

// Attach JWT token to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Redirect on 401
api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(err);
  }
);

export const login = (username, password) =>
  api.post('/api/v1/auth/login', { username, password });

export const getNodes = () => api.get('/api/v1/nodes');

export const getMetrics = (nodeId, from, to, limit = 100) =>
  api.get('/api/v1/metrics', { params: { node_id: nodeId, from, to, limit } });

export const getAlerts = (nodeId, limit = 50) =>
  api.get('/api/v1/alerts', { params: { node_id: nodeId, limit } });

export default api;
