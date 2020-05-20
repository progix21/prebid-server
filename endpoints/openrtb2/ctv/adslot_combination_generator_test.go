package ctv

import (
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testBidResponseMaxDurations = []struct {
	scenario                               string
	responseMaxDurations                   []uint64
	podMinDuration                         int64 // Pod Minimum duration value present in origin Video Ad Pod Request
	podMaxDuration                         int64 // Pod Maximum duration value present in origin Video Ad Pod Request
	minAds                                 int64 // Minimum Ads value present in origin Video Ad Pod Request
	maxAds                                 int64 // Maximum Ads value present in origin Video Ad Pod Request
	combinations                           [][]int64
	allowRepetitationsForEligibleDurations string
}{
	{
		scenario:             "TC1-Single_Value",
		responseMaxDurations: []uint64{14},
		podMinDuration:       10, podMaxDuration: 14, minAds: 1, maxAds: 2,
		combinations: [][]int64{{14}}},
	{
		scenario:             "TC2-Multi_Value",
		responseMaxDurations: []uint64{1, 2, 3, 4, 5},
		podMinDuration:       10, podMaxDuration: 14, minAds: 1, maxAds: 2,
		combinations: [][]int64{{14}}},
	{

		// 4 - c1, c2,    :  5 - c3 : 6 - c4, c5,  8 : c7
		scenario:             "TC3-max_ads = input_bid_durations",
		responseMaxDurations: []uint64{4, 5, 8, 7},
		podMinDuration:       10, podMaxDuration: 14, minAds: 3, maxAds: 4,
		combinations: [][]int64{{14}}, allowRepetitationsForEligibleDurations: "true"},
	{

		// 4 - c1, c2,    :  5 - c3 : 6 - c4, c5,  8 : c7
		scenario:             "TC4-max_ads < input_bid_durations (test 1)",
		responseMaxDurations: []uint64{4, 5, 8, 7},
		podMinDuration:       10, podMaxDuration: 14, minAds: 3, maxAds: 3,
		combinations: [][]int64{{14}}},
	{

		// 4 - c1, c2,    :  5 - c3 : 6 - c4, c5,  8 : c7
		scenario:             "TC5-max_ads  (1) < input_bid_durations (test 1)",
		responseMaxDurations: []uint64{4, 5, 8, 7},
		podMinDuration:       10, podMaxDuration: 14, minAds: 3, maxAds: 1,
		combinations: [][]int64{{14}}},

	{

		// 4 - c1, c2,    :  5 - c3 : 6 - c4, c5,  8 : c7
		scenario:             "TC6-max_ads < input_bid_durations (test 2)",
		responseMaxDurations: []uint64{4, 5, 8, 7},
		podMinDuration:       10, podMaxDuration: 14, minAds: 3, maxAds: 2,
		combinations: [][]int64{{14}}},
	{

		// 4 - c1, c2,    :  5 - c3 : 6 - c4, c5,  8 : c7
		scenario:             "TC7-max_ads > input_bid_durations (test 1)",
		responseMaxDurations: []uint64{4, 5, 8, 7},
		podMinDuration:       10, podMaxDuration: 14, minAds: 4, maxAds: 4,
		combinations: [][]int64{{14}}},
	{

		// 4 - c1, c2,    :  5 - c3 : 6 - c4, c5,  8 : c7
		scenario:             "TC8-max_ads (20 ads) > input_bid_durations (test 2)",
		responseMaxDurations: []uint64{4, 5, 8, 7},
		podMinDuration:       10, podMaxDuration: 14, minAds: 3, maxAds: 20,
		combinations: [][]int64{{14}}},
	{

		// 4 - c1, c2,    :  5 - c3 : 6 - c4, c5,  8 : c7
		scenario:             "TC6-max_ads (20 ads) > input_bid_durations-repeatation_not_allowed",
		responseMaxDurations: []uint64{4, 5, 8, 7},
		podMinDuration:       10, podMaxDuration: 14, minAds: 3, maxAds: 2,
		combinations:                           [][]int64{{14}},
		allowRepetitationsForEligibleDurations: "false"},
	{

		// 4 - c1, c2,    :  5 - c3 : 6 - c4, c5,  8 : c7
		scenario:             "TC8-max_ads (20 ads) > input_bid_durations (no repitations)",
		responseMaxDurations: []uint64{4, 5, 8, 7},
		podMinDuration:       10, podMaxDuration: 14, minAds: 3, maxAds: 20,
		combinations:                           [][]int64{{14}},
		allowRepetitationsForEligibleDurations: "false", // no repeitations
	},

	{

		// 4 - c1, c2,    :  5 - c3 : 6 - c4, c5,  8 : c7
		scenario:             "TC9-max_ads = input_bid_durations = 4",
		responseMaxDurations: []uint64{4, 4, 4, 4},
		podMinDuration:       10, podMaxDuration: 14, minAds: 3, maxAds: 4,
		combinations: [][]int64{{14}}},
	{

		// 4 - c1, c2,    :  5 - c3 : 6 - c4, c5,  8 : c7
		scenario:             "TC10-max_ads 0",
		responseMaxDurations: []uint64{4, 4, 4, 4},
		podMinDuration:       10, podMaxDuration: 14, minAds: 3, maxAds: 0,
		combinations: [][]int64{{14}}},
	{

		// 4 - c1, c2,    :  5 - c3 : 6 - c4, c5,  8 : c7
		scenario:             "TC11-max_ads =5-input-empty",
		responseMaxDurations: []uint64{},
		podMinDuration:       10, podMaxDuration: 14, minAds: 3, maxAds: 0,
		combinations: [][]int64{{14}}},

	{

		// 4 - c1, c2,    :  5 - c3 : 6 - c4, c5,  8 : c7
		scenario:             "TC12-max_ads =5-input-empty-no-repeatation",
		responseMaxDurations: []uint64{25, 30, 76, 10, 88},
		podMinDuration:       10, podMaxDuration: 14, minAds: 3, maxAds: 4,
		combinations:                           [][]int64{{14}},
		allowRepetitationsForEligibleDurations: "false",
	}, {

		// 4 - c1, c2,    :  5 - c3 : 6 - c4, c5,  8 : c7
		scenario:             "TC13-max_ads = input = 10-without-repeatation",
		responseMaxDurations: []uint64{25, 30, 76, 10, 88, 34, 37, 67, 89, 45},
		podMinDuration:       10, podMaxDuration: 14, minAds: 3, maxAds: 10,
		combinations:                           [][]int64{{14}},
		allowRepetitationsForEligibleDurations: "true",
	},
}

func TestBFSgenerator(t *testing.T) {
	input := []uint64{4, 5, 8, 7}

	baseCombValueIndex := 0
	baseCombValueIndexValueProvider := 0
	//groupCounter := 0
	duration := 0
	minads := 4
	var previousCombination []uint64
	for combCnt := 0; combCnt < 35; {
		baseCombination := make([]uint64, 1)
		baseCombination[0] = input[baseCombValueIndex]
		for ; duration < len(input); duration++ {

			newCombination := generateNexCombination(previousCombination, minads, input, baseCombValueIndex)
			combCnt++
			print("Comb Count = %v    -> %v ", combCnt, newCombination)
			previousCombination = newCombination

		}

		if baseCombValueIndex == len(input)-1 {
			// if baseCombValueIndexValueProvider >= len(input) {
			// 	baseCombValueIndexValueProvider = baseCombValueIndex - baseCombValueIndexValueProvider
			// }
			//

			// determine  first value after 7
			for i := len(input) - 1; i >= 0; i-- {
				if previousCombination[i] != input[len(input)-1] {
					index, _, err := getNextInputValue(previousCombination[i], input, baseCombValueIndex)
					if err == nil {
						baseCombValueIndex = index
						break
					}
				}
			}

			baseCombValueIndexValueProvider++
			baseCombValueIndex = baseCombValueIndexValueProvider

		} else {
			baseCombValueIndex++
		}
		// if baseCombValueIndex >= len(input) {
		// 	baseCombValueIndexValueProvider++
		// 	if baseCombValueIndexValueProvider >= len(input) {
		// 		baseCombValueIndexValueProvider = baseCombValueIndex - baseCombValueIndexValueProvider
		// 	}
		// 	baseCombValueIndex = baseCombValueIndexValueProvider
		// }
		duration = baseCombValueIndex
	}

}

// generate combination for specific level
func generateNexCombination(previousCombination []uint64, minAds int, inputDurations []uint64, baseCombValueIndex int) []uint64 {

	newCombination := make([]uint64, minAds)

	if previousCombination == nil {
		// init with base 0
		for i := 0; i < minAds; i++ {
			newCombination[i] = inputDurations[0]
		}
		return newCombination
	}
	// lastinput duration received  in the request
	endDurationValueFromInputRequest := inputDurations[len(inputDurations)-1]
	for index, prvCombDuration := range previousCombination {

		// check if value (index+1) next to prvCombDuration contains endDurationValueFromInputRequest
		var nextDuration uint64
		nextDurationSet := false
		if (index + 1) < len(previousCombination) {
			nextDuration = previousCombination[index+1]
			nextDurationSet = true
		}

		if nextDurationSet && nextDuration == endDurationValueFromInputRequest {
			// set new value next to last value hold by previsois combination

			// get last value from previous combination using same index
			lastValue := previousCombination[index]

			// if previousCombination[index] = nextDuration = endDurationValueFromInputRequest
			// then set value = newCombination[i-1]
			// for example
			// 4,7,7,7
			// 5,5,5,5

			if previousCombination[index] == nextDuration {
				newCombination[index] = newCombination[index-1]
			} else {
				// determine index of last value in original request
				_, nextValue, err := getNextInputValue(lastValue, inputDurations, baseCombValueIndex)
				if err == nil {
					newCombination[index] = nextValue
				}
			}

		} else {
			// default assign same duration received
			newCombination[index] = prvCombDuration
			if index == minAds-1 {
				_, nextVal, err := getNextInputValue(prvCombDuration, inputDurations, baseCombValueIndex)
				if err == nil {
					newCombination[minAds-1] = nextVal
				}
			}

		}
	}
	return newCombination
}

func getNextInputValue(lastValue uint64, inputDurations []uint64, baseCombValueIndex int) (int, uint64, error) {
	// get last value from previous combination using same index

	// determine index of last value in original request
	for i := 0; i < len(inputDurations); i++ {
		if lastValue == inputDurations[i] {
			// put next value as in same index of new combination
			// no need to check boundary conditions for next value
			// caller must take care of it
			if i+1 >= len(inputDurations) {
				return baseCombValueIndex, inputDurations[baseCombValueIndex], nil
			}
			return i + 1, inputDurations[i+1], nil
		}
	}

	return 0, 0, errors.New("no next input value found")
}

func combinationUtil(arr, data []int, start, end,
	index, r, n int) {
	// Current combination is ready to be printed, print it
	if index == r {
		for j := 0; j < r; j++ {
			print("%v ", data[j])
			println()
			return
		}

		// replace index with all possible elements. The condition
		// "end-i+1 >= r-index" makes sure that including one element
		// at index will make a combination with remaining elements
		// at remaining positions
		for i := start; i <= end && end-i+1 >= r-index; i++ {
			data[index] = arr[i]
			combinationUtil(arr, data, i+1, end, index+1, r, n)
		}
	}
}

func TestCombOnline(t *testing.T) {
	input := []int{4, 5, 8, 7}
	data := make([]int, 100)
	combinationUtil(input, data, 0, len(input), 2, 2, 4)
}

func TestAdSlotCombination(t *testing.T) {
	for _, test := range testBidResponseMaxDurations {

		if test.scenario != "TC13-max_ads = input = 10-without-repeatation" {
			continue
		}

		t.Run(test.scenario, func(t *testing.T) {
			c := new(AdSlotDurationCombinations)
			d := new(AdSlotDurationCombinations)

			log.Printf("Input = %v", test.responseMaxDurations)
			allowRepetitationsForEligibleDurations := true
			if test.allowRepetitationsForEligibleDurations == "false" {
				allowRepetitationsForEligibleDurations = false
			}

			c.Init(test.podMinDuration, test.podMaxDuration, test.minAds, test.maxAds, test.responseMaxDurations, allowRepetitationsForEligibleDurations)
			d.Init(test.podMinDuration, test.podMaxDuration, test.minAds, test.maxAds, test.responseMaxDurations, allowRepetitationsForEligibleDurations)

			//d.next()
			// expectedOutput := d.combinations

			expectedOutput := c.search1tr()
			actualOutput := make([][]uint64, len(expectedOutput))

			cnt := 0
			for c.HasNext() {
				//c.Next()
				comb := c.search1trlazy()
				fmt.Print(comb, "\n")
				//	fmt.Println("e = ", (expectedOutput)[cnt], "\t : a = ", comb)
				val := make([]uint64, len(comb))
				copy(val, comb)
				actualOutput[cnt] = val
				cnt++
			}

			if expectedOutput != nil {
				// compare results
				for i := uint64(0); i < uint64(len(expectedOutput)); i++ {
					for j := uint64(0); j < uint64(len(expectedOutput[i])); j++ {

						if expectedOutput[i][j] == actualOutput[i][j] {
						} else {

							assert.Fail(t, "expectedOutput[", i, "][", j, "] != actualOutput[", i, "][", j, "] ", string(expectedOutput[i][j]), " !=", string(actualOutput[i][j]))

						}
					}

				}
			}

			// assert.Equal(t, expectedOutput, lazyLoadOutput)
			// assert.ElementsMatch(t, expectedOutput, lazyLoadOutput)

			print("Total combinations generated = %v", c.currentCombinationCount)
			print("Total combinations expected = %v", c.totalExpectedCombinations)
		})
	}
}
