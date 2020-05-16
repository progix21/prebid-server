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
	currentCombination      *[]uint64

	totalExpectedCombinations uint64     // indicates total number for possible combinations
	combinations              [][]uint64 // May contains some/all combinations at given point of time

	state snapshot
}

type snapshot struct {
	currentSlotIndex        int
	currentSlotIndexHelper  int
	currentSlotIndexHelper1 int
	baseCombination         []uint64
	baseCombinationIndex    int
}

// Init ...
func (c *AdSlotDurationCombinations) Init(podMindDuration, podMaxDuration, minAds, maxAds int64, slotDurations []uint64) {
	c.noOfSlots = len(c.slotDurations)
	c.podMinDuration = uint64(podMindDuration)
	c.podMaxDuration = uint64(podMaxDuration)
	c.minAds = uint64(minAds)
	c.maxAds = uint64(maxAds)
	c.slotDurations = slotDurations
	c.totalExpectedCombinations = compute(c, c.maxAds, true)
	c.slotIndex = 0
	c.currentCombinationCount = 0
	c.state = snapshot{}
	c.state.currentSlotIndex = 0
	c.currentCombination = new([]uint64)

	c.state.currentSlotIndexHelper = 0
	c.state.currentSlotIndexHelper1 = 0
	print("Total possible combinations (without validations) = %v ", c.totalExpectedCombinations)
}

//Next - Get next ad slot combination
//returns empty array if next combination is not present
func (c *AdSlotDurationCombinations) Next() []uint64 {
	// iteratePolicy = 1. dfs 2.bfs
	return c.next()
	//return c.lazyNext()
}

func (c *AdSlotDurationCombinations) lazyNext() []uint64 {

	//generateSubTree(c, c.state.currentSlotIndex, false)
	//c.search(uint64(c.state.currentSlotIndex), *c.currentCombination, false)
	c.search(uint64(c.state.currentSlotIndex), false, 1)
	val := *c.currentCombination
	*c.currentCombination = c.state.baseCombination
	return val
}

// state to store i, newComb
func (c *AdSlotDurationCombinations) next() []uint64 {
	//newComb := make([]uint64, 0)
	//c.search(0, newComb, true)
	c.search(0, true, 1)
	print("Total combinations generated = %v", c.currentCombinationCount)
	print("Total combinations expected = %v", c.totalExpectedCombinations)
	return nil
}

func verticalCombgenerator(c AdSlotDurationCombinations) {

	vcombs := compute(&c, c.minAds, false)

	nodeLength := c.minAds
	lastValueIndexCount := uint64(0)
	lastValueIndexCountBack := lastValueIndexCount + 1
	for vcombIndex := uint64(0); vcombIndex < vcombs; vcombIndex++ {
		node := make([]uint64, nodeLength)
		for i := uint64(0); i < nodeLength; i++ {
			node[i] = c.slotDurations[0]
			if i == nodeLength-1 {
				node[i] = c.slotDurations[lastValueIndexCount]
				lastValueIndexCount++
				if lastValueIndexCount == uint64(len(c.slotDurations)) {
					lastValueIndexCount = lastValueIndexCountBack
					lastValueIndexCountBack++
					if lastValueIndexCountBack == uint64(len(c.slotDurations)) {
						lastValueIndexCountBack = 0
					}
				}
			}
		}
		//print("%v", node)
	}

}

// func (c *AdSlotDurationCombinations) generateSubTree(slotIndex uint64 /*baseCombination []uint64,*/, doRecursion bool) {

// 	base := make([]uint64, 0)
// 	base = *c.currentCombination

// 	baseCombination := base

// 	// stop when total length of base combination
// 	// is equal  to maxads
// 	if uint64(len(baseCombination)) == c.maxAds {
// 		if !doRecursion {

// 		} else {
// 			return
// 		}
// 	}

// 	c.search(slotIndex, baseCombination, doRecursion)
// }

// HasNext - true if next combination is present
// false if not
func (c AdSlotDurationCombinations) HasNext() bool {
	// return uint64(c.slotIndex) < c.totalExpectedCombinations
	return uint64(c.currentCombinationCount) < c.totalExpectedCombinations
}

func compute(c *AdSlotDurationCombinations, computeCombinationForTotalAds uint64, recursion bool) uint64 {
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
	if recursion {
		return noOfCombinations.Uint64() + compute(c, computeCombinationForTotalAds-1, recursion)
	}
	return noOfCombinations.Uint64()
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
}

func print(format string, v ...interface{}) {
	// log.Printf(format, v...)
	fmt.Printf(format+"\n", v...)
}

