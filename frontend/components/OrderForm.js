"use client";

import { useState } from "react";
import { tradingAPI } from "@/lib/api";
import { Loader2, TrendingUp, TrendingDown } from "lucide-react";

export default function OrderForm({ onOrderPlaced, showToast }) {
  const [formData, setFormData] = useState({
    stock: "AAPL",
    type: "BUY",
    price: "150.00",
    quantity: "10",
  });
  const [loading, setLoading] = useState(false);

  const stocks = [
    { value: "AAPL", label: "AAPL - Apple" },
    { value: "GOOGL", label: "GOOGL - Google" },
    { value: "MSFT", label: "MSFT - Microsoft" },
    { value: "TSLA", label: "TSLA - Tesla" },
    { value: "AMZN", label: "AMZN - Amazon" },
  ];

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);

    try {
      const order = {
        user_id: 1,
        stock: formData.stock,
        type: formData.type,
        price: parseFloat(formData.price),
        quantity: parseInt(formData.quantity),
      };

      const result = await tradingAPI.placeOrder(order);

      if (result.trades && result.trades.length > 0) {
        showToast(
          `Order matched! Executed ${result.trades.length} trade(s)`,
          "success"
        );
      } else {
        showToast("Order placed successfully", "success");
      }

      if (onOrderPlaced) onOrderPlaced(result);
    } catch (error) {
      showToast("Failed to place order: " + error.message, "error");
    } finally {
      setLoading(false);
    }
  };

  const autoGenerateOrders = async (count) => {
    setLoading(true);
    try {
      for (let i = 0; i < count; i++) {
        const randomOrder = {
          user_id: Math.floor(Math.random() * 10) + 1,
          stock: stocks[Math.floor(Math.random() * stocks.length)].value,
          type: Math.random() > 0.5 ? "BUY" : "SELL",
          price: 145 + Math.random() * 20,
          quantity: Math.floor(Math.random() * 50) + 10,
        };

        await tradingAPI.placeOrder(randomOrder);
        await new Promise((resolve) => setTimeout(resolve, 100));
      }

      showToast(`Generated ${count} random orders`, "success");
      if (onOrderPlaced) onOrderPlaced();
    } catch (error) {
      showToast("Failed to generate orders: " + error.message, "error");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="glass rounded-xl p-6 card-hover">
      <h2 className="text-2xl font-bold text-primary mb-6 flex items-center">
        <TrendingUp className="w-6 h-6 mr-2" />
        Place Order
      </h2>

      <form onSubmit={handleSubmit} className="space-y-4">
        {/* Stock Selection */}
        <div>
          <label className="block text-sm font-medium text-gray-400 mb-2">
            Stock Symbol
          </label>
          <select
            value={formData.stock}
            onChange={(e) =>
              setFormData({ ...formData, stock: e.target.value })
            }
            className="w-full bg-dark-200 border border-gray-700 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-primary transition-colors"
            required
          >
            {stocks.map((stock) => (
              <option key={stock.value} value={stock.value}>
                {stock.label}
              </option>
            ))}
          </select>
        </div>

        {/* Order Type */}
        <div>
          <label className="block text-sm font-medium text-gray-400 mb-2">
            Order Type
          </label>
          <div className="grid grid-cols-2 gap-3">
            <button
              type="button"
              onClick={() => setFormData({ ...formData, type: "BUY" })}
              className={`py-3 rounded-lg font-semibold transition-all flex items-center justify-center space-x-2 ${
                formData.type === "BUY"
                  ? "bg-success text-white"
                  : "bg-dark-200 text-gray-400 hover:bg-dark-100"
              }`}
            >
              <TrendingUp className="w-5 h-5" />
              <span>Buy</span>
            </button>
            <button
              type="button"
              onClick={() => setFormData({ ...formData, type: "SELL" })}
              className={`py-3 rounded-lg font-semibold transition-all flex items-center justify-center space-x-2 ${
                formData.type === "SELL"
                  ? "bg-danger text-white"
                  : "bg-dark-200 text-gray-400 hover:bg-dark-100"
              }`}
            >
              <TrendingDown className="w-5 h-5" />
              <span>Sell</span>
            </button>
          </div>
        </div>

        {/* Price */}
        <div>
          <label className="block text-sm font-medium text-gray-400 mb-2">
            Price ($)
          </label>
          <input
            type="number"
            step="0.01"
            min="0.01"
            value={formData.price}
            onChange={(e) =>
              setFormData({ ...formData, price: e.target.value })
            }
            className="w-full bg-dark-200 border border-gray-700 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-primary transition-colors"
            required
          />
        </div>

        {/* Quantity */}
        <div>
          <label className="block text-sm font-medium text-gray-400 mb-2">
            Quantity
          </label>
          <input
            type="number"
            min="1"
            value={formData.quantity}
            onChange={(e) =>
              setFormData({ ...formData, quantity: e.target.value })
            }
            className="w-full bg-dark-200 border border-gray-700 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-primary transition-colors"
            required
          />
        </div>

        {/* Submit Button */}
        <button
          type="submit"
          disabled={loading}
          className={`w-full py-3 rounded-lg font-semibold transition-all ${
            formData.type === "BUY"
              ? "bg-gradient-to-r from-success to-green-600 hover:glow"
              : "bg-gradient-to-r from-danger to-red-600 hover:glow"
          } text-white disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center space-x-2`}
        >
          {loading ? (
            <>
              <Loader2 className="w-5 h-5 animate-spin" />
              <span>Processing...</span>
            </>
          ) : (
            <span>Place {formData.type} Order</span>
          )}
        </button>
      </form>

      {/* Quick Fill Section */}
      <div className="mt-6 pt-6 border-t border-gray-700/50">
        <h3 className="text-sm font-medium text-gray-400 mb-3">
          Quick Fill for Testing
        </h3>
        <button
          onClick={() => autoGenerateOrders(10)}
          disabled={loading}
          className="w-full bg-dark-200 hover:bg-dark-100 border border-gray-700 text-gray-300 py-3 rounded-lg font-medium transition-all disabled:opacity-50"
        >
          Generate 10 Random Orders
        </button>
      </div>
    </div>
  );
}
