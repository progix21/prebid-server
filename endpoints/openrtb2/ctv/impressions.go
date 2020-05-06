// Package ctv provides functionalities for handling CTV specific Request  and responses
package ctv

import (
	"log"
	"math"

	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
)

// AdPodConfig contains Pod Minimum Duration, Pod Maximum Duration, Slot Minimum Duration and Slot Maximum Duration
// It holds additional attributes required by this algorithm for  internal computation.
// 	It contains Slots attribute. This  attribute holds the output of this algorithm
type AdPodConfig struct {
	MinAds          int64 // Minimum number of Ads / Slots allowed inside Ad Pod
	MaxAds          int64 // Maximum number of Ads / Slots allowed inside Ad Pod.
	SlotMinDuration int64 // Minimum duration (in seconds) for each Ad Slot inside Ad Pod. It is not original value from request. It holds the value closed to original value and multiples of X.
	SlotMaxDuration int64 // Maximum duration (in seconds) for each Ad Slot inside Ad Pod. It is not original value from request. It holds the value closed to original value and multiples of X.
	PodMinDuration  int64 // Minimum total duration (in seconds) of Ad Pod. It is not original value from request. It holds the value closed to original value and multiples of X.
	PodMaxDuration  int64 // Maximum total duration (in seconds) of Ad Pod. It is not original value from request. It holds the value closed to original value and multiples of X.

	RequestedPodMinDuration  int64      // Requested Ad Pod minimum duration (in seconds)
	RequestedPodMaxDuration  int64      // Requested Ad Pod maximum duration (in seconds)
	RequestedSlotMinDuration int64      // Requested Ad Slot minimum duration (in seconds)
	Slots                    [][2]int64 // Holds Minimum and Maximum duration (in seconds) for each Ad Slot. Length indicates total number of Ad Slots/ Impressions for given Ad Pod
	TotalSlotTime            *int64     // Total Sum of all Ad Slot durations (in seconds)
	FreeTime                 int64      // Remaining Time (in seconds) not allocated. It is compared with RequestedPodMaxDuration
}

// Value use to compute Ad Slot Durations and Pod Durations for internal computation
// Right now this value is set to 5, based on passed data observations
// Observed that typically video impression contains contains minimum and maximum duration in multiples of  5
var multipleOf = int64(5)

// Constucts the AdPodConfig object from openrtb_ext.VideoAdPod
// It computes durations for Ad Slot and Ad Pod in multiple of X
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

// Returns the number of Ad Slots/Impression  that input Ad Pod can have.
// It also returs Minimum and  Maximum duration. Dimension 1, represents Minimum duration. Dimension 2, represents Maximum Duration
// for each Ad Slot.
// Minimum Duratiuon can contain either RequestedSlotMinDuration or Duration computed by algorithm for the Ad Slot
// Maximum Duration only contains Duration computed by algorithm for the Ad Slot
// podMinDuration - Minimum duration of Pod, podMaxDuration Maximum duration of Pod, vPod Video Pod Object
func getImpressions(podMinDuration, podMaxDuration int64, vPod openrtb_ext.VideoAdPod) (AdPodConfig, [][2]int64) {

	cfg := init0(podMinDuration, podMaxDuration, vPod)
	totalAds := computeTotalAds(cfg)
	timeForEachSlot := computeTimeForEachAdSlot(cfg, totalAds)

	log.Printf("Pod Config (x5) = %+v\n", cfg)
	log.Println("totalAds =", totalAds)
	log.Println("timeForEachSlot = ", timeForEachSlot)

	cfg.Slots = make([][2]int64, totalAds)
	// iterate over total time till it is < cfg.RequestedPodMaxDuration
	time := int64(0)
	for time < cfg.RequestedPodMaxDuration {
		adjustedTime, slotsFull := cfg.addTime(timeForEachSlot)
		time += adjustedTime
		timeForEachSlot = computeTimeLeastValue(cfg.RequestedPodMaxDuration - time)
		if slotsFull {
			log.Println("All slots are full of their capacity. validating slots")
			break
		}
	}

	// validate slots
	cfg.validateSlots()

	// log free time if present to stats server
	// also check algoritm computed the no. of ads
	if cfg.RequestedPodMaxDuration-time > 0 && len(cfg.Slots) > 0 {
		cfg.FreeTime = cfg.RequestedPodMaxDuration - time
		log.Println("TO STATS SERVER : Free Time not allocated ", cfg.FreeTime, "sec")
	}

	log.Printf("\nTotal Impressions = %v, Total Allocated Time = %v sec (out of %v sec, Max Pod Duration)\n%v", len(cfg.Slots), *cfg.TotalSlotTime, cfg.RequestedPodMaxDuration, cfg.Slots)
	return cfg, cfg.Slots
}

