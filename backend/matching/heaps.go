package matching

import "trading/models"

// MaxHeap is used for BUY orders.
// Buyers want the highest price first (max-heap)
type MaxHeap []*models.Order

func (h MaxHeap) Len() int { return len(h) }

func (h MaxHeap) Less(i, j int) bool {
	// Higher price gets priority
	if h[i].Price != h[j].Price {
		return h[i].Price > h[j].Price
	}
	// If prices are same → earlier order goes first
	return h[i].Timestamp.Before(h[j].Timestamp)
}

func (h MaxHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *MaxHeap) Push(x any) {
	*h = append(*h, x.(*models.Order))
}

func (h *MaxHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

// MinHeap is used for SELL orders.
// Sellers want the lowest price first (min-heap)
type MinHeap []*models.Order

func (h MinHeap) Len() int { return len(h) }

func (h MinHeap) Less(i, j int) bool {
	// Lower price gets priority
	if h[i].Price != h[j].Price {
		return h[i].Price < h[j].Price
	}
	// If prices are same → earlier order goes first
	return h[i].Timestamp.Before(h[j].Timestamp)
}

func (h MinHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *MinHeap) Push(x any) {
	*h = append(*h, x.(*models.Order))
}

func (h *MinHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}
