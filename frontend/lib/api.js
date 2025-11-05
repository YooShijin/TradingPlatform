import axios from "axios";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api";

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    "Content-Type": "application/json",
  },
});

// Request interceptor
api.interceptors.request.use(
  (config) => {
    // Add auth token if needed
    // const token = localStorage.getItem('token');
    // if (token) config.headers.Authorization = `Bearer ${token}`;
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor
api.interceptors.response.use(
  (response) => response.data,
  (error) => {
    console.error("API Error:", error.response?.data || error.message);
    return Promise.reject(error);
  }
);

// API methods
export const tradingAPI = {
  // Orders
  placeOrder: (orderData) => api.post("/order/place", orderData),
  cancelOrder: (orderId) => api.delete(`/order/${orderId}`),
  getUserOrders: (userId) => api.get(`/orders/user/${userId}`),

  // Market Data
  getOrderBook: (stock) => api.get(`/orderbook/${stock}`),
  getTrades: (stock, params) => api.get(`/trades/${stock}`, { params }),
  getPrice: (stock) => api.get(`/price/${stock}`),

  // Portfolio
  getPortfolio: (userId) => api.get(`/portfolio/${userId}`),
  getBalance: (userId) => api.get(`/balance/${userId}`),

  // Analytics
  getPerformanceMetrics: (userId) =>
    api.get(`/analytics/performance/${userId}`),
  getTradeHistory: (userId, params) =>
    api.get(`/analytics/trades/${userId}`, { params }),
};

export default api;