// Returns total number of Ad Slots/ impressions that the Ad Pod can have
func computeTotalAds(cfg AdPodConfig) int64 {
	maxAds := cfg.PodMaxDuration / cfg.SlotMaxDuration
	minAds := cfg.PodMaxDuration / cfg.SlotMinDuration

	totalAds := max(minAds, maxAds)

	if totalAds < cfg.MinAds {
		totalAds = cfg.MinAds
	}
	if totalAds > cfg.MaxAds {
		totalAds = cfg.MaxAds
	}
	return totalAds
}

// Returns duration in seconds that can be allocated to each Ad Slot
// Accepts cfg containing algorithm configurations and totalAds containing Total number of
// Ad Slots / Impressions that the Ad Pod can have.
func computeTimeForEachAdSlot(cfg AdPodConfig, totalAds int64) int64 {
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
	return timeForEachSlot
}

// Checks if multipleOf can be used as least time value
// this will ensure eack slot to maximize its time if possible
// if multipleOf can not be used as least value then default input value is returned as is
// accepts time containing, which least value to be computed.
// Returns the least value based on multiple of X
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

// Validate the algorithm computations
//  1. Verifies if 2D slice containing Min duration and Max duration values are non-zero
//  2. Idenfies the Ad Slots / Impressions with either Min Duration or Max Duration or both
//     having zero value and removes it from 2D slice
//  3. Ensures  Minimum Pod duration <= TotalSlotTime <= Maximum Pod Duration
// if  any validation fails it removes all the alloated slots and  makes is of size 0
// and sets the FreeTime value as RequestedPodMaxDuration
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
			log.Printf("WARNING:Slot %v %v is having 0  duration\n", index, slot)
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
		log.Printf("ERROR: slotSize %v is either less than Min Ads (%v) or greater than Max Ads (%v)\n", len(config.Slots), config.MinAds, config.MaxAds)
		returnEmptySlots = true
	}

	// ensure if min pod duration = max pod duration
	// config.TotalSlotTime = pod duration
	if config.RequestedPodMinDuration == config.RequestedPodMaxDuration && *config.TotalSlotTime != config.RequestedPodMaxDuration {
		log.Printf("ERROR: Total Slot Duration %v sec is not matching with Total Pod Duration %v sec\n", *config.TotalSlotTime, config.RequestedPodMaxDuration)
		returnEmptySlots = true
	}

	// ensure slot duration lies between requested min pod duration and  requested max pod duration
	// Testcase #15
	if *config.TotalSlotTime < config.RequestedPodMinDuration || *config.TotalSlotTime > config.RequestedPodMaxDuration {
		log.Printf("ERROR: Total Slot Duration %v sec is either less than Requested Pod Min Duration (%v sec) or greater than Requested  Pod Max Duration (%v sec)\n", *config.TotalSlotTime, config.RequestedPodMinDuration, config.RequestedPodMaxDuration)
		returnEmptySlots = true
	}

	if returnEmptySlots {
		config.Slots = emptySlots
		config.FreeTime = config.RequestedPodMaxDuration
	}
}

