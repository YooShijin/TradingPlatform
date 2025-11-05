"use client";

import { useState, useEffect } from "react";
import { usePortfolio } from "@/lib/hooks";
import PerformanceMetrics from "@/components/PerformanceMetrics";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  AreaChart,
  Area,
} from "recharts";
import {
  TrendingUp,
  Activity,
  PieChart as PieChartIcon,
  BarChart3,
  Calendar,
} from "lucide-react";
import { motion } from "framer-motion";

export default function AnalyticsPage() {
  const userId = 1;
  const { portfolio } = usePortfolio(userId);
  const [timeRange, setTimeRange] = useState("7d");

  // Mock performance data (in real app, fetch from API)
  const performanceData = [
    { date: "Jan 10", value: 10000, profit: 0 },
    { date: "Jan 11", value: 10150, profit: 150 },
    { date: "Jan 12", value: 10080, profit: 80 },
    { date: "Jan 13", value: 10320, profit: 320 },
    { date: "Jan 14", value: 10500, profit: 500 },
    { date: "Jan 15", value: 10450, profit: 450 },
    { date: "Jan 16", value: 10680, profit: 680 },
  ];

  // Mock volume data
  const volumeData = [
    { stock: "AAPL", volume: 1500 },
    { stock: "GOOGL", volume: 800 },
    { stock: "MSFT", volume: 1200 },
    { stock: "TSLA", volume: 600 },
    { stock: "AMZN", volume: 400 },
  ];

  // Activity statistics
  const activityStats = [
    {
      icon: Activity,
      label: "Total Trades",
      value: "127",
      change: "+12%",
      positive: true,
    },
    {
      icon: TrendingUp,
      label: "Win Rate",
      value: "68.5%",
      change: "+5.2%",
      positive: true,
    },
    {
      icon: BarChart3,
      label: "Avg Trade Size",
      value: "$850",
      change: "-3%",
      positive: false,
    },
    {
      icon: Calendar,
      label: "Trading Days",
      value: "45",
      change: "Active",
      positive: true,
    },
  ];

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center space-y-4 sm:space-y-0">
        <div>
          <h1 className="text-3xl font-bold text-white mb-2">
            Performance Analytics
          </h1>
          <p className="text-gray-400">
            Detailed insights into your trading performance
          </p>
        </div>

        {/* Time Range Selector */}
        <div className="flex space-x-2">
          {["7d", "1m", "3m", "1y", "all"].map((range) => (
            <button
              key={range}
              onClick={() => setTimeRange(range)}
              className={`px-4 py-2 rounded-lg font-medium transition-all ${
                timeRange === range
                  ? "bg-primary text-white glow"
                  : "bg-dark-200 text-gray-400 hover:bg-dark-100"
              }`}
            >
              {range.toUpperCase()}
            </button>
          ))}
        </div>
      </div>

      {/* Activity Stats Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {activityStats.map((stat, idx) => {
          const Icon = stat.icon;
          return (
            <motion.div
              key={idx}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: idx * 0.1 }}
              className="glass rounded-xl p-6 card-hover"
            >
              <div className="flex items-start justify-between mb-4">
                <div
                  className={`w-12 h-12 rounded-lg flex items-center justify-center ${
                    stat.positive ? "bg-success/20" : "bg-danger/20"
                  }`}
                >
                  <Icon
                    className={`w-6 h-6 ${
                      stat.positive ? "text-success" : "text-danger"
                    }`}
                  />
                </div>
                <span
                  className={`text-sm font-semibold ${
                    stat.positive ? "text-success" : "text-danger"
                  }`}
                >
                  {stat.change}
                </span>
              </div>
              <p className="text-sm text-gray-400 mb-1">{stat.label}</p>
              <p className="text-2xl font-bold text-white">{stat.value}</p>
            </motion.div>
          );
        })}
      </div>

      {/* Performance Metrics */}
      <PerformanceMetrics portfolio={portfolio} />

      {/* Portfolio Value Over Time */}
      <div className="glass rounded-xl p-6">
        <h2 className="text-2xl font-bold text-primary mb-6 flex items-center">
          <TrendingUp className="w-6 h-6 mr-2" />
          Portfolio Value Over Time
        </h2>
        <ResponsiveContainer width="100%" height={350}>
          <AreaChart data={performanceData}>
            <defs>
              <linearGradient id="colorValue" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#00d4ff" stopOpacity={0.3} />
                <stop offset="95%" stopColor="#00d4ff" stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="#2a3f5f" />
            <XAxis
              dataKey="date"
              stroke="#8b95a5"
              style={{ fontSize: "12px" }}
            />
            <YAxis
              stroke="#8b95a5"
              style={{ fontSize: "12px" }}
              domain={["dataMin - 200", "dataMax + 200"]}
            />
            <Tooltip
              contentStyle={{
                backgroundColor: "#1a1f3a",
                border: "1px solid #2a3f5f",
                borderRadius: "8px",
                color: "#e0e6ed",
              }}
              formatter={(value) => ["$" + value.toFixed(2), "Portfolio Value"]}
            />
            <Area
              type="monotone"
              dataKey="value"
              stroke="#00d4ff"
              strokeWidth={3}
              fillOpacity={1}
              fill="url(#colorValue)"
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>

      {/* Trading Volume by Stock */}
      <div className="glass rounded-xl p-6">
        <h2 className="text-2xl font-bold text-primary mb-6 flex items-center">
          <BarChart3 className="w-6 h-6 mr-2" />
          Trading Volume by Stock
        </h2>
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={volumeData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#2a3f5f" />
            <XAxis
              dataKey="stock"
              stroke="#8b95a5"
              style={{ fontSize: "12px" }}
            />
            <YAxis stroke="#8b95a5" style={{ fontSize: "12px" }} />
            <Tooltip
              contentStyle={{
                backgroundColor: "#1a1f3a",
                border: "1px solid #2a3f5f",
                borderRadius: "8px",
                color: "#e0e6ed",
              }}
              formatter={(value) => [value + " shares", "Volume"]}
            />
            <Bar dataKey="volume" fill="#00d4ff" radius={[8, 8, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </div>

      {/* Profit/Loss Timeline */}
      <div className="glass rounded-xl p-6">
        <h2 className="text-2xl font-bold text-primary mb-6 flex items-center">
          <PieChartIcon className="w-6 h-6 mr-2" />
          Profit/Loss Timeline
        </h2>
        <ResponsiveContainer width="100%" height={300}>
          <AreaChart data={performanceData}>
            <defs>
              <linearGradient id="colorProfit" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#5cd85a" stopOpacity={0.3} />
                <stop offset="95%" stopColor="#5cd85a" stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="#2a3f5f" />
            <XAxis
              dataKey="date"
              stroke="#8b95a5"
              style={{ fontSize: "12px" }}
            />
            <YAxis stroke="#8b95a5" style={{ fontSize: "12px" }} />
            <Tooltip
              contentStyle={{
                backgroundColor: "#1a1f3a",
                border: "1px solid #2a3f5f",
                borderRadius: "8px",
                color: "#e0e6ed",
              }}
              formatter={(value) => ["$" + value.toFixed(2), "Profit/Loss"]}
            />
            <Area
              type="monotone"
              dataKey="profit"
              stroke="#5cd85a"
              strokeWidth={3}
              fillOpacity={1}
              fill="url(#colorProfit)"
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>

      {/* Trading Insights */}
      <div className="grid md:grid-cols-3 gap-6">
        <div className="glass rounded-xl p-6">
          <h3 className="text-lg font-semibold text-white mb-4">
            Best Performer
          </h3>
          <div className="bg-success/10 border border-success/30 rounded-lg p-4">
            <p className="text-success font-bold text-2xl mb-1">AAPL</p>
            <p className="text-gray-400 text-sm mb-2">Apple Inc.</p>
            <div className="flex items-center space-x-2">
              <TrendingUp className="w-5 h-5 text-success" />
              <span className="text-success font-semibold">+15.3%</span>
            </div>
          </div>
        </div>

        <div className="glass rounded-xl p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Most Traded</h3>
          <div className="bg-primary/10 border border-primary/30 rounded-lg p-4">
            <p className="text-primary font-bold text-2xl mb-1">MSFT</p>
            <p className="text-gray-400 text-sm mb-2">Microsoft Corp.</p>
            <div className="flex items-center space-x-2">
              <Activity className="w-5 h-5 text-primary" />
              <span className="text-primary font-semibold">1,200 shares</span>
            </div>
          </div>
        </div>

        <div className="glass rounded-xl p-6">
          <h3 className="text-lg font-semibold text-white mb-4">
            Average Hold Time
          </h3>
          <div className="bg-purple-500/10 border border-purple-500/30 rounded-lg p-4">
            <p className="text-purple-400 font-bold text-2xl mb-1">5.3 days</p>
            <p className="text-gray-400 text-sm mb-2">Per position</p>
            <div className="flex items-center space-x-2">
              <Calendar className="w-5 h-5 text-purple-400" />
              <span className="text-purple-400 font-semibold">Medium-term</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
