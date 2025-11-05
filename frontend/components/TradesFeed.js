"use client";

import { useTradesFeed } from "@/lib/hooks";
import { Activity } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";

export default function TradesFeed({ stock }) {
  const trades = useTradesFeed(stock, 20);

  return (
    <div className="glass rounded-xl p-6">
      <h2 className="text-2xl font-bold text-primary mb-6 flex items-center">
        <Activity className="w-6 h-6 mr-2" />
        Recent Trades
      </h2>

      <div className="space-y-2 max-h-[500px] overflow-y-auto pr-2">
        <AnimatePresence initial={false}>
          {trades && trades.length > 0 ? (
            trades.map((trade, idx) => (
              <motion.div
                key={trade.id || idx}
                initial={{ opacity: 0, y: -20 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, x: -100 }}
                transition={{ duration: 0.3 }}
                className="bg-dark-200 rounded-lg p-4 border-l-4 border-primary hover:bg-dark-100 transition-colors"
              >
                <div className="flex justify-between items-start mb-2">
                  <span className="font-semibold text-white">
                    {trade.stock}
                  </span>
                  <span className="text-xs text-gray-400">
                    {new Date(trade.executed_at).toLocaleTimeString()}
                  </span>
                </div>
                <div className="flex justify-between items-center">
                  <div>
                    <span className="text-gray-400 text-sm">
                      {trade.quantity} shares @
                    </span>
                    <span className="ml-2 font-mono text-primary font-semibold">
                      ${trade.price.toFixed(2)}
                    </span>
                  </div>
                  <div className="text-right">
                    <p className="text-xs text-gray-400">Total Value</p>
                    <p className="font-semibold text-white">
                      ${(trade.price * trade.quantity).toFixed(2)}
                    </p>
                  </div>
                </div>
              </motion.div>
            ))
          ) : (
            <div className="text-center py-12 text-gray-500">
              <Activity className="w-12 h-12 mx-auto mb-4 opacity-50" />
              <p>No recent trades</p>
            </div>
          )}
        </AnimatePresence>
      </div>
    </div>
  );
}
