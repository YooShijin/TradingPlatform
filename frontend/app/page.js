"use client";

import { useState } from "react";
import OrderForm from "@/components/OrderForm";
import OrderBook from "@/components/OrderBook";
import TradesFeed from "@/components/TradesFeed";
import StockChart from "@/components/StockChart";
import { useToast } from "@/lib/hooks";
import { X } from "lucide-react";

export default function TradingPage() {
  const [selectedStock, setSelectedStock] = useState("AAPL");
  const { toasts, showToast } = useToast();

  const stocks = ["AAPL", "GOOGL", "MSFT", "TSLA", "AMZN"];

  const handleOrderPlaced = () => {
    // Trigger refresh of order book
  };

  return (
    <div className="space-y-6">
      {/* Toast Notifications */}
      <div className="fixed top-20 right-4 z-50 space-y-2">
        {toasts.map((toast) => (
          <div
            key={toast.id}
            className={`toast glass rounded-lg p-4 shadow-lg flex items-center space-x-3 min-w-[300px] ${
              toast.type === "success"
                ? "border-l-4 border-success"
                : toast.type === "error"
                ? "border-l-4 border-danger"
                : "border-l-4 border-primary"
            }`}
          >
            <p className="flex-1 text-white">{toast.message}</p>
            <button className="text-gray-400 hover:text-white">
              <X className="w-4 h-4" />
            </button>
          </div>
        ))}
      </div>

      {/* Stock Selector */}
      <div className="glass rounded-xl p-4">
        <div className="flex items-center space-x-3 overflow-x-auto">
          <span className="text-gray-400 font-medium whitespace-nowrap">
            Select Stock:
          </span>
          {stocks.map((stock) => (
            <button
              key={stock}
              onClick={() => setSelectedStock(stock)}
              className={`px-6 py-2 rounded-lg font-semibold transition-all whitespace-nowrap ${
                selectedStock === stock
                  ? "bg-primary text-white glow"
                  : "bg-dark-200 text-gray-400 hover:bg-dark-100"
              }`}
            >
              {stock}
            </button>
          ))}
        </div>
      </div>

      {/* Price Chart */}
      <StockChart stock={selectedStock} />

      {/* Main Grid */}
      <div className="grid lg:grid-cols-3 gap-6">
        {/* Order Form */}
        <div className="lg:col-span-1">
          <OrderForm onOrderPlaced={handleOrderPlaced} showToast={showToast} />
        </div>

        {/* Order Book */}
        <div className="lg:col-span-2">
          <OrderBook stock={selectedStock} />
        </div>
      </div>

      {/* Trades Feed */}
      <TradesFeed stock={selectedStock} />
    </div>
  );
}
