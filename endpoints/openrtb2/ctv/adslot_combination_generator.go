package ctv

//AdSlotDurationCombinations holds all the combinations based
//on Video Ad Pod request and Bid Response Max duration
type AdSlotDurationCombinations struct {
	podMinDuration int64 // Pod Minimum duration value present in origin Video Ad Pod Request
	podMaxDuration int64 // Pod Maximum duration value present in origin Video Ad Pod Request
	minAds         int64 // Minimum Ads value present in origin Video Ad Pod Request
	maxAds         int64 // Maximum Ads value present in origin Video Ad Pod Request

	slotDurations []int64 // input slot durations for which

	combinations [][]int64 // May contains some/all combinations at given point of time
}

//Next - Get next ad slot combination
//returns empty array if next combination is not present
func (c AdSlotDurationCombinations) Next() []int64 {
	return nil
}

func (c AdSlotDurationCombinations) next() []int64 {

}

// HasNext - true if next combination is present
// false if not
func (comb AdSlotDurationCombinations) HasNext() bool {
	return false
}
