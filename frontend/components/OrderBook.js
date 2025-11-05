"use client";

import { useOrderBook } from "@/lib/hooks";
import { Loader2, RefreshCw } from "lucide-react";

export default function OrderBook({ stock }) {
  const { orderBook, loading, error, refresh } = useOrderBook(stock);

  if (error) {
    return (
      <div className="glass rounded-xl p-6">
        <p className="text-danger">Error loading order book: {error}</p>
      </div>
    );
  }

  return (
    <div className="glass rounded-xl p-6">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-primary">
          Order Book - {stock}
        </h2>
        <button
          onClick={refresh}
          disabled={loading}
          className="p-2 hover:bg-dark-100 rounded-lg transition-colors"
        >
          <RefreshCw className={`w-5 h-5 ${loading ? "animate-spin" : ""}`} />
        </button>
      </div>

      <div className="grid md:grid-cols-2 gap-6">
        {/* Bids (Buy Orders) */}
        <div>
          <h3 className="text-lg font-semibold text-success mb-4 flex items-center">
            <span className="w-3 h-3 bg-success rounded-full mr-2"></span>
            Bids (Buy)
          </h3>
          <div className="space-y-2 max-h-[400px] overflow-y-auto pr-2">
            {loading ? (
              <div className="flex justify-center py-8">
                <Loader2 className="w-8 h-8 animate-spin text-primary" />
              </div>
            ) : orderBook.bids && orderBook.bids.length > 0 ? (
              orderBook.bids.slice(0, 15).map((bid, idx) => (
                <div
                  key={idx}
                  className="flex justify-between items-center p-3 bg-success/10 border-l-4 border-success rounded-lg hover:bg-success/20 transition-colors"
                >
                  <span className="font-mono text-primary font-semibold">
                    ${bid.price.toFixed(2)}
                  </span>
                  <span className="text-gray-400">{bid.quantity} shares</span>
                </div>
              ))
            ) : (
              <p className="text-center text-gray-500 py-8">
                No bids available
              </p>
            )}
          </div>
        </div>

        {/* Asks (Sell Orders) */}
        <div>
          <h3 className="text-lg font-semibold text-danger mb-4 flex items-center">
            <span className="w-3 h-3 bg-danger rounded-full mr-2"></span>
            Asks (Sell)
          </h3>
          <div className="space-y-2 max-h-[400px] overflow-y-auto pr-2">
            {loading ? (
              <div className="flex justify-center py-8">
                <Loader2 className="w-8 h-8 animate-spin text-primary" />
              </div>
            ) : orderBook.asks && orderBook.asks.length > 0 ? (
              orderBook.asks.slice(0, 15).map((ask, idx) => (
                <div
                  key={idx}
                  className="flex justify-between items-center p-3 bg-danger/10 border-l-4 border-danger rounded-lg hover:bg-danger/20 transition-colors"
                >
                  <span className="font-mono text-primary font-semibold">
                    ${ask.price.toFixed(2)}
                  </span>
                  <span className="text-gray-400">{ask.quantity} shares</span>
                </div>
              ))
            ) : (
              <p className="text-center text-gray-500 py-8">
                No asks available
              </p>
            )}
          </div>
        </div>
      </div>

      {orderBook.bids &&
        orderBook.bids.length > 0 &&
        orderBook.asks &&
        orderBook.asks.length > 0 && (
          <div className="mt-6 p-4 bg-dark-200 rounded-lg">
            <div className="flex justify-between items-center">
              <div>
                <p className="text-xs text-gray-400 mb-1">Best Bid</p>
                <p className="text-lg font-mono font-semibold text-success">
                  ${orderBook.bids[0].price.toFixed(2)}
                </p>
              </div>
              <div className="text-center">
                <p className="text-xs text-gray-400 mb-1">Spread</p>
                <p className="text-lg font-mono font-semibold text-primary">
                  $
                  {(orderBook.asks[0].price - orderBook.bids[0].price).toFixed(
                    2
                  )}
                </p>
              </div>
              <div className="text-right">
                <p className="text-xs text-gray-400 mb-1">Best Ask</p>
                <p className="text-lg font-mono font-semibold text-danger">
                  ${orderBook.asks[0].price.toFixed(2)}
                </p>
              </div>
            </div>
          </div>
        )}
    </div>
  );
}
