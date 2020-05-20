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
	currentSlotIndex                  int
	baseCombination                   []uint64
	baseCombinationIndex              int
	currentCombinationInitAsPerMinAds bool

	/// new states
	start              uint64
	index              int64
	r                  uint64
	lastCombination    []uint64
	stateUpdated       bool
	valueUpdated       bool
	combinationCounter uint64
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

	/// new states
	c.state.start = uint64(0)
	c.state.index = 0
	c.state.r = c.minAds

}

//Next - Get next ad slot combination
//returns empty array if next combination is not present
func (c *AdSlotDurationCombinations) Next() []uint64 {
	// iteratePolicy = 1. dfs 2.bfs
	return c.next()
	//return c.lazyNext()
	// return c.search1tr()

}

func setCombinationAsPerMinAds(c *AdSlotDurationCombinations, baseValue uint64) []uint64 {
	if c.minAds > 1 {
		currentCombinationLength := c.minAds - 1
		combination := make([]uint64, currentCombinationLength)
		// create combination as  per  0th index element
		for ad := uint64(0); ad < currentCombinationLength; ad++ {
			(combination)[ad] = baseValue
		}
		return combination
	}
	return nil
}

func (c *AdSlotDurationCombinations) lazyNext() []uint64 {
	//if c.state.currentCombinationInitAsPerMinAds == false {
	if len(c.state.baseCombination) <= 1 && c.minAds > 1 {
		// *c.currentCombination = setCombinationAsPerMinAds(c, c.slotDurations[0])
		baseCombination := c.slotDurations[0]
		if c.state.baseCombination != nil && len(c.state.baseCombination) == 1 {
			baseCombination = c.slotDurations[c.state.currentSlotIndex]
		}
		*c.currentCombination = setCombinationAsPerMinAds(c, baseCombination)
		c.state.currentCombinationInitAsPerMinAds = true
	}
	//}
	c.search(uint64(c.state.currentSlotIndex), false, 1)
	val := *c.currentCombination
	*c.currentCombination = c.state.baseCombination
	return val
}

// state to store i, newComb
func (c *AdSlotDurationCombinations) next() []uint64 {
	*c.currentCombination = setCombinationAsPerMinAds(c, c.slotDurations[0])
	c.search(0, true, 1)
	print("Total combinations generated = %v", c.currentCombinationCount)
	print("Total combinations expected = %v", c.totalExpectedCombinations)
	return nil
}

// HasNext - true if next combination is present
// false if not
func (c AdSlotDurationCombinations) HasNext() bool {
	return uint64(c.currentCombinationCount) < c.totalExpectedCombinations
}

func compute(c *AdSlotDurationCombinations, computeCombinationForTotalAds uint64, recursion bool) uint64 {
	if computeCombinationForTotalAds < c.minAds {
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
		// compute combintations without repeatation
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
		if c.currentCombinationCount == 64 {
			fmt.Println("test")
		}

		print("%v", *c.currentCombination)
		val := make([]uint64, len(*c.currentCombination))
		copy(val, *c.currentCombination)
		c.combinations = append(c.combinations, val)

	}

	c.currentCombinationCount++
}

