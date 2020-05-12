package ctv

import (
	"testing"

	"github.com/influxdata/influxdb/pkg/testing/assert"
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
}

func TestAdSlotCombination(t *testing.T) {
	for _, testBidReseponseMaxDuration := range testBidResponseMaxDurations {
		expectedCombinations := testBidReseponseMaxDuration.combinations
		t.Run(testBidReseponseMaxDuration.scenario, func(t *testing.T) {
			slotDurations := AdSlotDurationCombinations{}
			assert.Equal(t, len(slotDurations.combinations), len(expectedCombinations))
			for i := 0; i < len(expectedCombinations); i++ {
				if slotDurations.HasNext() {
					assert.Equal(t, slotDurations.Next(), expectedCombinations[i])
				}
			}
		})
	}
}
