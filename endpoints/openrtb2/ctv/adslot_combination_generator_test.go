package ctv

import (
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
	allowRepetitationsForEligibleDurations bool
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
		combinations: [][]int64{{14}}},
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
		podMinDuration:       10, podMaxDuration: 14, minAds: 3, maxAds: 5,
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
		allowRepetitationsForEligibleDurations: false},
	{

		// 4 - c1, c2,    :  5 - c3 : 6 - c4, c5,  8 : c7
		scenario:             "TC8-max_ads (20 ads) > input_bid_durations (no repitations)",
		responseMaxDurations: []uint64{4, 5, 8, 7},
		podMinDuration:       10, podMaxDuration: 14, minAds: 3, maxAds: 20,
		combinations:                           [][]int64{{14}},
		allowRepetitationsForEligibleDurations: false, // no repeitations
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
		allowRepetitationsForEligibleDurations: false,
	},
}

func TestAdSlotCombination(t *testing.T) {
	for _, test := range testBidResponseMaxDurations {
		if test.scenario != "TC6-max_ads (20 ads) > input_bid_durations-repeatation_not_allowed" {
			continue
		}

		t.Run(test.scenario, func(t *testing.T) {
			c := new(AdSlotDurationCombinations)
			d := new(AdSlotDurationCombinations)

			log.Printf("Input = %v", test.responseMaxDurations)

			c.Init(test.podMinDuration, test.podMaxDuration, test.minAds, test.maxAds, test.responseMaxDurations, test.allowRepetitationsForEligibleDurations)
			d.Init(test.podMinDuration, test.podMaxDuration, test.minAds, test.maxAds, test.responseMaxDurations, test.allowRepetitationsForEligibleDurations)

			d.next()
			expectedOutput := d.combinations
			var lazyLoadOutput = make([][]uint64, d.totalExpectedCombinations)
			cnt := 0
			for c.HasNext() {
				//c.Next()
				comb := c.Next()
				//fmt.Print(comb, "\n")
				fmt.Println("e = ", (expectedOutput)[cnt], "\t : a = ", comb)
				val := make([]uint64, len(comb))
				copy(val, comb)
				lazyLoadOutput[cnt] = val

				cnt++
			}

			if expectedOutput != nil {
				// compare results
				for i := uint64(0); i < uint64(len(expectedOutput)); i++ {
					for j := uint64(0); j < uint64(len(expectedOutput[i])); j++ {

						if expectedOutput[i][j] == lazyLoadOutput[i][j] {
						} else {
							assert.Fail(t, "expectedOutput[", i, "][", j, "] != lazyLoadOutput[", i, "][", j, "] ", expectedOutput[i][j], " !=", lazyLoadOutput[i][j])

						}
					}

				}
			}

			//assert.Equal(t, expectedOutput, lazyLoadOutput)
			//assert.ElementsMatch(t, expectedOutput, lazyLoadOutput)

			print("Total combinations generated = %v", c.currentCombinationCount)
			print("Total combinations expected = %v", c.totalExpectedCombinations)
		})
	}
}
