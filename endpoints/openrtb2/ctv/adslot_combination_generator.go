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
	currentCombinationCount int
	currentCombination      *[]uint64

	totalExpectedCombinations uint64     // indicates total number for possible combinations
	combinations              [][]uint64 // May contains some/all combinations at given point of time

	state snapshot

	// configurations

	// Indicates whether this algorithm should consider repetitations
	// For Example: Input durations are 10 23 40 56. For duration 23 there are
	// multiple ads present. In such case if this value is true, algorithm will generate
	// repetitations only for 23 duration.
	// NOTE: Repetitations will be of consecative durations only.
	// It means 10,23,23,23  10,23,23,56 will be generated
	// But 10,23,40,23  23, 10, 23, 23 will not be generated
	allowRepetitationsForEligibleDurations bool
}

type snapshot struct {
	currentSlotIndex     int
	baseCombination      []uint64
	baseCombinationIndex int
}

// Init ...
func (c *AdSlotDurationCombinations) Init(podMindDuration, podMaxDuration, minAds, maxAds int64, slotDurations []uint64, allowRepetitationsForEligibleDurations bool) {
	c.noOfSlots = len(c.slotDurations)
	c.podMinDuration = uint64(podMindDuration)
	c.podMaxDuration = uint64(podMaxDuration)
	c.minAds = uint64(minAds)
	c.maxAds = uint64(maxAds)
	c.slotDurations = slotDurations
	c.currentCombinationCount = 0
	c.state = snapshot{}
	c.state.currentSlotIndex = 0
	c.currentCombination = new([]uint64)
	// default configurations
	c.allowRepetitationsForEligibleDurations = allowRepetitationsForEligibleDurations

	// compute no of possible combinations (without validations)
	// using configurationss
	c.totalExpectedCombinations = compute(c, c.maxAds, true)
	// c.combinations = make([][]uint64, c.totalExpectedCombinations)
	print("Allow Repeatation = %v", c.allowRepetitationsForEligibleDurations)
	print("Total possible combinations (without validations) = %v ", c.totalExpectedCombinations)
}

//Next - Get next ad slot combination
//returns empty array if next combination is not present
func (c *AdSlotDurationCombinations) Next() []uint64 {
	// iteratePolicy = 1. dfs 2.bfs
	//return c.next()
	// if c.currentCombinationCount == 68 {
	// 	fmt.Println("check")
	// }
	return c.lazyNext()
	//Next()
}

func (c *AdSlotDurationCombinations) lazyNext() []uint64 {
	c.search(uint64(c.state.currentSlotIndex), false, 1)
	val := *c.currentCombination
	*c.currentCombination = c.state.baseCombination
	return val
}

// state to store i, newComb
func (c *AdSlotDurationCombinations) next() []uint64 {
	c.search(0, true, 1)
	print("Total combinations generated = %v", c.currentCombinationCount)
	print("Total combinations expected = %v", c.totalExpectedCombinations)
	return nil
}

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

	var noOfCombinations *big.Int

	if c.allowRepetitationsForEligibleDurations {
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

		noOfCombinations = nmrt.Div(&nmrt, d3)
	} else {
		// compute combintations without repitition
		// Formula (Pure combination Formula)
		//			 n!
		//      ------------
		//       r! (n - r)!
		n := uint64(len(c.slotDurations))
		r := computeCombinationForTotalAds
		if r > n {
			noOfCombinations = big.NewInt(0)
			print("Can not generate combination for maxads = %v, with  %v input bid response durations and repeatations allowed", r, n)
			return noOfCombinations.Uint64()
		}
		numerator := fact(n)
		d1 := fact(r)
		d2 := fact(n - r)
		denominator := d1.Mul(&d1, &d2)
		noOfCombinations = numerator.Div(&numerator, denominator)
	}

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

	if doRecursion {
		//*c.combinations = append(*c.combinations, *c.currentCombination)
		if c.currentCombinationCount == 64 {
			fmt.Println("test")
		}

		print("%v", *c.currentCombination)
		val := make([]uint64, len(*c.currentCombination))
		copy(val, *c.currentCombination)
		//c.combinations[c.currentCombinationCount] = val
		c.combinations = append(c.combinations, val)

	}

	c.currentCombinationCount++
}

