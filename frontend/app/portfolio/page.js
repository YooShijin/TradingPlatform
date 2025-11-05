"use client";

import { useState } from "react";
import PortfolioSummary from "@/components/PortfolioSummary";
import { usePortfolio } from "@/lib/hooks";
import {
  ArrowUpRight,
  ArrowDownRight,
  Calendar,
  Download,
  Filter,
} from "lucide-react";
import { motion } from "framer-motion";

export default function PortfolioPage() {
  const userId = 1;
  const { portfolio } = usePortfolio(userId);
  const [filterStock, setFilterStock] = useState("all");

  // Mock trade history (in real app, fetch from API)
  const tradeHistory = [
    {
      id: 1,
      date: "2024-01-15",
      stock: "AAPL",
      type: "BUY",
      quantity: 50,
      price: 150.0,
      total: 7500.0,
    },
    {
      id: 2,
      date: "2024-01-16",
      stock: "GOOGL",
      type: "BUY",
      quantity: 25,
      price: 140.0,
      total: 3500.0,
    },
    {
      id: 3,
      date: "2024-01-17",
      stock: "TSLA",
      type: "SELL",
      quantity: 10,
      price: 200.0,
      total: 2000.0,
    },
    {
      id: 4,
      date: "2024-01-18",
      stock: "MSFT",
      type: "BUY",
      quantity: 30,
      price: 300.0,
      total: 9000.0,
    },
    {
      id: 5,
      date: "2024-01-19",
      stock: "AAPL",
      type: "SELL",
      quantity: 20,
      price: 155.0,
      total: 3100.0,
    },
  ];

  const filteredHistory =
    filterStock === "all"
      ? tradeHistory
      : tradeHistory.filter((t) => t.stock === filterStock);

  const uniqueStocks = ["all", ...new Set(tradeHistory.map((t) => t.stock))];

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-white mb-2">
            Portfolio Overview
          </h1>
          <p className="text-gray-400">
            Track your investments and performance
          </p>
        </div>
        <button className="flex items-center space-x-2 bg-primary hover:bg-secondary text-white px-6 py-3 rounded-lg font-semibold transition-all hover:glow">
          <Download className="w-5 h-5" />
          <span>Export Report</span>
        </button>
      </div>

      {/* Portfolio Summary */}
      <PortfolioSummary userId={userId} detailed={true} />

      {/* Trade History Section */}
      <div className="glass rounded-xl p-6">
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center mb-6 space-y-4 sm:space-y-0">
          <h2 className="text-2xl font-bold text-primary flex items-center">
            <Calendar className="w-6 h-6 mr-2" />
            Trade History
          </h2>

          {/* Filter */}
          <div className="flex items-center space-x-3">
            <Filter className="w-5 h-5 text-gray-400" />
            <select
              value={filterStock}
              onChange={(e) => setFilterStock(e.target.value)}
              className="bg-dark-200 border border-gray-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-primary"
            >
              {uniqueStocks.map((stock) => (
                <option key={stock} value={stock}>
                  {stock === "all" ? "All Stocks" : stock}
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* Trade History Table */}
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-700">
                <th className="text-left py-3 px-4 text-gray-400 font-medium">
                  Date
                </th>
                <th className="text-left py-3 px-4 text-gray-400 font-medium">
                  Stock
                </th>
                <th className="text-left py-3 px-4 text-gray-400 font-medium">
                  Type
                </th>
                <th className="text-right py-3 px-4 text-gray-400 font-medium">
                  Quantity
                </th>
                <th className="text-right py-3 px-4 text-gray-400 font-medium">
                  Price
                </th>
                <th className="text-right py-3 px-4 text-gray-400 font-medium">
                  Total
                </th>
              </tr>
            </thead>
            <tbody>
              {filteredHistory.map((trade, idx) => (
                <motion.tr
                  key={trade.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: idx * 0.05 }}
                  className="border-b border-gray-800 hover:bg-dark-100 transition-colors"
                >
                  <td className="py-4 px-4 text-gray-300">{trade.date}</td>
                  <td className="py-4 px-4">
                    <span className="font-semibold text-white">
                      {trade.stock}
                    </span>
                  </td>
                  <td className="py-4 px-4">
                    <span
                      className={`inline-flex items-center space-x-1 px-3 py-1 rounded-full text-sm font-medium ${
                        trade.type === "BUY"
                          ? "bg-success/20 text-success"
                          : "bg-danger/20 text-danger"
                      }`}
                    >
                      {trade.type === "BUY" ? (
                        <ArrowUpRight className="w-4 h-4" />
                      ) : (
                        <ArrowDownRight className="w-4 h-4" />
                      )}
                      <span>{trade.type}</span>
                    </span>
                  </td>
                  <td className="py-4 px-4 text-right text-gray-300">
                    {trade.quantity}
                  </td>
                  <td className="py-4 px-4 text-right font-mono text-primary">
                    ${trade.price.toFixed(2)}
                  </td>
                  <td className="py-4 px-4 text-right font-semibold text-white">
                    ${trade.total.toFixed(2)}
                  </td>
                </motion.tr>
              ))}
            </tbody>
          </table>
        </div>

        {filteredHistory.length === 0 && (
          <div className="text-center py-12 text-gray-500">
            <Calendar className="w-12 h-12 mx-auto mb-4 opacity-50" />
            <p>No trade history found</p>
          </div>
        )}
      </div>
    </div>
  );
}
