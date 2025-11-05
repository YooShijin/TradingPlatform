"use client";

import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  Legend,
  Tooltip,
} from "recharts";
import { TrendingUp, Activity, Target, DollarSign } from "lucide-react";

export default function PerformanceMetrics({ portfolio }) {
  if (!portfolio || !portfolio.holdings || portfolio.holdings.length === 0) {
    return (
      <div className="glass rounded-xl p-6">
        <p className="text-center text-gray-500">No data available</p>
      </div>
    );
  }

  // Portfolio allocation data
  const allocationData = portfolio.holdings.map((h) => ({
    name: h.stock,
    value: h.quantity * h.current_price,
  }));

  const COLORS = [
    "#00d4ff",
    "#5cd85a",
    "#ff4757",
    "#ffa502",
    "#a29bfe",
    "#fd79a8",
  ];

  // Calculate metrics
  const totalInvested = portfolio.holdings.reduce(
    (sum, h) => sum + h.quantity * h.avg_buy_price,
    0
  );
  const totalCurrent = portfolio.holdings.reduce(
    (sum, h) => sum + h.quantity * h.current_price,
    0
  );
  const totalPL = totalCurrent - totalInvested;
  const plPercentage = totalInvested > 0 ? (totalPL / totalInvested) * 100 : 0;

  const metrics = [
    {
      icon: DollarSign,
      label: "Total Invested",
      value: `$${totalInvested.toFixed(2)}`,
      color: "text-primary",
    },
    {
      icon: TrendingUp,
      label: "Current Value",
      value: `$${totalCurrent.toFixed(2)}`,
      color: "text-success",
    },
    {
      icon: Activity,
      label: "Total Return",
      value: `${totalPL >= 0 ? "+" : ""}$${totalPL.toFixed(2)}`,
      color: totalPL >= 0 ? "text-success" : "text-danger",
    },
    {
      icon: Target,
      label: "Return %",
      value: `${plPercentage >= 0 ? "+" : ""}${plPercentage.toFixed(2)}%`,
      color: plPercentage >= 0 ? "text-success" : "text-danger",
    },
  ];

  return (
    <div className="space-y-6">
      {/* Metrics Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {metrics.map((metric, idx) => {
          const Icon = metric.icon;
          return (
            <div key={idx} className="glass rounded-xl p-4 card-hover">
              <div className="flex items-center space-x-3">
                <div
                  className={`w-10 h-10 ${metric.color} bg-opacity-20 rounded-lg flex items-center justify-center`}
                >
                  <Icon className={`w-5 h-5 ${metric.color}`} />
                </div>
                <div>
                  <p className="text-xs text-gray-400">{metric.label}</p>
                  <p className={`text-lg font-bold ${metric.color}`}>
                    {metric.value}
                  </p>
                </div>
              </div>
            </div>
          );
        })}
      </div>{" "}
      {/* Portfolio Allocation Chart */}
      <div className="glass rounded-xl p-6">
        <h3 className="text-xl font-bold text-primary mb-6">
          Portfolio Allocation
        </h3>
        <div className="grid md:grid-cols-2 gap-8">
          <ResponsiveContainer width="100%" height={250}>
            <PieChart>
              <Pie
                data={allocationData}
                cx="50%"
                cy="50%"
                labelLine={false}
                label={({ name, percent }) =>
                  `${name}: ${(percent * 100).toFixed(0)}%`
                }
                outerRadius={80}
                fill="#8884d8"
                dataKey="value"
              >
                {allocationData.map((entry, index) => (
                  <Cell
                    key={`cell-${index}`}
                    fill={COLORS[index % COLORS.length]}
                  />
                ))}
              </Pie>
              <Tooltip
                contentStyle={{
                  backgroundColor: "#1a1f3a",
                  border: "1px solid #2a3f5f",
                  borderRadius: "8px",
                  color: "#e0e6ed",
                }}
                formatter={(value) => "$" + value.toFixed(2)}
              />
            </PieChart>
          </ResponsiveContainer>

          {/* Holdings List */}
          <div className="space-y-3">
            {portfolio.holdings.map((holding, idx) => (
              <div
                key={holding.stock}
                className="flex items-center justify-between"
              >
                <div className="flex items-center space-x-3">
                  <div
                    className="w-4 h-4 rounded-full"
                    style={{ backgroundColor: COLORS[idx % COLORS.length] }}
                  ></div>
                  <div>
                    <p className="font-semibold text-white">{holding.stock}</p>
                    <p className="text-xs text-gray-400">
                      {holding.quantity} shares
                    </p>
                  </div>
                </div>
                <div className="text-right">
                  <p className="font-semibold text-white">
                    ${(holding.quantity * holding.current_price).toFixed(2)}
                  </p>
                  <p
                    className={`text-xs ${
                      holding.profit_loss >= 0 ? "text-success" : "text-danger"
                    }`}
                  >
                    {holding.profit_loss >= 0 ? "+" : ""}$
                    {holding.profit_loss.toFixed(2)}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
