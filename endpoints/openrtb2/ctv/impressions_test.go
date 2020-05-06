package ctv

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
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

var debugOn = true

func TestIsMultipleOf(t *testing.T) {
	if isMultipleOf(5, 6) {
		t.Error("Expected not multiple of")
	}

}

func TestClosedPodMinDuration(t *testing.T) {
	//pod := newTestPod(6, 1, 1, 1, 1, 1)
	pod := newTestPod(1, 90, 11, 15, 2, 8)
	// multipleOf = 5
	cfg, _ := getImpressions(pod.podMinDuration, pod.podMaxDuration, pod.vPod)
	validateClosedMinDuration(t, cfg, 5)
}

func TestCase2(t *testing.T) {
	pod := newTestPod(1, 90, 11, 15, 2, 8)
	expected := Expected{
		impressionCount:       6,
		freeTime:              0.0,
		adSlotTimeInSec:       []int64{15, 15, 15, 15, 15, 15},
		output:                [][2]int64{{15, 15}, {15, 15}, {15, 15}, {15, 15}, {15, 15}, {15, 15}},
		closedMinDuration:     5,
		closedMaxDuration:     90,
		closedSlotMinDuration: 15,
		closedSlotMaxDuration: 15,
	}
	pod.test(t, expected)
}

func TestCase3(t *testing.T) {
	pod := newTestPod(1, 90, 11, 15, 2, 4)
	expected := Expected{
		impressionCount: 4,
		freeTime:        30.0,
		adSlotTimeInSec: []int64{15, 15, 15, 15},
		output:          [][2]int64{{15, 15}, {15, 15}, {15, 15}, {15, 15}},

		closedMinDuration:     5,
		closedMaxDuration:     90,
		closedSlotMinDuration: 15,
		closedSlotMaxDuration: 15,
	}
	pod.test(t, expected)
}

func TestCase4(t *testing.T) {
	pod := newTestPod(1, 15, 1, 15, 1, 1)
	expected := Expected{
		impressionCount: 1,
		freeTime:        0.0,
		adSlotTimeInSec: []int64{15},
		output:          [][2]int64{{15, 15}},

		closedMinDuration:     5,
		closedMaxDuration:     15,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 15,
	}
	pod.test(t, expected)
}

func TestCase5(t *testing.T) {
	pod := newTestPod(1, 15, 1, 15, 1, 2)
	expected := Expected{
		impressionCount: 2,
		freeTime:        0.0,
		adSlotTimeInSec: []int64{10, 5},
		output:          [][2]int64{{10, 10}, {5, 5}},

		closedMinDuration:     5,
		closedMaxDuration:     15,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 15,
	}
	pod.test(t, expected)
}

func TestCase6(t *testing.T) {
	pod := newTestPod(1, 90, 1, 15, 1, 8)
	expected := Expected{
		impressionCount: 8,
		freeTime:        0.0,
		adSlotTimeInSec: []int64{15, 15, 10, 10, 10, 10, 10, 10},
		output:          [][2]int64{{15, 15}, {15, 15}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}},

		closedMinDuration:     5,
		closedMaxDuration:     90,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 15,
	}
	pod.test(t, expected)
}

func TestCase7(t *testing.T) {
	pod := newTestPod(15, 30, 8, 15, 1, 1)
	expected := Expected{
		impressionCount: 1,
		freeTime:        15.0,
		adSlotTimeInSec: []int64{15},
		output:          [][2]int64{{15, 15}},

		closedMinDuration:     15,
		closedMaxDuration:     30,
		closedSlotMinDuration: 10,
		closedSlotMaxDuration: 15,
	}
	pod.test(t, expected)
}

func TestCase8(t *testing.T) {
	pod := newTestPod(35, 35, 10, 35, 3, 40)
	expected := Expected{
		impressionCount: 3,
		freeTime:        0.0,
		adSlotTimeInSec: []int64{15, 10, 10},
		output:          [][2]int64{{15, 15}, {10, 10}, {10, 10}},

		closedMinDuration:     35,
		closedMaxDuration:     35,
		closedSlotMinDuration: 10,
		closedSlotMaxDuration: 35,
	}
	pod.test(t, expected)
}

func TestCase9(t *testing.T) {
	pod := newTestPod(35, 35, 10, 35, 6, 40)
	expected := Expected{
		impressionCount: 0,
		freeTime:        pod.podMaxDuration,
		adSlotTimeInSec: []int64{},
		output:          [][2]int64{},

		closedMinDuration:     35,
		closedMaxDuration:     35,
		closedSlotMinDuration: 10,
		closedSlotMaxDuration: 35,
	}
	pod.test(t, expected)
}

