package ctv

import (
	"testing"

	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

type TestAdPod struct {
	vPod           openrtb_ext.VideoAdPod
	podMinDuration int64
	podMaxDuration int64
}

type Expected struct {
	impressionCount int
	// Time remaining after ad breaking is done
	// if no ad breaking i.e. 0 then freeTime = pod.maxduration
	freeTime        int64
	adSlotTimeInSec []int64

	// close bounds
	closedMinDuration     int64 // pod
	closedMaxDuration     int64 // pod
	closedSlotMinDuration int64 // ad slot
	closedSlotMaxDuration int64 // ad slot

	output [][2]int64
}

var impressionsTests = []struct {
	scenario string   // Testcase scenario
	in       []int    // Testcase input
	out      Expected // Testcase execpted output
}{
	{scenario: "TC2", in: []int{1, 90, 11, 15, 2, 8}, out: Expected{
		impressionCount:       6,
		freeTime:              0.0,
		adSlotTimeInSec:       []int64{15, 15, 15, 15, 15, 15},
		output:                [][2]int64{{15, 15}, {15, 15}, {15, 15}, {15, 15}, {15, 15}, {15, 15}},
		closedMinDuration:     5,
		closedMaxDuration:     90,
		closedSlotMinDuration: 15,
		closedSlotMaxDuration: 15,
	}},
	{scenario: "TC3", in: []int{1, 90, 11, 15, 2, 4}, out: Expected{
		impressionCount: 4,
		freeTime:        30.0,
		adSlotTimeInSec: []int64{15, 15, 15, 15},
		output:          [][2]int64{{15, 15}, {15, 15}, {15, 15}, {15, 15}},

		closedMinDuration:     5,
		closedMaxDuration:     90,
		closedSlotMinDuration: 15,
		closedSlotMaxDuration: 15,
	}},
	{scenario: "TC4", in: []int{1, 15, 1, 15, 1, 1}, out: Expected{
		impressionCount: 1,
		freeTime:        0.0,
		adSlotTimeInSec: []int64{15},
		output:          [][2]int64{{15, 15}},

		closedMinDuration:     5,
		closedMaxDuration:     15,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 15,
	}},
	{scenario: "TC5", in: []int{1, 15, 1, 15, 1, 2}, out: Expected{
		impressionCount: 2,
		freeTime:        0.0,
		adSlotTimeInSec: []int64{10, 5},
		output:          [][2]int64{{10, 10}, {5, 5}},

		closedMinDuration:     5,
		closedMaxDuration:     15,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 15,
	}},
	{scenario: "TC6", in: []int{1, 90, 1, 15, 1, 8}, out: Expected{
		impressionCount: 8,
		freeTime:        0.0,
		adSlotTimeInSec: []int64{15, 15, 10, 10, 10, 10, 10, 10},
		output:          [][2]int64{{15, 15}, {15, 15}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}},

		closedMinDuration:     5,
		closedMaxDuration:     90,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 15,
	}},
	{scenario: "TC7", in: []int{15, 30, 8, 15, 1, 1}, out: Expected{
		impressionCount: 1,
		freeTime:        15.0,
		adSlotTimeInSec: []int64{15},
		output:          [][2]int64{{15, 15}},

		closedMinDuration:     15,
		closedMaxDuration:     30,
		closedSlotMinDuration: 10,
		closedSlotMaxDuration: 15,
	}},
	{scenario: "TC8", in: []int{35, 35, 10, 35, 3, 40}, out: Expected{
		impressionCount: 3,
		freeTime:        0.0,
		adSlotTimeInSec: []int64{15, 10, 10},
		output:          [][2]int64{{15, 15}, {10, 10}, {10, 10}},

		closedMinDuration:     35,
		closedMaxDuration:     35,
		closedSlotMinDuration: 10,
		closedSlotMaxDuration: 35,
	}},
	{scenario: "TC9", in: []int{35, 35, 10, 35, 6, 40}, out: Expected{
		impressionCount: 0,
		freeTime:        35,
		adSlotTimeInSec: []int64{},
		output:          [][2]int64{},

		closedMinDuration:     35,
		closedMaxDuration:     35,
		closedSlotMinDuration: 10,
		closedSlotMaxDuration: 35,
	}},
	{scenario: "TC10", in: []int{35, 65, 10, 35, 6, 40}, out: Expected{
		impressionCount: 6,
		freeTime:        0.0,
		adSlotTimeInSec: []int64{15, 10, 10, 10, 10, 10},
		output:          [][2]int64{{15, 15}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}},

		closedMinDuration:     35,
		closedMaxDuration:     65,
		closedSlotMinDuration: 10,
		closedSlotMaxDuration: 35,
	}},
	{scenario: "TC11", in: []int{35, 65, 9, 35, 7, 40}, out: Expected{
		impressionCount: 0, //7,
		freeTime:        65,
		adSlotTimeInSec: []int64{}, // []int64{10, 10, 10, 10, 10, 10, 5},
		output:          [][2]int64{},

		closedMinDuration:     35,
		closedMaxDuration:     65,
		closedSlotMinDuration: 10,
		closedSlotMaxDuration: 35,
	}},
	{scenario: "TC12", in: []int{100, 100, 10, 35, 6, 40}, out: Expected{
		impressionCount: 10,
		freeTime:        0.0,
		adSlotTimeInSec: []int64{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
		output:          [][2]int64{{10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}},

		closedMinDuration:     100,
		closedMaxDuration:     100,
		closedSlotMinDuration: 10,
		closedSlotMaxDuration: 35,
	}},
	{scenario: "TC13", in: []int{60, 60, 5, 9, 1, 6}, out: Expected{
		impressionCount: 0,
		freeTime:        60,
		adSlotTimeInSec: []int64{},
		output:          [][2]int64{},

		closedMinDuration:     60,
		closedMaxDuration:     60,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 5,
	}},
	{scenario: "TC14", in: []int{30, 60, 5, 9, 1, 6}, out: Expected{
		impressionCount: 6,
		freeTime:        30,
		adSlotTimeInSec: []int64{5, 5, 5, 5, 5, 5},
		output:          [][2]int64{{5, 5}, {5, 5}, {5, 5}, {5, 5}, {5, 5}, {5, 5}},

		closedMinDuration:     30,
		closedMaxDuration:     60,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 5,
	}},
	{scenario: "TC15", in: []int{30, 60, 5, 9, 1, 5}, out: Expected{
		impressionCount: 0,
		freeTime:        60,
		adSlotTimeInSec: []int64{},
		output:          [][2]int64{},

		closedMinDuration:     30,
		closedMaxDuration:     60,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 5,
	}},
	{scenario: "TC16", in: []int{126, 126, 1, 12, 7, 13}, out: Expected{
		impressionCount: 13,
		freeTime:        0,
		adSlotTimeInSec: []int64{10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 6},
		output:          [][2]int64{{10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {1, 6}},

		closedMinDuration:     125,
		closedMaxDuration:     125,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 10,
	}},
	{scenario: "TC17", in: []int{127, 128, 1, 12, 7, 13}, out: Expected{
		impressionCount: 13,
		freeTime:        0,
		adSlotTimeInSec: []int64{10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 8},
		output:          [][2]int64{{10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {1, 8}},

		closedMinDuration:     130,
		closedMaxDuration:     125,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 10,
	}},
	{scenario: "TC18", in: []int{125, 125, 4, 4, 1, 1}, out: Expected{
		impressionCount: 0,
		freeTime:        125,
		adSlotTimeInSec: []int64{},
		output:          [][2]int64{},

		closedMinDuration:     125,
		closedMaxDuration:     125,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 4,
	}},
	{scenario: "TC19", in: []int{90, 90, 7, 9, 3, 5}, out: Expected{
		impressionCount: 0,
		freeTime:        90,
		adSlotTimeInSec: []int64{}, // 90 -25 = 65
		output:          [][2]int64{},

		closedMinDuration:     90,
		closedMaxDuration:     90,
		closedSlotMinDuration: 10,
		closedSlotMaxDuration: 5,
	}},
	{scenario: "TC20", in: []int{90, 90, 5, 10, 1, 11}, out: Expected{
		impressionCount: 9,
		freeTime:        0,
		adSlotTimeInSec: []int64{10, 10, 10, 10, 10, 10, 10, 10, 10},
		output:          [][2]int64{{10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}},

		closedMinDuration:     90,
		closedMaxDuration:     90,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 10,
	}},
	{scenario: "TC23", in: []int{118, 124, 4, 17, 6, 15}, out: Expected{
		impressionCount: 12,
		freeTime:        0,
		adSlotTimeInSec: []int64{14, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
		output:          [][2]int64{{4, 14}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}},

		closedMinDuration:     120,
		closedMaxDuration:     120,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 15,
	}},
	{scenario: "TC24", in: []int{134, 134, 60, 90, 2, 3}, out: Expected{
		impressionCount: 2,
		freeTime:        0,
		adSlotTimeInSec: []int64{69, 65},
		output:          [][2]int64{{69, 69}, {65, 65}},

		closedMinDuration:     135,
		closedMaxDuration:     135,
		closedSlotMinDuration: 60,
		closedSlotMaxDuration: 90,
	}},
	{scenario: "TC26", in: []int{90, 90, 45, 45, 2, 3}, out: Expected{
		impressionCount:       2,
		freeTime:              0,
		adSlotTimeInSec:       []int64{45, 45},
		output:                [][2]int64{{45, 45}, {45, 45}},
		closedMinDuration:     90,
		closedMaxDuration:     90,
		closedSlotMinDuration: 45,
		closedSlotMaxDuration: 45,
	}},
	{scenario: "TC27", in: []int{5, 90, 2, 45, 2, 3}, out: Expected{
		impressionCount:       3,
		freeTime:              0,
		adSlotTimeInSec:       []int64{30, 30, 30},
		output:                [][2]int64{{30, 30}, {30, 30}, {30, 30}},
		closedMinDuration:     5,
		closedMaxDuration:     90,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 45,
	}},
	{scenario: "TC28", in: []int{5, 180, 2, 90, 2, 6}, out: Expected{
		impressionCount:       6,
		freeTime:              0,
		adSlotTimeInSec:       []int64{30, 30, 30, 30, 30, 30},
		output:                [][2]int64{{30, 30}, {30, 30}, {30, 30}, {30, 30}, {30, 30}, {30, 30}},
		closedMinDuration:     5,
		closedMaxDuration:     180,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 90,
	}},
	{scenario: "TC29", in: []int{5, 65, 2, 35, 2, 3}, out: Expected{
		impressionCount: 3,
		freeTime:        0,
		adSlotTimeInSec: []int64{25, 20, 20},
		output:          [][2]int64{{25, 25}, {20, 20}, {20, 20}},

		closedMinDuration:     5,
		closedMaxDuration:     65,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 35,
	}},
}

func TestGetImpressions(t *testing.T) {
	for _, impTest := range impressionsTests {
		if impTest.scenario != "TC9" {
			continue
		}
		t.Run(impTest.scenario, func(t *testing.T) {
			p := newTestPod(int64(impTest.in[0]), int64(impTest.in[1]), impTest.in[2], impTest.in[3], impTest.in[4], impTest.in[5])
			cfg, _ := getImpressions(p.podMinDuration, p.podMaxDuration, p.vPod)
			expected := impTest.out

			validateImpressionCount(t, cfg, expected.impressionCount)
			validateTimeForEachAdSlot(t, cfg, expected.adSlotTimeInSec)
			validateFreeTime(t, cfg, expected.freeTime)
			// // validate closed bounds
			validateClosedMinDuration(t, cfg, expected.closedMinDuration)
			validateClosedMaxDuration(t, cfg, expected.closedMaxDuration)
			validateClosedSlotMinDuration(t, cfg, expected.closedSlotMinDuration)
			validateClosedSlotMaxDuration(t, cfg, expected.closedSlotMaxDuration)
			validate2dArrayOutput(t, cfg, expected.output)
		})
	}
}

/* Benchmarking Tests */
func BenchmarkGetImpressions(b *testing.B) {
	for _, impTest := range impressionsTests {
		b.Run(impTest.scenario, func(b *testing.B) {
			p := newTestPod(int64(impTest.in[0]), int64(impTest.in[1]), impTest.in[2], impTest.in[3], impTest.in[4], impTest.in[5])
			for n := 0; n < b.N; n++ {
				getImpressions(p.podMinDuration, p.podMaxDuration, p.vPod)
			}
		})
	}
}

func validate2dArrayOutput(t *testing.T, pod adPodConfig, expectedOutput [][2]int64) {
	assert.Equal(t, len(expectedOutput), len(pod.Slots), "Expected Number of Ad Slots %v  . But Found %v", len(expectedOutput), len(pod.Slots))
	// check each output
	for i := 0; i < len(expectedOutput); i++ {
		assert.Equal(t, expectedOutput[i][0], pod.Slots[i][0], "Expected Min Duration for Ad Slot %v = %v  . But Found %v", i, expectedOutput[i][0], pod.Slots[i][0])
		assert.Equal(t, expectedOutput[i][1], pod.Slots[i][1], "Expected Min Duration for Ad Slot %v = %v  . But Found %v", i, expectedOutput[i][1], pod.Slots[i][1])
	}
}

func validateTimeForEachAdSlot(t *testing.T, pod adPodConfig, expectedAdSlotTimeInSec []int64) {
	assert.Equal(t, len(expectedAdSlotTimeInSec), len(pod.Slots), "Expected Number of Ad Slots %v  . But Found %v", len(expectedAdSlotTimeInSec), len(pod.Slots))
	for i := 0; i < len(pod.Slots); i++ {
		assert.Equal(t, expectedAdSlotTimeInSec[i], pod.Slots[i][1], "Expected Slot time for Ad Slot %v[1] = %v . But Found %v", i, expectedAdSlotTimeInSec[i], (pod.Slots)[i][1])
	}
}

func validateImpressionCount(t *testing.T, pod adPodConfig, expectedImpressionCount int) {
	assert.Equal(t, expectedImpressionCount, len(pod.Slots), "Expected impression count = %v . But Found %v", expectedImpressionCount, len(pod.Slots))
}

func validateFreeTime(t *testing.T, pod adPodConfig, expectedFreeTime int64) {
	assert.Equal(t, expectedFreeTime, pod.freeTime, "Expected Free Time = %v . But Found %v", expectedFreeTime, pod.freeTime)
}

func validateClosedMinDuration(t *testing.T, pod adPodConfig, expectedClosedMinDuration int64) {
	assert.Equal(t, expectedClosedMinDuration, pod.podMinDuration, "Expected closedMinDuration= %v . But Found %v", expectedClosedMinDuration, pod.podMinDuration)
}

func validateClosedMaxDuration(t *testing.T, pod adPodConfig, expectedClosedMaxDuration int64) {
	assert.Equal(t, expectedClosedMaxDuration, pod.podMaxDuration, "Expected closedMinDuration= %v . But Found %v", expectedClosedMaxDuration, pod.podMaxDuration)
}

func validateClosedSlotMinDuration(t *testing.T, pod adPodConfig, expectedClosedSlotMinDuration int64) {
	assert.Equal(t, expectedClosedSlotMinDuration, pod.slotMinDuration, "Expected closedSlotMinDuration= %v . But Found %v", expectedClosedSlotMinDuration, pod.slotMinDuration)
}

func validateClosedSlotMaxDuration(t *testing.T, pod adPodConfig, expectedClosedSlotMaxDuration int64) {
	assert.Equal(t, expectedClosedSlotMaxDuration, pod.slotMaxDuration, "Expected closedSlotMinDuration= %v . But Found %v", expectedClosedSlotMaxDuration, pod.slotMaxDuration)
}

func newTestPod(podMinDuration, podMaxDuration int64, slotMinDuration, slotMaxDuration, minAds, maxAds int) *TestAdPod {
	testPod := TestAdPod{}

	pod := openrtb_ext.VideoAdPod{}

	pod.MinDuration = &slotMinDuration
	pod.MaxDuration = &slotMaxDuration
	pod.MinAds = &minAds
	pod.MaxAds = &maxAds

	testPod.vPod = pod
	testPod.podMinDuration = podMinDuration
	testPod.podMaxDuration = podMaxDuration
	return &testPod
}
