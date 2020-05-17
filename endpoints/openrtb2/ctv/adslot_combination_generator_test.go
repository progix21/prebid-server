package ctv

import (
	"fmt"
	"log"
	"testing"
)

var testBidResponseMaxDurations = []struct {
	scenario             string
	responseMaxDurations []uint64
	podMinDuration       int64 // Pod Minimum duration value present in origin Video Ad Pod Request
	podMaxDuration       int64 // Pod Maximum duration value present in origin Video Ad Pod Request
	minAds               int64 // Minimum Ads value present in origin Video Ad Pod Request
	maxAds               int64 // Maximum Ads value present in origin Video Ad Pod Request
	combinations         [][]int64
}{
	{
		scenario:             "Single_Value",
		responseMaxDurations: []uint64{14},
		podMinDuration:       10, podMaxDuration: 14, minAds: 1, maxAds: 2,
		combinations: [][]int64{{14}}},
	{
		scenario:             "Multi_Value",
		responseMaxDurations: []uint64{1, 2, 3, 4, 5},
		podMinDuration:       10, podMaxDuration: 14, minAds: 1, maxAds: 2,
		combinations: [][]int64{{14}}},
	{

		// 4 - c1, c2,    :  5 - c3 : 6 - c4, c5,  8 : c7
		scenario:             "Multi_Value_1",
		responseMaxDurations: []uint64{4, 5, 8, 7},
		podMinDuration:       10, podMaxDuration: 14, minAds: 3, maxAds: 4,
		combinations: [][]int64{{14}}},
}

func TestAdSlotCombination(t *testing.T) {
	for _, test := range testBidResponseMaxDurations {
		if test.scenario != "Multi_Value_1" {
			continue
		}

		t.Run(test.scenario, func(t *testing.T) {
			c := new(AdSlotDurationCombinations)
			c.Init(test.podMinDuration, test.podMaxDuration, test.minAds, test.maxAds, test.responseMaxDurations)
			log.Printf("Input = %v", test.responseMaxDurations)
			for c.HasNext() {
				//c.Next()
				comb := c.Next()
				fmt.Print(comb, "\n")
			}

			print("Total combinations generated = %v", c.currentCombinationCount)
			print("Total combinations expected = %v", c.totalExpectedCombinations)
		})
	}
}