func (c *AdSlotDurationCombinations) search(slotIndex uint64 /*, baseCombination []uint64*/, doRecursion bool, recCount int) {

	if c.totalExpectedCombinations <= 0 {
		return
	}

	var baseCombination []uint64
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

		if doRecursion {
			if recCount == 1 && i > 0 && c.minAds > 1 {
				baseCombination = setCombinationAsPerMinAds(c, c.slotDurations[i])
			}
		}
		if !c.allowRepetitationsForEligibleDurations {
			// check if c.slotDurations[i] value is already
			// present in baseCombination
			// only in consecutive manner
			_, exists := find(baseCombination, c.slotDurations[i])
			if exists && doRecursion {
				continue // with next elememt
			}
			if exists {
				c.state.baseCombination = c.state.baseCombination[:len(c.state.baseCombination)-1]
				c.state.currentSlotIndex++
				if c.state.currentSlotIndex == len(c.slotDurations) {
					c.state.currentSlotIndex = 0
					newCombination := append(baseCombination, c.slotDurations[i])
					determineSlotIndex(c, newCombination, baseCombination, maxCombinationLength)
					baseCombination = c.state.baseCombination
				}
				i = c.state.currentSlotIndex
			}
		}
		newCombination := append(baseCombination, c.slotDurations[i])
		//fmt.Printf("Level: %v, Base Comb  : %v\t:: ", recCount, baseCombination)
		//fmt.Printf("%v ::\t", baseCombination)
		updateCurrentCombination(c, newCombination, doRecursion)
		if doRecursion {
			c.search(uint64(i), doRecursion, recCount+1)
		} else {
			// store base combination
			c.state.baseCombination = newCombination
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
					c.state.currentSlotIndex++
					determineSlotIndex(c, newCombination, baseCombination, maxCombinationLength)
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

	if lastNoOccurance == maxCombinationLength {
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

func (c *AdSlotDurationCombinations) search1tr() [][]uint64 {
	start := uint64(0)
	index := uint64(0)

	merged := c.combinations
	for r := c.minAds; r <= c.maxAds; r++ {
		data := make([]uint64, r)
		c.search1(data, start, index, r, merged, false, 0)
	}
	print("Total combinations generated = %v", c.currentCombinationCount)
	print("Total combinations expected = %v", c.totalExpectedCombinations)
	c.currentCombinationCount = 0 //reset
	result := make([][]uint64, c.totalExpectedCombinations)
	copy(result, c.combinations)
	return result
}

func (c *AdSlotDurationCombinations) search1trlazy() []uint64 {

	//	merged := make([][]uint64, c.totalExpectedCombinations)
	start := c.state.start
	index := c.state.index
	r := c.state.r

	// reset last combination
	// by deleting previous values
	if c.state.lastCombination == nil {
		c.combinations = make([][]uint64, 0)
	}
	merged := c.combinations
	// data := make([]uint64, r)
	data := new([]uint64)
	data = &c.state.lastCombination
	if *data == nil || uint64(len(*data)) != r {
		*data = make([]uint64, r)
	}
	c.state.stateUpdated = false
	c.state.valueUpdated = false
	for ; r <= c.maxAds; r++ {

		//common change
		//index = index + 1
		if c.allowRepetitationsForEligibleDurations {
		} else {
			start = start + 1
		}
		c.search1(*data, start, uint64(index), r, merged, true, 0)
		c.state.stateUpdated = false // reset
		c.state.valueUpdated = false
		break
	}

	result := make([]uint64, len(*data))
	copy(result, *data)
	return result
}

//
//static void combinationUtil(
//	int arr[], int data[], int start, int end, int index, int r)
func (c *AdSlotDurationCombinations) search1(data []uint64, start, index, r uint64, merged [][]uint64, lazyLoad bool, reursionCount int) []uint64 {

	end := uint64(len(c.slotDurations) - 1)

	// Current combination is ready to be printed, print it
	if index == r {
		c.currentCombinationCount++
		data1 := make([]uint64, len(data))
		for j := uint64(0); j < r; j++ {
			// fmt.Print(data[j])
			// fmt.Print(" ")
			data1[j] = data[j]
		}

		c.combinations = append(c.combinations, data1)
		// println("")
		// fmt.Println(c.currentCombinationCount, " :: index	=", index, ", i=", start, " :: slot = ", data)
		fmt.Println(data1)
		c.state.valueUpdated = true
		return data1

	}

	_index := index
	if c.allowRepetitationsForEligibleDurations {

		// for (int i=start; i<=end && end+1 >= r-index; i++)
		// {
		// 	data[index] = arr[i];
		// 		combinationUtil(arr, data, i, end, index+1, r);
		// }

		for i := start; i <= end && end+1+c.maxAds >= r-index; i++ {
			// if lazyLoad && c.state.stateUpdated {
			// 	return data
			// }

			if lazyLoad && c.state.valueUpdated {
				if uint64(reursionCount) <= r && !c.state.stateUpdated {
					updateState(c, lazyLoad, r, reursionCount, end, i, index)
				}
				if reursionCount == 1 {
					break
				}

				if i <= end && end+1+c.maxAds >= r-index {
					_index = uint64(index)
				}
				return data
			}

			data[index] = c.slotDurations[i]

			//fmt.Println(c.currentCombinationCount, " :: index =", index, ", recursioncnt = ", reursionCount, ", i=", start, " :: slot = ", data)

			//combinationUtil1(arr, data, i+1, end, index+1, r)
			//      data , start, index, r , merged , lazyLoad
			_index = index - uint64(reursionCount)
			c.search1(data, i, index+1, r, merged, lazyLoad, reursionCount+1)
			//fmt.Println("returned from ", reursionCount+1)
		}
	} else {

		// replace index with all possible elements. The condition
		// "end-i+1 >= r-index" makes sure that including one element
		// at index will make a combination with remaining elements
		// at remaining positions
		for i := start; i <= end && end-i+1 >= r-index; i++ {
			data[index] = c.slotDurations[i]
			//combinationUtil1(arr, data, i+1, end, index+1, r)
			c.search1(data, i+1, index+1, r, merged, lazyLoad, reursionCount+1)

		}
	}

	if lazyLoad && !c.state.stateUpdated {
		c.state.combinationCounter++
		index = _index
		index = uint64(c.state.index) - 1

		//index = r - 2
		//azyLoad , r , reursionCount , end , i                    , index
		updateState(c, lazyLoad, r, reursionCount, end, c.state.combinationCounter, index)
	}
	return data
}

// assuming arr contains unique values
// other wise next elemt will be returned when first matching value of val found
// returns nextValue and its index
func getNextElement(arr []uint64, val uint64) (uint64, uint64) {
	for i, e := range arr {
		if e == val && i+1 < len(arr) {
			return uint64(i) + 1, arr[i+1]
		}
	}
	// assuming durations will never be 0
	return 0, 0
}

func updateState(c *AdSlotDurationCombinations, lazyLoad bool, r uint64, reursionCount int, end uint64, i uint64, index uint64) {

	valueAtEnd := c.slotDurations[end]

	// if lazyLoad && uint64(reursionCount) == r-1 {
	if lazyLoad {

		// set index

		// c.state.start = c.state.combinationCounter

		c.state.start = i
		// set c.state.index = 0 when
		// lastCombination contains, number X len(input) - 1 times starting from last index
		// where X = last number present in the input
		occurance := uint64(0)
		for i := len(c.state.lastCombination) - 1; i >= 0; i-- {
			if c.state.lastCombination[i] == valueAtEnd {
				occurance++
			}
		}
		//c.state.index = int64(c.state.combinationCounter)
		// c.state.index = int64(index)
		c.state.index = int64(index)
		if occurance == r {
			c.state.index = 0
		}

		// set c.state.combinationCounter
		//	c.state.combinationCounter++
		if c.state.combinationCounter >= r {
			// LOGIC : to determine next value
			// 1. get the value P at 0th index present in lastCombination
			// 2. get the index of P
			// 3. determine the next index i.e. index(p) + 1 = q
			// 4. if q == r then set to 0
			diff := (uint64(len(c.state.lastCombination)) - occurance)
			if diff > 0 {
				eleIndex := diff - 1
				c.state.combinationCounter, _ = getNextElement(c.slotDurations, c.state.lastCombination[eleIndex])
				if c.state.combinationCounter == r {
					//			c.state.combinationCounter = 0
				}
				c.state.start = c.state.combinationCounter
			} else {
				fmt.Println("Must be end of r ", r)
			}
		}

		// set r
		// increament value of r if occurance == r
		if occurance == r {
			c.state.start = 0
			c.state.index = 0
			c.state.combinationCounter = 0
			c.state.r++
		}

		c.state.stateUpdated = true
	}
}
