package ctv

import (
	"fmt"
	"math/big"
)

//AdSlotDurationCombinations holds all the combinations based
//on Video Ad Pod request and Bid Response Max duration
type AdSlotDurationCombinations struct {
	podMinDuration uint64 // Pod Minimum duration value present in origin Video Ad Pod Request
	podMaxDuration uint64 // Pod Maximum duration value present in origin Video Ad Pod Request
	minAds         uint64 // Minimum Ads value present in origin Video Ad Pod Request
	maxAds         uint64 // Maximum Ads value present in origin Video Ad Pod Request

	slotDurations []uint64 // input slot durations for which
	noOfSlots     int      // Number of slots to be consider (from left to right)

	// cursors
	slotIndex               int
	currentCombinationCount int
	currentCombination      []uint64

	totalExpectedCombinations uint64     // indicates total number for possible combinations
	combinations              [][]uint64 // May contains some/all combinations at given point of time
}

// Init ...
func (c *AdSlotDurationCombinations) Init(podMindDuration, podMaxDuration, minAds, maxAds int64, slotDurations []uint64) {
	c.noOfSlots = len(c.slotDurations)
	c.podMinDuration = uint64(podMindDuration)
	c.podMaxDuration = uint64(podMaxDuration)
	c.minAds = uint64(minAds)
	c.maxAds = uint64(maxAds)
	c.slotDurations = slotDurations
	c.totalExpectedCombinations = compute(c, c.maxAds)
	c.slotIndex = 0
	c.currentCombinationCount = 0
	print("Total possible combinations (without validations) = %v ", c.totalExpectedCombinations)
}

//Next - Get next ad slot combination
//returns empty array if next combination is not present
func (c *AdSlotDurationCombinations) Next() []uint64 {
	// iteratePolicy = 1. dfs 2.bfs
	return c.next()
}

func (c *AdSlotDurationCombinations) next() []uint64 {
	for i := 0; i < len(c.slotDurations); i++ {
		newComb := make([]uint64, 1)
		newComb[0] = c.slotDurations[i]
		updateCurrentCombination(c, newComb)
		generateSubTree(c, i)
	}
	print("Total combinations generated = %v", c.currentCombinationCount)
	print("Total combinations expected = %v", c.totalExpectedCombinations)
	return nil
}

func (c *AdSlotDurationCombinations) generateSubTree(slotIndex uint64, baseCombination []uint64) {
	// stop when total length of base combination
	// is equal  to maxads
	if uint64(len(baseCombination)) == c.maxAds {
		return
	}

	for i := int(slotIndex); i < len(c.slotDurations); i++ {
		newCombination := append(baseCombination, c.slotDurations[i])
		updateCurrentCombination(c, newCombination)
		generateSubTree(c, i)
	}
}

// HasNext - true if next combination is present
// false if not
func (c AdSlotDurationCombinations) HasNext() bool {
	// return uint64(c.slotIndex) < c.totalExpectedCombinations
	return uint64(c.currentCombinationCount) < c.totalExpectedCombinations
}

func compute(c *AdSlotDurationCombinations, computeCombinationForTotalAds uint64) uint64 {
	if computeCombinationForTotalAds == 0 {
		return 0
	}
	// Formula
	//		(r + n - 1)!
	//      ------------
	//       r! (n - 1)!
	n := uint64(len(c.slotDurations))
	r := uint64(computeCombinationForTotalAds)
	d1 := fact(uint64(r))
	d2 := fact(n - 1)
	d3 := d1.Mul(&d1, &d2)
	nmrt := fact(r + n - 1)

	noOfCombinations := nmrt.Div(&nmrt, d3)
	print("%v", noOfCombinations)
	return noOfCombinations.Uint64() + compute(c, computeCombinationForTotalAds-1)
}

func fact(no uint64) big.Int {
	if no == 0 {
		return *big.NewInt(int64(1))
	}
	var bigNo big.Int
	bigNo.SetUint64(no)

	fact := fact(no - 1)
	mult := bigNo.Mul(&bigNo, &fact)

	return *mult
	//return mult.Uint64()
	//return no * fact(no-1)
}

func print(format string, v ...interface{}) {
	// log.Printf(format, v...)
	fmt.Printf(format+"\n", v...)
}

func updateCurrentCombination(c *AdSlotDurationCombinations, newCombination []uint64) {
	c.currentCombination = newCombination
	c.currentCombinationCount++
	print("%v", c.currentCombination)
}

func generateSubTree(c *AdSlotDurationCombinations, slotIndex int) {
	base := make([]uint64, 0)
	base = c.currentCombination
	c.generateSubTree(uint64(slotIndex), base)
}