func updateCurrentCombination(c *AdSlotDurationCombinations, newCombination []uint64, doRecursion bool) {
	*c.currentCombination = newCombination
	c.currentCombinationCount++
	if doRecursion {
		print("%v", *c.currentCombination)
	}
}

// //generateSubTree - wrapper around AdSlotDurationCombinations.generateSubTree
// func generateSubTree(c *AdSlotDurationCombinations, slotIndex int, doRecursion bool) {
// 	base := make([]uint64, 0)
// 	base = *c.currentCombination
// 	c.generateSubTree(uint64(slotIndex), base, doRecursion)
// }

func (c *AdSlotDurationCombinations) search(slotIndex uint64 /*, baseCombination []uint64*/, doRecursion bool, recCount int) {

	var baseCombination []uint64

	base := make([]uint64, 0)
	base = *c.currentCombination

	baseCombination = base

	// stop when total length of base combination
	// is equal  to maxads
	if uint64(len(baseCombination)) == c.maxAds {
		if !doRecursion {

		} else {
			return
		}
	}

	for i := int(slotIndex); i < len(c.slotDurations); i++ {
		newCombination := append(baseCombination, c.slotDurations[i])

		fmt.Printf("Level: %v, Base Comb  : %v\t:: ", recCount, baseCombination)

		updateCurrentCombination(c, newCombination, doRecursion)
		if doRecursion {
			//generateSubTree(c, i, true)
			//c.search(uint64(i), baseCombination, doRecursion)
			c.search(uint64(i), doRecursion, recCount+1)
		} else {
			// store base combination
			c.state.baseCombination = newCombination

			inputSlotsLength := len(c.slotDurations)

			// if len(newCombination) = len(input slot array)
			// then increment last index by 1 till  it  not reaches = len(input slot array)) -1
			if len(newCombination) == inputSlotsLength && i+1 < inputSlotsLength {
				c.state.baseCombination = c.state.baseCombination[:len(c.state.baseCombination)-1]
				c.state.currentSlotIndex++
			}

			// if i+1 == len(input slot array) and basecombination size = len()
			// then reset c.state.currentSlotIndex
			// but not to previous one
			// e.g. 4 5 8 7 if previous is 4 then now it must be 5
			if i+1 == inputSlotsLength {
				if len(newCombination) == inputSlotsLength {

					// c.state.baseCombination = c.state.baseCombination[:len(c.state.baseCombination)-2]
					c.state.currentSlotIndex++

					if c.state.currentSlotIndex == inputSlotsLength {
						c.state.currentSlotIndex = 0
					}

					// remove indices from last index such that
					// no of indices  to remove = no of last value present in newCombination

					lastNoOccurance := 0
					// find no of occurances of last value present in newCombination
					for _, val := range newCombination {
						if val == c.slotDurations[inputSlotsLength-1] {
							lastNoOccurance++
						}
					}

					if lastNoOccurance == inputSlotsLength {
						print(" lastNoOccurance == inputSlotsLength =  %v", lastNoOccurance)
						return
					}

					upperRange := len(baseCombination) - lastNoOccurance

					// get the next item required for plotting combination
					// LOGIC : to determine that
					//Level: 4, Base Comb  : [4 5 7]	:: [4 5 7 7]
					//Level: 2, Base Comb  : [4]		:: [4 8]
					// we have removed 5 and 7 from Level 4 base comb
					// the number at last removed place from R -> L was 5
					// as per input list 4 5 8 7, we should select no next to 5
					// i.e. Hence we can form L2 as using its base comb
					// which will be [4,8]
					lastRemovedValue := baseCombination[upperRange]

					// find the  inedex of this value
					for ind, val := range c.slotDurations {
						if val == lastRemovedValue {
							// then get element next to  it
							c.state.currentSlotIndex = ind + 1
							if c.state.currentSlotIndex == inputSlotsLength {
								c.state.currentSlotIndex = 0
							}

							break
						}
					}

					//fmt.Println(nextItemValue)

					// remove indices from last from baseCombination
					// and assign to c.state.baseCombination
					c.state.baseCombination = baseCombination[:upperRange]
					c.state.currentSlotIndexHelper++
					if c.state.currentSlotIndexHelper == inputSlotsLength {
						c.state.currentSlotIndexHelper1++
						c.state.currentSlotIndexHelper = c.state.currentSlotIndexHelper1

					}
					//c.state.currentSlotIndex = c.state.currentSlotIndexHelper

				} else {
					// there few more possible combination w.r.t. last element
					//fmt.Println("last elemt possible comb")
				}
			}

			return
		}

	}
}
