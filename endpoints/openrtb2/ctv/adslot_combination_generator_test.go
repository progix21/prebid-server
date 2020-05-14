package ctv

import (
	"log"
	"testing"
)

var testBidResponseMaxDurations = []struct {
	scenario             string
	responseMaxDurations []int64
	podMinDuration       int64 // Pod Minimum duration value present in origin Video Ad Pod Request
	podMaxDuration       int64 // Pod Maximum duration value present in origin Video Ad Pod Request
	minAds               int64 // Minimum Ads value present in origin Video Ad Pod Request
	maxAds               int64 // Maximum Ads value present in origin Video Ad Pod Request
	combinations         [][]int64
}{
	{
		scenario:             "Single_Value",
		responseMaxDurations: []int64{14},
		podMinDuration:       10, podMaxDuration: 14, minAds: 1, maxAds: 2,
		combinations: [][]int64{{14}}},
	{
		scenario:             "Multi_Value",
		responseMaxDurations: []int64{1, 2, 3, 4, 5},
		podMinDuration:       10, podMaxDuration: 14, minAds: 1, maxAds: 2,
		combinations: [][]int64{{14}}},
	{
		scenario:             "Multi_Value_1",
		responseMaxDurations: []int64{4, 5, 8, 7},
		podMinDuration:       10, podMaxDuration: 14, minAds: 1, maxAds: 5,
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
				c.Next()
				//comb := c.Next()
				//fmt.Println(comb)
			}

		})
	}
}
