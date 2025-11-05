package matching

import "trading/models"

// MaxHeap implements heap.Interface for buy orders (highest price first)
type MaxHeap []*models.Order

func (h MaxHeap) Len() int { return len(h) }

func (h MaxHeap) Less(i, j int) bool {
	// Higher price has priority
	if h[i].Price != h[j].Price {
		return h[i].Price > h[j].Price
	}
	// If prices equal, earlier timestamp wins (FIFO)
	return h[i].Timestamp.Before(h[j].Timestamp)
}

func (h MaxHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *MaxHeap) Push(x interface{}) {
	*h = append(*h, x.(*models.Order))
}

func (h *MaxHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// MinHeap implements heap.Interface for sell orders (lowest price first)
type MinHeap []*models.Order

func (h MinHeap) Len() int { return len(h) }

func (h MinHeap) Less(i, j int) bool {
	// Lower price has priority
	if h[i].Price != h[j].Price {
		return h[i].Price < h[j].Price
	}
	// If prices equal, earlier timestamp wins (FIFO)
	return h[i].Timestamp.Before(h[j].Timestamp)
}

func (h MinHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *MinHeap) Push(x interface{}) {
	*h = append(*h, x.(*models.Order))
}

func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
