"use client";

import { usePortfolio } from "@/lib/hooks";
import { Wallet, TrendingUp, TrendingDown, RefreshCw } from "lucide-react";
import { motion } from "framer-motion";

export default function PortfolioSummary({ userId, detailed = false }) {
  const { portfolio, loading, refresh } = usePortfolio(userId);

  if (loading && !portfolio) {
    return (
      <div className="glass rounded-xl p-6 animate-pulse">
        <div className="h-32 bg-dark-200 rounded"></div>
      </div>
    );
  }

  if (!portfolio) return null;

  const totalPL =
    portfolio.holdings?.reduce((sum, h) => sum + h.profit_loss, 0) || 0;
  const totalValue = (portfolio.balance || 0) + (portfolio.total_value || 0);

  return (
    <div className="space-y-6">
      {/* Balance Card */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="glass rounded-xl p-6 bg-gradient-to-br from-purple-500/10 to-pink-500/10 border border-purple-500/20"
      >
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center space-x-3">
            <div className="w-12 h-12 bg-gradient-to-br from-purple-500 to-pink-500 rounded-lg flex items-center justify-center">
              <Wallet className="w-6 h-6 text-white" />
            </div>
            <div>
              <p className="text-sm text-gray-400">Account Balance</p>
              <h3 className="text-3xl font-bold text-white">
                ${portfolio.balance?.toFixed(2) || "0.00"}
              </h3>
            </div>
          </div>
          <button
            onClick={refresh}
            className="p-2 hover:bg-white/10 rounded-lg transition-colors"
          >
            <RefreshCw className="w-5 h-5 text-gray-400" />
          </button>
        </div>

        <div className="grid grid-cols-2 gap-4 mt-6 pt-6 border-t border-white/10">
          <div>
            <p className="text-xs text-gray-400 mb-1">Total Portfolio Value</p>
            <p className="text-xl font-semibold text-white">
              ${totalValue.toFixed(2)}
            </p>
          </div>
          <div>
            <p className="text-xs text-gray-400 mb-1">Total P/L</p>
            <div className="flex items-center space-x-2">
              {totalPL >= 0 ? (
                <TrendingUp className="w-5 h-5 text-success" />
              ) : (
                <TrendingDown className="w-5 h-5 text-danger" />
              )}
              <p
                className={`text-xl font-semibold ${
                  totalPL >= 0 ? "text-success" : "text-danger"
                }`}
              >
                {totalPL >= 0 ? "+" : ""}${totalPL.toFixed(2)}
              </p>
            </div>
          </div>
        </div>
      </motion.div>

      {/* Holdings */}
      {detailed && (
        <div className="glass rounded-xl p-6">
          <h3 className="text-xl font-bold text-primary mb-4">Holdings</h3>
          <div className="space-y-3">
            {portfolio.holdings && portfolio.holdings.length > 0 ? (
              portfolio.holdings.map((holding, idx) => (
                <motion.div
                  key={holding.stock}
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: idx * 0.1 }}
                  className="bg-dark-200 rounded-lg p-4 hover:bg-dark-100 transition-colors"
                >
                  <div className="flex justify-between items-start mb-2">
                    <div>
                      <h4 className="text-lg font-semibold text-white">
                        {holding.stock}
                      </h4>
                      <p className="text-sm text-gray-400">
                        {holding.quantity} shares @ $
                        {holding.avg_buy_price.toFixed(2)}
                      </p>
                    </div>
                    <div className="text-right">
                      <p className="text-lg font-semibold text-primary">
                        ${(holding.quantity * holding.current_price).toFixed(2)}
                      </p>
                      <p
                        className={`text-sm font-medium flex items-center justify-end space-x-1 ${
                          holding.profit_loss >= 0
                            ? "text-success"
                            : "text-danger"
                        }`}
                      >
                        {holding.profit_loss >= 0 ? (
                          <TrendingUp className="w-4 h-4" />
                        ) : (
                          <TrendingDown className="w-4 h-4" />
                        )}
                        <span>
                          {holding.profit_loss >= 0 ? "+" : ""}$
                          {holding.profit_loss.toFixed(2)}
                        </span>
                      </p>
                    </div>
                  </div>
                  <div className="w-full bg-dark-300 rounded-full h-2 mt-2">
                    <div
                      className={`h-2 rounded-full transition-all ${
                        holding.profit_loss >= 0 ? "bg-success" : "bg-danger"
                      }`}
                      style={{
                        width: `${Math.min(
                          Math.abs(
                            holding.profit_loss / holding.avg_buy_price
                          ) * 100,
                          100
                        )}%`,
                      }}
                    ></div>
                  </div>
                </motion.div>
              ))
            ) : (
              <p className="text-center text-gray-500 py-8">No holdings yet</p>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
