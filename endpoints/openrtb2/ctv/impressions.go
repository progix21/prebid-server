package ctv

import (
	"fmt"
	"math"

	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
)

//Configuration for Algorithm
type AdPodConfig struct {
	MinAds          int64 `json:"minads,omitempty"`      //Default 1 if not specified
	MaxAds          int64 `json:"maxads,omitempty"`      //Default 1 if not specified
	SlotMinDuration int64 `json:"MinDuration,omitempty"` // adpod.MinDuration*adpod.minads should be greater than or equal to video.MinDuration
	SlotMaxDuration int64 `json:"maxduration,omitempty"` // adpod.maxduration*adpod.maxads should be less than or equal to video.maxduration + video.maxextended
	PodMinDuration  int64
	PodMaxDuration  int64

	RequestedPodMinDuration  int64 // requested pod min duration
	RequestedPodMaxDuration  int64 // requested pod max duration
	RequestedSlotMinDuration int64 // request slot mid duration
	Slots                    [][2]int64
	TotalSlotTime            *int64 // sum all of slot time
	FreeTime                 int64  // remaining time (not allocated)
}

var multipleOf = int64(5)

func init0(podMinDuration, podMaxDuration int64, vPod openrtb_ext.VideoAdPod) AdPodConfig {
	config := AdPodConfig{}

	config.RequestedPodMinDuration = podMinDuration
	config.RequestedPodMaxDuration = podMaxDuration
	config.RequestedSlotMinDuration = int64(*vPod.MinDuration)
	if config.RequestedPodMinDuration == config.RequestedPodMaxDuration {
		/*TestCase 16*/
		config.PodMinDuration = getClosetFactor(config.RequestedPodMinDuration, multipleOf)
		config.PodMaxDuration = config.PodMinDuration
	} else {
		config.PodMinDuration = getClosetFactorForMinDuration(config.RequestedPodMinDuration, multipleOf)
		config.PodMaxDuration = getClosetFactorForMaxDuration(config.RequestedPodMaxDuration, multipleOf)
	}
	config.SlotMinDuration = getClosetFactorForMinDuration(int64(config.RequestedSlotMinDuration), multipleOf)
	config.SlotMaxDuration = getClosetFactorForMaxDuration(int64(*vPod.MaxDuration), multipleOf)
	config.MinAds = int64(*vPod.MinAds)
	config.MaxAds = int64(*vPod.MaxAds)

	config.TotalSlotTime = new(int64)
	return config
}

// 1, 90, 11, 15, 2, 8
func getImpressionObjectsv2(podMinDuration, podMaxDuration int64, vPod openrtb_ext.VideoAdPod) AdPodConfig {

	cfg := init0(podMinDuration, podMaxDuration, vPod)

	// Compute no of ads
	// pod max durationo / min ad duration and max ad duration
	maxAds := cfg.PodMaxDuration / cfg.SlotMaxDuration
	minAds := cfg.PodMaxDuration / cfg.SlotMinDuration

	totalAds := max(minAds, maxAds)

	if totalAds < cfg.MinAds {
		totalAds = cfg.MinAds
	}
	if totalAds > cfg.MaxAds {
		totalAds = cfg.MaxAds
	}

	// Compute time for each ad
	timeForEachSlot := cfg.PodMaxDuration / totalAds

	if timeForEachSlot < cfg.SlotMinDuration {
		timeForEachSlot = cfg.SlotMinDuration
	}

	if timeForEachSlot > cfg.SlotMaxDuration {
		timeForEachSlot = cfg.SlotMaxDuration
	}

	// ensure timeForEachSlot is multipleof given number
	if !isMultipleOf(timeForEachSlot, multipleOf) {
		// get close to value of multiple
		// here we muse get either cfg.SlotMinDuration or cfg.SlotMaxDuration
		// these values are already pre-computed in multiples of given number
		timeForEachSlot = getClosetFactor(timeForEachSlot, multipleOf)
	}

	fmt.Printf("Pod Config (x5) = %+v\n", cfg)
	fmt.Println("totalAds =", totalAds)
	fmt.Println("timeForEachSlot = ", timeForEachSlot)

	cfg.Slots = make([][2]int64, totalAds)
	// iterate over total time till it is < cfg.RequestedPodMaxDuration
	time := int64(0)
	for time < cfg.RequestedPodMaxDuration {
		adjustedTime, slotsFull := cfg.addTime(timeForEachSlot)
		time += adjustedTime
		timeForEachSlot = computeTimeLeastValue(cfg.RequestedPodMaxDuration - time)
		if slotsFull {
			fmt.Println("All slots are full of their capacity. validating slots")
			break
		}
	}

	// validate slots
	cfg.validateSlots()

	// log it free time if present to stats server
	// also check algoritm computed the no. of ads
	if cfg.RequestedPodMaxDuration-time > 0 && len(cfg.Slots) > 0 {
		cfg.FreeTime = cfg.RequestedPodMaxDuration - time
		fmt.Println("TO STATS SERVER : Free Time not allocated ", cfg.FreeTime, "sec")
	}

	fmt.Printf("\nTotal Impressions = %v, Total Allocated Time = %v sec (out of %v sec, Max Pod Duration)\n%v", len(cfg.Slots), *cfg.TotalSlotTime, cfg.RequestedPodMaxDuration, cfg.Slots)
	return cfg
}

