package main

type MinHeap struct {
	arr []int
}

func left(i int) int {
	return 2*i + 1
}
func right(i int) int {
	return left(i) + 1
}

func parent(i int) int {
	return (i - 1) / 2
}

func (h *MinHeap) swap(i, j int) {
	h.arr[i], h.arr[j] = h.arr[j], h.arr[i]
}
func (h *MinHeap) size() int {
	return len(h.arr)
}

func (h *MinHeap) insert(x int) {
	h.arr = append(h.arr, x)
	i := h.size() - 1
	for i > 0 {
		if h.arr[i] < h.arr[parent(i)] {
			h.swap(i, parent(i))
			i = parent(i)
		} else {
			break
		}
	}
}

func (h *MinHeap) top() int {
	if h.size() == 0 {
		panic("min heap is empty")
	}

	return h.arr[0]
}

func (h *MinHeap) pop() int {
	top := h.top()
	h.swap(0, h.size()-1)
	h.arr = h.arr[:h.size()-1]

	i := 0
	for left(i) < h.size() {
		index := i
		if h.arr[left(i)] < h.arr[i] {
			index = left(i)
		}

		if right(i) < h.size() && h.arr[right(i)] < h.arr[index] {
			index = right(i)
		}

		if i == index {
			break
		}

		h.swap(i, index)
		i = index
	}

	return top
}

type MaxHeap struct {
	arr []int
}

func (h *MaxHeap) swap(i, j int) {
	h.arr[i], h.arr[j] = h.arr[j], h.arr[i]
}
func (h *MaxHeap) size() int {
	return len(h.arr)
}

func (h *MaxHeap) insert(x int) {
	h.arr = append(h.arr, x)
	i := h.size() - 1
	for i > 0 {
		if h.arr[i] > h.arr[parent(i)] {
			h.swap(i, parent(i))
			i = parent(i)
		} else {
			break
		}
	}
}

func (h *MaxHeap) top() int {
	if h.size() == 0 {
		panic("max heap is empty")
	}

	return h.arr[0]
}

func (h *MaxHeap) pop() int {
	top := h.top()
	h.swap(0, h.size()-1)
	h.arr = h.arr[:h.size()-1]

	i := 0
	for left(i) < h.size() {
		index := i
		if h.arr[left(i)] > h.arr[i] {
			index = left(i)
		}

		if right(i) < h.size() && h.arr[right(i)] > h.arr[index] {
			index = right(i)
		}

		if i == index {
			break
		}

		h.swap(i, index)
		i = index
	}

	return top
}