// Adds time to possible slots and returns total added time
//
// Checks following for each Ad Slot
//  1. Can Ad Slot adjust the input time
//  2. If addition of new time to any slot not exeeding Total Pod Max Duration
// Performs the following operations
//  1. Populates Minimum duration slot[][0] - Either Slot Minimum Duration or Actual Slot Time computed
//  2. Populates Maximum duration slot[][1] - Always actual Slot Time computed
//  3. Counts the number of Ad Slots / Impressons full with  duration  capacity. If all Ad Slots / Impressions
//     are full of capacity it returns true as second return argument, indicating all slots are full with capacity
//  4. Keeps track of TotalSlotDuration when each new time is added to the Ad Slot
//  5. Keeps track of difference between computed PodMaxDuration and RequestedPodMaxDuration (TestCase #16) and used in step #2 above
// Returns argument 1 indicating total time adusted, argument 2 whether all slots are full of duration capacity
func (config AdPodConfig) addTime(timeForEachSlot int64) (int64, bool) {
	time := int64(0)

	// iterate over each ad
	slotCountFullWithCapacity := 0
	for ad := int64(0); ad < int64(len(config.Slots)); ad++ {

		slot := &config.Slots[ad]
		// check
		// 1. time(slot(0)) <= config.SlotMaxDuration
		// 2. if adding new time  to slot0 not exeeding config.SlotMaxDuration
		// 3. if sum(slot time) +  timeForEachSlot  <= config.RequestedPodMaxDuration
		canAdjustTime := (slot[0] + timeForEachSlot) <= config.SlotMaxDuration
		totalSlotTimeWithNewTimeLessThanRequestedPodMaxDuration := *config.TotalSlotTime+timeForEachSlot <= config.RequestedPodMaxDuration
		maxPodDurationMatchUpTime := config.RequestedPodMaxDuration - config.PodMaxDuration
		if slot[0] <= config.SlotMaxDuration && canAdjustTime && totalSlotTimeWithNewTimeLessThanRequestedPodMaxDuration {
			slot[0] += timeForEachSlot

			// if we are adjusting the free time which will match up with config.RequestedPodMaxDuration
			// then set config.SlotMinDuration as min value for this slot
			// TestCase #16
			if timeForEachSlot == maxPodDurationMatchUpTime {
				// override existing value of slot[0] here
				slot[0] = config.RequestedSlotMinDuration
			}

			slot[1] += timeForEachSlot
			*config.TotalSlotTime += timeForEachSlot
			time += timeForEachSlot
			log.Printf("Slot %v = Added %v sec (New Time = %v)\n", ad, timeForEachSlot, slot[1])
		}
		// check slot capabity
		// !canAdjustTime - TestCase18
		if slot[1] == config.SlotMaxDuration || !canAdjustTime {
			// slot is full
			slotCountFullWithCapacity++
		}
	}
	log.Println("adjustedTime = ", time)
	return time, slotCountFullWithCapacity == len(config.Slots)
}

// Returns Maximum number out off 2 input numbers
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

// Returns true if num is multipleof second argument. False otherwise
func isMultipleOf(num, multipleOf int64) bool {
	return math.Mod(float64(num), float64(multipleOf)) == 0
}

// Returns closet factor for num, with  respect  input multipleOf
//  Example: Closet Factor of 9, in multiples of 5 is '10'
func getClosetFactor(num, multipleOf int64) int64 {
	return int64(math.Round(float64(num)/float64(multipleOf)) * float64(multipleOf))
}

// Returns closetfactor of MinDuration, with  respect to multipleOf
// If computed factor < MinDuration then it will ensure and return
// close factor >=  MinDuration
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

// Returns closetfactor of maxduration, with  respect to multipleOf
// If computed factor > maxduration then it will ensure and return
// close factor <=  maxduration
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