func TestCase10(t *testing.T) {
	pod := newTestPod(35, 65, 10, 35, 6, 40)
	expected := Expected{
		impressionCount: 6,
		freeTime:        0.0,
		adSlotTimeInSec: []int64{15, 10, 10, 10, 10, 10},
		output:          [][2]int64{{15, 15}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}},

		closedMinDuration:     35,
		closedMaxDuration:     65,
		closedSlotMinDuration: 10,
		closedSlotMaxDuration: 35,
	}
	pod.test(t, expected)
}

func TestCase11(t *testing.T) {
	pod := newTestPod(35, 65, 9, 35, 7, 40)
	expected := Expected{
		impressionCount: 0, //7,
		freeTime:        pod.podMaxDuration,
		adSlotTimeInSec: []int64{}, // []int64{10, 10, 10, 10, 10, 10, 5},
		output:          [][2]int64{},

		closedMinDuration:     35,
		closedMaxDuration:     65,
		closedSlotMinDuration: 10,
		closedSlotMaxDuration: 35,
	}
	pod.test(t, expected)
}

func TestCase12(t *testing.T) {
	pod := newTestPod(100, 100, 10, 35, 6, 40)
	expected := Expected{
		impressionCount: 10,
		freeTime:        0.0,
		adSlotTimeInSec: []int64{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
		output:          [][2]int64{{10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}},

		closedMinDuration:     100,
		closedMaxDuration:     100,
		closedSlotMinDuration: 10,
		closedSlotMaxDuration: 35,
	}
	pod.test(t, expected)
}

func TestCase13(t *testing.T) {
	pod := newTestPod(60, 60, 5, 9, 1, 6)
	expected := Expected{
		impressionCount: 0,
		freeTime:        pod.podMaxDuration,
		adSlotTimeInSec: []int64{},
		output:          [][2]int64{},

		closedMinDuration:     60,
		closedMaxDuration:     60,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 5,
	}
	pod.test(t, expected)
}

func TestCase14(t *testing.T) {
	pod := newTestPod(30, 60, 5, 9, 1, 6)
	expected := Expected{
		impressionCount: 6,
		freeTime:        30,
		adSlotTimeInSec: []int64{5, 5, 5, 5, 5, 5},
		output:          [][2]int64{{5, 5}, {5, 5}, {5, 5}, {5, 5}, {5, 5}, {5, 5}},

		closedMinDuration:     30,
		closedMaxDuration:     60,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 5,
	}

	pod.test(t, expected)

}

func TestCase15(t *testing.T) {
	pod := newTestPod(30, 60, 5, 9, 1, 5)
	expected := Expected{
		impressionCount: 0,
		freeTime:        pod.podMaxDuration,
		adSlotTimeInSec: []int64{},
		output:          [][2]int64{},

		closedMinDuration:     30,
		closedMaxDuration:     60,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 5,
	}
	pod.test(t, expected)
}

func TestCase16(t *testing.T) {
	pod := newTestPod(126, 126, 1, 12, 7, 13)
	expected := Expected{
		impressionCount: 13,
		freeTime:        0,
		adSlotTimeInSec: []int64{10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 6},
		output:          [][2]int64{{10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {1, 6}},

		closedMinDuration:     125,
		closedMaxDuration:     125,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 10,
	}
	pod.test(t, expected)
}

func TestCase17(t *testing.T) {
	pod := newTestPod(127, 128, 1, 12, 7, 13)
	expected := Expected{
		impressionCount: 13,
		freeTime:        0,
		adSlotTimeInSec: []int64{10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 8},
		output:          [][2]int64{{10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {1, 8}},

		closedMinDuration:     130,
		closedMaxDuration:     125,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 10,
	}
	pod.test(t, expected)
}

func TestCase18(t *testing.T) {
	pod := newTestPod(125, 125, 4, 4, 1, 1)
	expected := Expected{
		impressionCount: 0,
		freeTime:        pod.podMaxDuration,
		adSlotTimeInSec: []int64{},
		output:          [][2]int64{},

		closedMinDuration:     125,
		closedMaxDuration:     125,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 4,
	}
	pod.test(t, expected)
}

func TestCase19(t *testing.T) {
	pod := newTestPod(90, 90, 7, 9, 3, 5)
	expected := Expected{
		impressionCount: 0,
		freeTime:        pod.podMaxDuration,
		adSlotTimeInSec: []int64{}, // 90 -25 = 65
		output:          [][2]int64{},

		closedMinDuration:     90,
		closedMaxDuration:     90,
		closedSlotMinDuration: 10,
		closedSlotMaxDuration: 5,
	}
	pod.test(t, expected)
}

func TestCase20(t *testing.T) {
	pod := newTestPod(90, 90, 5, 10, 1, 11)
	expected := Expected{
		impressionCount: 9,
		freeTime:        0,
		adSlotTimeInSec: []int64{10, 10, 10, 10, 10, 10, 10, 10, 10},
		output:          [][2]int64{{10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}},

		closedMinDuration:     90,
		closedMaxDuration:     90,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 10,
	}
	pod.test(t, expected)
}

func TestCase23(t *testing.T) {
	pod := newTestPod(118, 124, 4, 17, 6, 15)
	expected := Expected{
		impressionCount: 12,
		freeTime:        0,
		adSlotTimeInSec: []int64{14, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
		output:          [][2]int64{{4, 14}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}},

		closedMinDuration:     120,
		closedMaxDuration:     120,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 15,

		// 24,8  => 24 => 15
		// 120 / 15 = 8 => 10
	}
	pod.test(t, expected)
}

func TestCase24(t *testing.T) {
	pod := newTestPod(134, 134, 60, 90, 2, 3)
	expected := Expected{
		impressionCount: 2,
		freeTime:        0,
		adSlotTimeInSec: []int64{69, 65},
		output:          [][2]int64{{69, 69}, {65, 65}},

		closedMinDuration:     135,
		closedMaxDuration:     135,
		closedSlotMinDuration: 60,
		closedSlotMaxDuration: 90,
	}
	pod.test(t, expected)
}

// Test case when only video min and max duration is passed
func TestCase26(t *testing.T) {
	pod := newTestPod(90, 90, 45, 45, 2, 3)
	expected := Expected{
		impressionCount:       2,
		freeTime:              0,
		adSlotTimeInSec:       []int64{45, 45},
		output:                [][2]int64{{45, 45}, {45, 45}},
		closedMinDuration:     90,
		closedMaxDuration:     90,
		closedSlotMinDuration: 45,
		closedSlotMaxDuration: 45,
	}
	pod.test(t, expected)
}

func TestCase27(t *testing.T) {
	pod := newTestPod(5, 90, 2, 45, 2, 3)
	expected := Expected{
		impressionCount:       3,
		freeTime:              0,
		adSlotTimeInSec:       []int64{30, 30, 30},
		output:                [][2]int64{{30, 30}, {30, 30}, {30, 30}},
		closedMinDuration:     5,
		closedMaxDuration:     90,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 45,
	}
	pod.test(t, expected)
}

func TestCase28(t *testing.T) {
	pod := newTestPod(5, 180, 2, 90, 2, 6)
	expected := Expected{
		impressionCount:       6,
		freeTime:              0,
		adSlotTimeInSec:       []int64{30, 30, 30, 30, 30, 30},
		output:                [][2]int64{{30, 30}, {30, 30}, {30, 30}, {30, 30}, {30, 30}, {30, 30}},
		closedMinDuration:     5,
		closedMaxDuration:     180,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 90,
	}
	pod.test(t, expected)
}

func TestCase29(t *testing.T) {
	pod := newTestPod(5, 65, 2, 35, 2, 3)
	expected := Expected{
		impressionCount: 3,
		freeTime:        0,
		adSlotTimeInSec: []int64{25, 20, 20},
		output:          [][2]int64{{25, 25}, {20, 20}, {20, 20}},

		closedMinDuration:     5,
		closedMaxDuration:     65,
		closedSlotMinDuration: 5,
		closedSlotMaxDuration: 35,
	}
	pod.test(t, expected)
}

func (p TestAdPod) test(t *testing.T, expected Expected) {
	log.SetOutput(ioutil.Discard)
	pod, _ := getImpressions(p.podMinDuration, p.podMaxDuration, p.vPod)
	log.Println("")
	validateImpressionCount(t, pod, expected.impressionCount)
	validateTimeForEachAdSlot(t, pod, expected.adSlotTimeInSec)
	validateFreeTime(t, pod, expected.freeTime)
	// validate closed bounds
	validateClosedMinDuration(t, pod, expected.closedMinDuration)
	validateClosedMaxDuration(t, pod, expected.closedMaxDuration)
	validateClosedSlotMinDuration(t, pod, expected.closedSlotMinDuration)
	validateClosedSlotMaxDuration(t, pod, expected.closedSlotMaxDuration)
	validate2dArrayOutput(t, pod, expected.output)
}

func validate2dArrayOutput(t *testing.T, pod AdPodConfig, expectedOutput [][2]int64) {
	if len(pod.Slots) != len(expectedOutput) {
		t.Errorf("Expecte Number of Ad Slots %v  . But Found %v", len(expectedOutput), len(pod.Slots))
		return
	} else {
		if debugOn {
			log.Printf("** Got Number of Ad Slots = %v\n", len(pod.Slots))
		}
	}

	// check each output
	for i := 0; i < len(expectedOutput); i++ {

		if expectedOutput[i][0] != pod.Slots[i][0] {
			t.Errorf("Expected Min Duration for Ad Slot %v = %v  . But Found %v", i, expectedOutput[i][0], pod.Slots[i][0])
		} else {
			if debugOn {
				log.Printf("** Got Min Duration for Ad Slot %v = %v \n", i, pod.Slots[i][0])
			}
		}

		if expectedOutput[i][1] != pod.Slots[i][1] {
			t.Errorf("Expected Min Duration for Ad Slot %v = %v  . But Found %v", i, expectedOutput[i][1], pod.Slots[i][1])
		} else {
			if debugOn {
				log.Printf("** Got Min Duration for Ad Slot %v = %v \n", i, pod.Slots[i][1])
			}
		}

	}
}

func validateTimeForEachAdSlot(t *testing.T, pod AdPodConfig, expectedAdSlotTimeInSec []int64) {
	if len(pod.Slots) != len(expectedAdSlotTimeInSec) {
		t.Errorf("Expected Number of Ad Slots %v  . But Found %v", len(expectedAdSlotTimeInSec), len(pod.Slots))

	} else {
		if debugOn {
			log.Printf("** Got Number of Ad Slots = %v\n", len(pod.Slots))
		}
	}
	for i := 0; i < len(pod.Slots); i++ {
		if pod.Slots[i][1] != expectedAdSlotTimeInSec[i] {
			t.Errorf("Expected Slot time for Ad Slot %v[1] = %v . But Found %v", i, expectedAdSlotTimeInSec[i], (pod.Slots)[i][1])
		} else {
			if debugOn {
				log.Printf("** Got Expected Slot time for Ad Slot = %v\n", (pod.Slots)[i])
			}
		}
	}
}

func validateImpressionCount(t *testing.T, pod AdPodConfig, expectedImpressionCount int) {
	if !(len(pod.Slots) == expectedImpressionCount) {
		t.Errorf("Expected impression count = %v . But Found %v", expectedImpressionCount, len(pod.Slots))
	} else {
		if debugOn {
			log.Printf("** Got Expected impression count = %v\n", len(pod.Slots))
		}
	}
}

func validateFreeTime(t *testing.T, pod AdPodConfig, expectedFreeTime int64) {
	if pod.FreeTime != expectedFreeTime {
		t.Errorf("Expected Free Time = %v . But Found %v", expectedFreeTime, pod.FreeTime)
	} else {
		if debugOn {
			log.Printf("** Got Expected Free Time = %v\n", pod.FreeTime)
		}
	}
}

func validateClosedMinDuration(t *testing.T, pod AdPodConfig, expectedClosedMinDuration int64) {
	if pod.PodMinDuration != expectedClosedMinDuration {
		t.Errorf("Expected closedMinDuration= %v . But Found %v", expectedClosedMinDuration, pod.PodMinDuration)
	} else {
		if debugOn {
			log.Printf("** Got Expected closedMinDuration = %v\n", pod.PodMinDuration)
		}
	}
}

func validateClosedMaxDuration(t *testing.T, pod AdPodConfig, expectedClosedMaxDuration int64) {
	if pod.PodMaxDuration != expectedClosedMaxDuration {
		t.Errorf("Expected closedMinDuration= %v . But Found %v", expectedClosedMaxDuration, pod.PodMaxDuration)
	} else {
		if debugOn {
			log.Printf("** Got Expected closedMinDuration = %v\n", pod.PodMaxDuration)
		}
	}
}

func validateClosedSlotMinDuration(t *testing.T, pod AdPodConfig, expectedClosedSlotMinDuration int64) {
	if pod.SlotMinDuration != expectedClosedSlotMinDuration {
		t.Errorf("Expected closedSlotMinDuration= %v . But Found %v", expectedClosedSlotMinDuration, pod.SlotMinDuration)
	} else {
		if debugOn {
			log.Printf("** Got Expected closedSlotMinDuration = %v\n", pod.SlotMinDuration)
		}
	}
}

func validateClosedSlotMaxDuration(t *testing.T, pod AdPodConfig, expectedClosedSlotMaxDuration int64) {
	if pod.SlotMaxDuration != expectedClosedSlotMaxDuration {
		t.Errorf("Expected closedSlotMinDuration= %v . But Found %v", expectedClosedSlotMaxDuration, pod.SlotMaxDuration)
	} else {
		if debugOn {
			log.Printf("** Got Expected closedSlotMinDuration = %v\n", pod.SlotMaxDuration)
		}
	}
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

/* Benchmarking Tests */
func BenchmarkTestCase2(b *testing.B) {
	p := newTestPod(1, 90, 11, 15, 2, 8)
	for n := 0; n < b.N; n++ {
		getImpressions(p.podMinDuration, p.podMaxDuration, p.vPod)
	}
}
