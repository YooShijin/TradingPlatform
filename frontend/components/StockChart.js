"use client";

import { useState, useEffect } from "react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { tradingAPI } from "@/lib/api";

export default function StockChart({ stock }) {
  const [trades, setTrades] = useState([]);
  const [data, setData] = useState([]);
  const [timeframe, setTimeframe] = useState("1h");

  useEffect(() => {
    const fetchTradeData = async () => {
      try {
        const fetched = await tradingAPI.getTrades(stock, {
          hours: timeframe === "1h" ? 1 : 24,
        });
        const tradesData = Array.isArray(fetched) ? fetched : [];

        setTrades(tradesData); // ✅ store fetched trades

        // Aggregate trades by 5-minute intervals
        const aggregated = tradesData.reduce((acc, trade) => {
          const time = new Date(trade.executed_at);
          const interval =
            Math.floor(time.getTime() / (5 * 60 * 1000)) * (5 * 60 * 1000);
          const key = new Date(interval).toLocaleTimeString();

          if (!acc[key]) {
            acc[key] = { time: key, price: trade.price, volume: 0 };
          }
          acc[key].volume += trade.quantity;

          return acc;
        }, {});

        setData(Object.values(aggregated).slice(-20));
      } catch (error) {
        console.error("Failed to fetch chart data:", error);
        setData([]); // safety fallback
      }
    };

    fetchTradeData();
    const interval = setInterval(fetchTradeData, 10000);

    return () => clearInterval(interval);
  }, [stock, timeframe]);

  return (
    <div className="glass rounded-xl p-6">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-primary">
          Price Chart - {stock}
        </h2>
        <div className="flex space-x-2">
          <button
            onClick={() => setTimeframe("1h")}
            className={`px-4 py-2 rounded-lg font-medium transition-colors ${
              timeframe === "1h"
                ? "bg-primary text-white"
                : "bg-dark-200 text-gray-400 hover:bg-dark-100"
            }`}
          >
            1H
          </button>
          <button
            onClick={() => setTimeframe("24h")}
            className={`px-4 py-2 rounded-lg font-medium transition-colors ${
              timeframe === "24h"
                ? "bg-primary text-white"
                : "bg-dark-200 text-gray-400 hover:bg-dark-100"
            }`}
          >
            24H
          </button>
        </div>
      </div>

      <ResponsiveContainer width="100%" height={300}>
        <LineChart data={data}>
          <CartesianGrid strokeDasharray="3 3" stroke="#2a3f5f" />
          <XAxis dataKey="time" stroke="#8b95a5" style={{ fontSize: "12px" }} />
          <YAxis
            stroke="#8b95a5"
            style={{ fontSize: "12px" }}
            domain={["dataMin - 5", "dataMax + 5"]}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: "#1a1f3a",
              border: "1px solid #2a3f5f",
              borderRadius: "8px",
              color: "#e0e6ed",
            }}
            formatter={(value) => ["$" + value.toFixed(2), "Price"]}
          />
          <Line
            type="monotone"
            dataKey="price"
            stroke="#00d4ff"
            strokeWidth={2}
            dot={false}
            activeDot={{ r: 6, fill: "#00d4ff" }}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