// checks if multipleOf can be used as least time value
// this will ensure eack slot to maximize its time if possible
// if multipleOf can not be used as least value then default input value is returned as is
func computeTimeLeastValue(time int64) int64 {
	// time if Testcase#6
	// 1. multiple of x - get smallest factor N of multiple of x for time
	// 2. not multiple of x - try to obtain smallet no N multipe of x
	// ensure N <= timeForEachSlot
	leastFactor := multipleOf
	if leastFactor < time {
		time = leastFactor
	}
	return time
}

func (config *AdPodConfig) validateSlots() {

	// default return value if validation fails
	emptySlots := make([][2]int64, 0)
	if len(config.Slots) == 0 {
		return
	}

	// check slot with 0 values
	// remove them from config.Slots
	emptySlotCount := 0
	for index, slot := range config.Slots {
		if slot[0] == 0 || slot[1] == 0 {
			fmt.Printf("WARNING:Slot %v %v is having 0  duration\n", index, slot)
			emptySlotCount++
		}
	}

	// remove empty slot
	optimizedSlots := make([][2]int64, len(config.Slots)-emptySlotCount)
	for index, slot := range config.Slots {
		if slot[0] == 0 || slot[1] == 0 {
		} else {
			optimizedSlots[index][0] = slot[0]
			optimizedSlots[index][1] = slot[1]
		}
	}
	config.Slots = optimizedSlots

	returnEmptySlots := false
	if int64(len(config.Slots)) < config.MinAds || int64(len(config.Slots)) > config.MaxAds {
		fmt.Printf("ERROR: slotSize %v is either less than Min Ads (%v) or greater than Max Ads (%v)\n", len(config.Slots), config.MinAds, config.MaxAds)
		returnEmptySlots = true
	}

	// ensure if min pod duration = max pod duration
	// config.TotalSlotTime = pod duration
	if config.RequestedPodMinDuration == config.RequestedPodMaxDuration && *config.TotalSlotTime != config.RequestedPodMaxDuration {
		fmt.Printf("ERROR: Total Slot Duration %v sec is not matching with Total Pod Duration %v sec\n", *config.TotalSlotTime, config.RequestedPodMaxDuration)
		returnEmptySlots = true
	}

	// ensure slot duration lies between requested min pod duration and  requested max pod duration
	// Testcase #15
	if *config.TotalSlotTime < config.RequestedPodMinDuration || *config.TotalSlotTime > config.RequestedPodMaxDuration {
		fmt.Printf("ERROR: Total Slot Duration %v sec is either less than Requested Pod Min Duration (%v sec) or greater than Requested  Pod Max Duration (%v sec)\n", *config.TotalSlotTime, config.RequestedPodMinDuration, config.RequestedPodMaxDuration)
		returnEmptySlots = true
	}

	if returnEmptySlots {
		config.Slots = emptySlots
		config.FreeTime = config.RequestedPodMaxDuration
	}
}

