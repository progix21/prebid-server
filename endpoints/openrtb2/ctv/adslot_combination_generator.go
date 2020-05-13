package ctv

import (
	"log"
)

//AdSlotDurationCombinations holds all the combinations based
//on Video Ad Pod request and Bid Response Max duration
type AdSlotDurationCombinations struct {
	podMinDuration int64 // Pod Minimum duration value present in origin Video Ad Pod Request
	podMaxDuration int64 // Pod Maximum duration value present in origin Video Ad Pod Request
	minAds         int64 // Minimum Ads value present in origin Video Ad Pod Request
	maxAds         int64 // Maximum Ads value present in origin Video Ad Pod Request

	slotDurations []int64 // input slot durations for which
	noOfSlots     int     // Number of slots to be consider (from left to right)

	// cursors
	slotIndex               int
	currentCombinationCount int
	currentCombination      []int64

	totalCombinations int       // indicates total number for possible combinations
	combinations      [][]int64 // May contains some/all combinations at given point of time
}

// Init ...
func (c *AdSlotDurationCombinations) Init(podMindDuration, podMaxDuration, minAds, maxAds int64, slotDurations []int64) {
	c.noOfSlots = len(c.slotDurations)
	c.podMinDuration = podMindDuration
	c.podMaxDuration = podMaxDuration
	c.minAds = minAds
	c.maxAds = maxAds
	c.slotDurations = slotDurations
	c.totalCombinations = fact(len(c.slotDurations))
	c.slotIndex = 0
	c.currentCombinationCount = 0
	print("Total possible combinations (without validations) = %v ", c.totalCombinations)
}

//Next - Get next ad slot combination
//returns empty array if next combination is not present
func (c *AdSlotDurationCombinations) Next() []int64 {
	return c.next()
}

func (c *AdSlotDurationCombinations) next() []int64 {
	for i := 0; i < len(c.slotDurations); i++ {
		c.currentCombination = make([]int64, 1)
		c.currentCombination[0] = c.slotDurations[i]
		c.currentCombinationCount++
		print("%v", c.currentCombination)
		base := make([]int64, 0)
		base = c.currentCombination
		c.generateSubTree(int64(i), base)
	}
	print("Total combinations generated = %v", c.currentCombinationCount)
	return nil
}

func (c *AdSlotDurationCombinations) generateSubTree(slotIndex int64, baseCombination []int64) {

	for i := int(slotIndex); i < len(c.slotDurations); i++ {
		// c.currentCombination = make([]int64, 1)
		// c.currentCombination[0] = c.slotDurations[slotIndex]
		c.currentCombination = append(baseCombination, c.slotDurations[i])
		c.currentCombinationCount++
		print("%v", c.currentCombination)
		base := make([]int64, 0)
		base = c.currentCombination
		c.generateSubTree(slotIndex+1, base)
	}
}

// HasNext - true if next combination is present
// false if not
func (c AdSlotDurationCombinations) HasNext() bool {
	return c.slotIndex < c.totalCombinations
}

func fact(no int) int {
	if no == 0 {
		return 1
	}
	return no * fact(no-1)
}

func print(format string, v ...interface{}) {
	log.Printf(format, v...)
}