func (c *AdSlotDurationCombinations) search(slotIndex uint64 /*, baseCombination []uint64*/, doRecursion bool, recCount int) {

	if c.totalExpectedCombinations <= 0 {
		return
	}

	var baseCombination []uint64

	// base := make([]uint64, 0)
	// if c.maxAds > 1 {
	// 	base = *c.currentCombination
	// }

	baseCombination = *c.currentCombination
	maxCombinationLength := int(c.maxAds)

	// stop when total length of base combination
	// is equal  to maxads
	if uint64(len(baseCombination)) == c.maxAds {

		if !doRecursion {
			baseCombination = c.state.baseCombination[:len(c.state.baseCombination)-1]
		} else {

			return
		}

	}

	for i := int(slotIndex); i < len(c.slotDurations); i++ {

		// if !doRecursion && i == len(c.slotDurations)-1 && uint64(len(baseCombination)) == c.maxAds {
		// 	baseCombination = baseCombination[:len(baseCombination)-1]
		// }

		if !c.allowRepetitationsForEligibleDurations {
			// check if c.slotDurations[i] value is already
			// present in baseCombination
			// only in consecutive manner
			_, exists := find(baseCombination, c.slotDurations[i])
			if exists && doRecursion {
				continue // with next elememt
			}

			if exists {
				//if len(newCombination) == maxCombinationLength && i < maxCombinationLength {
				c.state.baseCombination = c.state.baseCombination[:len(c.state.baseCombination)-1]
				c.state.currentSlotIndex++
				if c.state.currentSlotIndex == len(c.slotDurations) {
					c.state.currentSlotIndex = 0
					newCombination := append(baseCombination, c.slotDurations[i])
					determineSlotIndex(c, newCombination, baseCombination, maxCombinationLength)
					baseCombination = c.state.baseCombination
				}

				i = c.state.currentSlotIndex
				//}
			}
		}

		newCombination := append(baseCombination, c.slotDurations[i])

		// fmt.Printf("Level: %v, Base Comb  : %v\t:: ", recCount, baseCombination)

		fmt.Printf("%v ::\t", baseCombination)

		updateCurrentCombination(c, newCombination, doRecursion)
		if doRecursion {
			//generateSubTree(c, i, true)
			//c.search(uint64(i), baseCombination, doRecursion)
			c.search(uint64(i), doRecursion, recCount+1)
		} else {
			// store base combination

			c.state.baseCombination = newCombination

			// maxCombinationLength := len(c.slotDurations)

			// if len(newCombination) = len(input slot array)
			// then increment last index by 1 till  it  not reaches = len(input slot array)) -1
			if len(newCombination) == maxCombinationLength && i < maxCombinationLength {
				c.state.baseCombination = c.state.baseCombination[:len(c.state.baseCombination)-1]
				c.state.currentSlotIndex++
				determineSlotIndex(c, newCombination, baseCombination, maxCombinationLength)
			}

			// if i+1 == len(input slot array) and basecombination size = len()
			// then reset c.state.currentSlotIndex
			// but not to previous one
			// e.g. 4 5 8 7 if previous is 4 then now it must be 5
			// i > maxCombinationLength : require when e.g. maxads =2, input = 5 len
			if i >= maxCombinationLength /*|| maxCombinationLength == 1*/ {
				if len(newCombination) >= maxCombinationLength {

					// c.state.baseCombination = c.state.baseCombination[:len(c.state.baseCombination)-2]
					c.state.currentSlotIndex++

					// if c.state.currentSlotIndex == maxCombinationLength {
					// 	c.state.currentSlotIndex = 0
					// }

					determineSlotIndex(c, newCombination, baseCombination, maxCombinationLength)

					// // remove indices from last from baseCombination
					// // and assign to c.state.baseCombination
					// if upperRange != -1 && upperRange < len(baseCombination) {
					// 	c.state.baseCombination = baseCombination[:upperRange]
					// }

				} else {
					// there few more possible combination w.r.t. last element
					//fmt.Println("last elemt possible comb")
				}
			}

			return
		}

	}
}

func determineSlotIndex(c *AdSlotDurationCombinations, newCombination, baseCombination []uint64, maxCombinationLength int) int {

	// remove indices from last index such that
	// no of indices  to remove = no of last value present in newCombination

	totalInputSlots := len(c.slotDurations)

	lastNoOccurance := 0
	// find no of occurances of last value present in newCombination
	for _, val := range newCombination {
		if val == c.slotDurations[totalInputSlots-1] {
			lastNoOccurance++
		}
	}

	// if lastNoOccurance == maxCombinationLength {
	// 	// print(" lastNoOccurance == maxCombinationLength =  %v", lastNoOccurance)
	// 	return 0
	// }
	if lastNoOccurance == maxCombinationLength {
		//if lastNoOccurance == totalInputSlots {
		//print(" lastNoOccurance == maxCombinationLength =  %v", lastNoOccurance)
		return 0
	}

	upperRange := len(baseCombination) - lastNoOccurance

	if upperRange >= len(baseCombination) {
		return -1
	}

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
			if c.state.currentSlotIndex == totalInputSlots {
				c.state.currentSlotIndex = 0
			}

			break
		}
	}

	// remove indices from last from baseCombination
	// and assign to c.state.baseCombination
	if upperRange != -1 && upperRange < len(baseCombination) {
		c.state.baseCombination = baseCombination[:upperRange]
	}
	return upperRange
}

func find(array []uint64, element uint64) (int, bool) {
	for index, aE := range array {
		if aE == element {
			return index, true
		}
	}
	return -1, false
}