// add time to possible slots and returns total added time
func (cfg AdPodConfig) addTime(timeForEachSlot int64) (int64, bool) {
	time := int64(0)

	// iterate over each ad
	slotCountFullWithCapacity := 0
	for ad := int64(0); ad < int64(len(cfg.Slots)); ad++ {

		slot := &cfg.Slots[ad]
		// check
		// 1. time(slot(0)) <= cfg.SlotMaxDuration
		// 2. if adding new time  to slot0 not exeeding cfg.SlotMaxDuration
		// 3. if sum(slot time) +  timeForEachSlot  <= cfg.RequestedPodMaxDuration
		canAdjustTime := (slot[0] + timeForEachSlot) <= cfg.SlotMaxDuration
		totalSlotTimeWithNewTimeLessThanRequestedPodMaxDuration := *cfg.TotalSlotTime+timeForEachSlot <= cfg.RequestedPodMaxDuration
		maxPodDurationMatchUpTime := cfg.RequestedPodMaxDuration - cfg.PodMaxDuration
		if slot[0] <= cfg.SlotMaxDuration && canAdjustTime && totalSlotTimeWithNewTimeLessThanRequestedPodMaxDuration {
			slot[0] += timeForEachSlot

			// if we are adjusting the free time which will match up with cfg.RequestedPodMaxDuration
			// then set cfg.SlotMinDuration as min value for this slot
			// TestCase #16
			if timeForEachSlot == maxPodDurationMatchUpTime {
				// override existing value of slot[0] here
				slot[0] = cfg.RequestedSlotMinDuration
			}

			slot[1] += timeForEachSlot
			*cfg.TotalSlotTime += timeForEachSlot
			time += timeForEachSlot
			fmt.Printf("Slot %v = Added %v sec (New Time = %v)\n", ad, timeForEachSlot, slot[1])
		}
		// check slot capabity
		// !canAdjustTime - TestCase18
		if slot[1] == cfg.SlotMaxDuration || !canAdjustTime {
			// slot is full
			slotCountFullWithCapacity++
		}
	}
	fmt.Println("adjustedTime = ", time)
	return time, slotCountFullWithCapacity == len(cfg.Slots)
}

func max(num1, num2 int64) int64 {

	if num1 > num2 {
		return num1
	}

	if num2 > num1 {
		return num2
	}
	// both must be equal here
	return num1
}

func isMultipleOf(num, multipleOf int64) bool {
	return math.Mod(float64(num), float64(multipleOf)) == 0
}

func getClosetFactor(num, multipleOf int64) int64 {
	return int64(math.Round(float64(num)/float64(multipleOf)) * float64(multipleOf))
}

func getClosetFactorForMinDuration(MinDuration int64, multipleOf int64) int64 {
	closedMinDuration := getClosetFactor(MinDuration, multipleOf)

	if closedMinDuration == 0 {
		return multipleOf
	}

	if closedMinDuration == MinDuration {
		return MinDuration
	}

	if closedMinDuration < MinDuration {
		return closedMinDuration + multipleOf
	}

	return closedMinDuration
}

func getClosetFactorForMaxDuration(maxduration, multipleOf int64) int64 {
	closedMaxDuration := getClosetFactor(maxduration, multipleOf)
	if closedMaxDuration == maxduration {
		return maxduration
	}

	// set closet maxduration closed to masduration
	for i := closedMaxDuration; i <= maxduration; {
		if closedMaxDuration < maxduration {
			closedMaxDuration = i + multipleOf
			i = closedMaxDuration
		}
	}

	if closedMaxDuration > maxduration {
		duration := closedMaxDuration - multipleOf
		if duration == 0 {
			// return input value as is instead of zero to avoid NPE
			return maxduration
		}
		return duration
	}

	return closedMaxDuration
}

func abs(num int64) int64 {
	return int64(math.Abs(float64(num)))
}

func check(podMin, podMax int64, slotMin, slotMax, minAd, maxAd int) {
	fmt.Println("IP = ", podMin, podMax, slotMin, slotMax, minAd, maxAd)
	// 1, 90, 11, 15, 2, 8
	cfg := openrtb_ext.VideoAdPod{}

	// slot min duratio
	cfg.MinDuration = new(int)
	*cfg.MinDuration = slotMin

	// slot max duration
	cfg.MaxDuration = new(int)
	*cfg.MaxDuration = slotMax

	cfg.MinAds = new(int)
	*cfg.MinAds = minAd
	cfg.MaxAds = new(int)
	*cfg.MaxAds = maxAd
	getImpressionObjectsv2(podMin, podMax, cfg)
}

func main() {
	//check(1, 90, 11, 15, 2, 8)
	//check(134, 134, 60, 90, 2, 3) // OLD OP = 70  + 64  + (0)          = 134 sec
	check(126, 126, 1, 12, 7, 13) // OLD OP = 10  + 10  + 10  + 10  + 10  + 10  + 10  + 10  + 10  + 10  + 10  + 10  + 6  + (0) = 126 sec
	//check(127, 128, 1, 12, 7, 13)
	//check(5, 65, 2, 35, 2, 3) //#29
	//check(35, 35, 10, 35, 6, 40) //#9
	//check(1, 15, 1, 13, 2, 2)
	//check(2, 170, 3, 9, 4, 9)
	//check(35, 65, 10, 35, 6, 40)

	//check(1000, 1000, 15, 35, 4, 5) // exact pod duration
}
