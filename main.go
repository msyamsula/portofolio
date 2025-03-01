package main

import "fmt"

var (
	inf = int(1e10)
)

// func recurse(nums []int, maxN, minN int) []int {
// 	pivot := (maxN + minN) / 2

// 	n := len(nums)
// 	count := 0
// 	xor := 0
// 	totalXor := 0
// 	for i := 0; i < n; i++ {
// 		if nums[i] <= pivot {
// 			count++
// 			xor ^= nums[i]
// 		}
// 		totalXor ^= nums[i]
// 	}

// 	if count%2 == 1 {
// 		return []int{xor, xor ^ totalXor}
// 	}

// 	if xor == 0 {
// 		return recurse(nums, maxN, pivot)
// 	}

// 	return recurse(nums, minN, pivot)
// }

// func singleNumber(nums []int) []int {

// 	n := len(nums)
// 	// idx := rand.New(rand.NewSource(time.Now().Unix())).Int63n(int64(n))

// 	// pivot := nums[idx]
// 	maxN := -inf
// 	minN := inf
// 	for i := 0; i < n; i++ {
// 		maxN = max(maxN, nums[i])
// 		minN = min(minN, nums[i])
// 	}

// 	return recurse(nums, maxN, minN)
// 	// idx := rand.NewSource()

// 	/*
// 	   totalXor := a^b

// 	   even:
// 	       a,b,pivot
// 	       a == odd
// 	           we find u1 and u2
// 	       a == even
// 	           xor == 0
// 	               work with the righ
// 	           xor != 0
// 	               work with left
// 	   odd:
// 	       a,b,pivot
// 	       a == odd:
// 	           there is one one the left, got u1 and u2
// 	       a == even:
// 	           xor == 0:
// 	               work with right
// 	           xor != 0
// 	               work with left
// 	*/

// }
func main() {
	a := -3
	b := 0
	fmt.Println((a + b) / 2)
}
