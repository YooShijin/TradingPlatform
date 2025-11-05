"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { tradingAPI } from "./api";

// Gorilla WebSocket compatible hook
export const useWebSocket = (url = "ws://localhost:8080/ws/trades") => {
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState(null);
  const wsRef = useRef(null);

  useEffect(() => {
    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log(" Connected to WebSocket:", url);
      setIsConnected(true);
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        setLastMessage(data);
      } catch (error) {
        console.error("Invalid WebSocket message:", event.data);
      }
    };

    ws.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    ws.onclose = () => {
      console.log(" WebSocket disconnected");
      setIsConnected(false);
    };

    return () => {
      ws.close();
    };
  }, [url]);

  return { isConnected, lastMessage };
};

// Order book hook with auto-refresh
export const useOrderBook = (stock, refreshInterval = 2000) => {
  const [orderBook, setOrderBook] = useState({ bids: [], asks: [] });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchOrderBook = useCallback(async () => {
    try {
      setLoading(true);
      const data = await tradingAPI.getOrderBook(stock);
      setOrderBook(data);
      setError(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [stock]);

  useEffect(() => {
    fetchOrderBook();
    const interval = setInterval(fetchOrderBook, refreshInterval);
    return () => clearInterval(interval);
  }, [fetchOrderBook, refreshInterval]);

  return { orderBook, loading, error, refresh: fetchOrderBook };
};

// Portfolio hook
export const usePortfolio = (userId, refreshInterval = 5000) => {
  const [portfolio, setPortfolio] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchPortfolio = useCallback(async () => {
    try {
      setLoading(true);
      const data = await tradingAPI.getPortfolio(userId);
      setPortfolio(data);
      setError(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [userId]);

  useEffect(() => {
    fetchPortfolio();
    const interval = setInterval(fetchPortfolio, refreshInterval);
    return () => clearInterval(interval);
  }, [fetchPortfolio, refreshInterval]);

  return { portfolio, loading, error, refresh: fetchPortfolio };
};

// Trades feed hook
export const useTradesFeed = (stock, limit = 20) => {
  const [trades, setTrades] = useState([]);
  const { lastMessage } = useWebSocket();

  useEffect(() => {
    const fetchTrades = async () => {
      try {
        const data = await tradingAPI.getTrades(stock, { limit });
        setTrades(data);
      } catch (err) {
        console.error("Failed to fetch trades:", err);
      }
    };

    fetchTrades();
  }, [stock, limit]);

  useEffect(() => {
    if (lastMessage && (!stock || lastMessage.stock === stock)) {
      setTrades((prev) => [lastMessage, ...prev].slice(0, limit));
    }
  }, [lastMessage, stock, limit]);

  return trades;
};

// Toast notification hook
export const useToast = () => {
  const [toasts, setToasts] = useState([]);

  const showToast = useCallback((message, type = "info", duration = 3000) => {
    const id = Date.now();
    setToasts((prev) => [...prev, { id, message, type }]);

    setTimeout(() => {
      setToasts((prev) => prev.filter((toast) => toast.id !== id));
    }, duration);
  }, []);

  return { toasts, showToast };
};

// Local storage hook
export const useLocalStorage = (key, initialValue) => {
  const [storedValue, setStoredValue] = useState(() => {
    if (typeof window === "undefined") return initialValue;

    try {
      const item = window.localStorage.getItem(key);
      return item ? JSON.parse(item) : initialValue;
    } catch (error) {
      console.error(error);
      return initialValue;
    }
  });

  const setValue = (value) => {
    try {
      const valueToStore =
        value instanceof Function ? value(storedValue) : value;
      setStoredValue(valueToStore);
      if (typeof window !== "undefined") {
        window.localStorage.setItem(key, JSON.stringify(valueToStore));
      }
    } catch (error) {
      console.error(error);
    }
  };

  return [storedValue, setValue];
};
