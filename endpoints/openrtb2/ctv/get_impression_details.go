package ctv

import (
	"fmt"
	"math"
	"strconv"

	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
)

//AdPod
type AdPodConfig struct {
	MinAds          int64 `json:"minads,omitempty"`      //Default 1 if not specified
	MaxAds          int64 `json:"maxads,omitempty"`      //Default 1 if not specified
	SlotMinDuration int64 `json:"MinDuration,omitempty"` // adpod.MinDuration*adpod.minads should be greater than or equal to video.MinDuration
	SlotMaxDuration int64 `json:"maxduration,omitempty"` // adpod.maxduration*adpod.maxads should be less than or equal to video.maxduration + video.maxextended
	MinDuration     int64
	MaxDuration     int64

	slots                 *[]int64
	imps                  [][2]int64
	impCnt                int64
	freeTime              float64
	ClosedSlotMinDuration int64
	ClosedSlotMaxDuration int64
	ClosedMinDuration     float64
	ClosedMaxDuration     float64
	// indicates whether there are one/more slots with no time fill
	slotsWithZeroTime int
	// indicates if closed min and max duration are overlapping
	// Testcase #16
	adjustWithPodMaxDuration bool

	totalSlotTime       int64
	slotWithNonZeroTime int64
}

var multipleOf float64 = 5

func initAlgoParams(config *AdPodConfig) {
	if config.MinDuration == config.MaxDuration {
		// in this case compute closed factor w.r.t multiple of
		config.ClosedMinDuration = getClosetFactor(float64(config.MinDuration), multipleOf)
		config.ClosedMaxDuration = config.ClosedMinDuration
		config.adjustWithPodMaxDuration = true
	} else {
		// get closet pod min duration
		config.ClosedMinDuration = getClosetFactorForMinDuration(float64(config.MinDuration), multipleOf)
		// get closet pod max duration
		config.ClosedMaxDuration = getClosetFactorForMaxDuration(float64(config.MaxDuration), multipleOf)
		//config.adjustWithPodMaxDuration = false
	}

	if config.SlotMinDuration == config.SlotMaxDuration {
		// in this case compute closed factor w.r.t multiple of
		config.ClosedSlotMinDuration = int64(getClosetFactor(float64(config.SlotMinDuration), multipleOf))
		config.ClosedSlotMaxDuration = config.ClosedSlotMinDuration
	} else {
		config.ClosedSlotMinDuration = int64(getClosetFactorForMinDuration(float64(config.SlotMinDuration), multipleOf))
		config.ClosedSlotMaxDuration = int64(getClosetFactorForMaxDuration(float64(config.SlotMaxDuration), multipleOf))
	}

	// handling  when closed durations are overlapping
	config.adjustIfOverlappingClosedMinMaxPodDuration()
}

func getImpressionObjects(podMinDuration, podMaxDuration int64, vPod openrtb_ext.VideoAdPod) AdPodConfig {
	config := AdPodConfig{}

	config.MinDuration = podMinDuration
	config.MaxDuration = podMaxDuration
	config.MinAds = int64(*vPod.MinAds)
	config.MaxAds = int64(*vPod.MaxAds)
	config.SlotMinDuration = int64(*vPod.MinDuration)
	config.SlotMaxDuration = int64(*vPod.MaxDuration)

	initAlgoParams(&config)
	config.getImpressionCount()
	fmt.Printf("\n\n%+v\n\n", config)
	config.getTimeForEachSlot(multipleOf)
	// absoluteTime := config.getTimeForEachSlot(multipleOf)
	//config.constructImpressionObjects(absoluteTime)
	config.adjustFreeTime()
	config.validateTotalImpressions()
	fmt.Println("\nTotal Impressions =", config.impCnt, "(With Free Time Adjusted As Much as possible)")
	fmt.Println("Free Time Unable To Adjust = ", config.freeTime, "sec")
	if config.impCnt > 0 {
		config.print(config.freeTime)
	}
	return config
}

func isMultipleOf(num float64, multipleOf float64) bool {
	return math.Mod(num, multipleOf) == 0
}

func getClosetFactor(num float64, multipleOf float64) float64 {
	return math.Round(num/multipleOf) * float64(multipleOf)
}

func getClosetFactorForMinDuration(MinDuration float64, multipleOf float64) float64 {
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

func getClosetFactorForMaxDuration(maxduration float64, multipleOf float64) float64 {
	closedMaxDuration := getClosetFactor(maxduration, multipleOf)
	if closedMaxDuration == maxduration {
		return maxduration
	}

	// set closet maxduration closed to masduration
	for i := closedMaxDuration; i <= maxduration; {
		if closedMaxDuration < maxduration {
			closedMaxDuration = i + 5
			i = closedMaxDuration
		}
	}

	if closedMaxDuration > maxduration {
		return closedMaxDuration - multipleOf
	}

	return closedMaxDuration
}

func (config *AdPodConfig) getImpressionCount() int64 {
	// compute imp count based
	//config.ClosedSlotMinDuration = int(getClosetFactorForMinDuration(float64(config.SlotMinDuration), multipleOf))
	impCntByMinAdSlotDuration := int64(config.ClosedMaxDuration) / config.ClosedSlotMinDuration
	//config.ClosedSlotMaxDuration = int(getClosetFactorForMaxDuration(float64(config.SlotMaxDuration), multipleOf))
	impCntByMaxAdSlotDuration := int64(config.ClosedMaxDuration) / config.ClosedSlotMaxDuration

	// get max impression count
	if impCntByMaxAdSlotDuration > impCntByMinAdSlotDuration {
		config.impCnt = impCntByMaxAdSlotDuration
	} else {
		config.impCnt = impCntByMinAdSlotDuration
	}

	if config.impCnt > config.MaxAds {
		fmt.Println("max impression count ", config.impCnt, ">", config.MaxAds, "Max allowed ads. Hence, setting to max allowed")
		config.impCnt = config.MaxAds
	}

	if config.impCnt < config.MinAds {
		fmt.Println("max impression count ", config.impCnt, "<", config.MinAds, "Min allowed ads. Hence, setting to min allowed")
		config.impCnt = config.MinAds
	}
	return config.impCnt
}

func (config *AdPodConfig) getTimeForEachSlot(multipier float64) float64 {
	absslottime, mantissa := math.Modf(float64(config.ClosedMaxDuration) / float64(config.impCnt))

	if float64(config.ClosedSlotMinDuration) > float64(config.ClosedSlotMaxDuration) {
		fmt.Println("config.ClosedSlotMinDuration > config.ClosedSlotMaxDuration")
		config.constructImpressionObjects(0)
		return 0

	}

	closeFactorAdjustment := 0.0
	absslottime = getClosetFactor(absslottime, multipleOf)

	// if int(absslottime) > config.ClosedSlotMinDuration {
	if int64(absslottime) > config.ClosedSlotMaxDuration {
		// add  subtracted time for each slot  in   freeTime
		//config.freeTime += (absslottime - float64(config.ClosedSlotMaxDuration)) * float64(config.impCnt)
		closeFactor := float64(config.ClosedSlotMaxDuration)
		closeFactorAdjustment = absslottime - closeFactor
		absslottime = closeFactor
	}

	if absslottime < float64(config.ClosedSlotMinDuration) {
		closeFactor := float64(config.ClosedSlotMinDuration)
		closeFactorAdjustment = closeFactor - absslottime
		absslottime = closeFactor
	}

	config.constructImpressionObjects(absslottime)

	// compute total slots with non-zero value
	// and slot contains value = absslottime
	slotCountFieldWithAbsSlotTime := 0
	for _, slot := range *config.slots {
		if slot != 0 && float64(slot) == absslottime {
			slotCountFieldWithAbsSlotTime++
		}
	}

	if mantissa == 0.0 {
		config.freeTime += closeFactorAdjustment * float64(slotCountFieldWithAbsSlotTime)
	} else {
		if closeFactorAdjustment != 0.0 {
			// it means each slot is alloted with less seconds than  computed absslottime (before closeFactorAdjustment)
			// hence consider that many seconds in free time
			// i.e. plus the value = impression count * closeFactorAdjustment
			// NOTE: both statements in above if condition and here looks same
			// but in this case final value will be positive number. Because here closeFactorAdjustment < 0.0
			closeFactorAdjustment = float64(config.impCnt) * closeFactorAdjustment * -1
		}
		config.freeTime += math.Abs((config.ClosedMaxDuration) - absslottime*float64(slotCountFieldWithAbsSlotTime))
	}

	fmt.Println("Possible number of Ad Slots / Impression with timing assigned (Exclusive of Free Time)")
	fmt.Println("Free Time to be adjused =", config.freeTime, "sec")
	config.print(0)

	return absslottime
}

func (config *AdPodConfig) constructImpressionObjects(absslottime float64) {

	slots := make([]int64, config.impCnt)
	config.imps = make([][2]int64, config.impCnt)
	config.slots = &slots

	totalSumOfSlotTime := 0.0

	for i := int64(0); i < config.impCnt; i++ {
		if totalSumOfSlotTime+absslottime <= config.ClosedMaxDuration {
			slots[i] = int64(absslottime)
			totalSumOfSlotTime += (absslottime)
			config.imps[i][1] = slots[i]
			config.imps[i][0] = config.imps[i][1]
		} else {
			// flag that there are some slots with 0 time
			config.slotsWithZeroTime++
			break
		}
	}

}

func (config *AdPodConfig) validateTotalImpressions() {

	// update impCnt with slotWithNonZeroTime
	config.impCnt = config.slotWithNonZeroTime
	// update slots inside pod object
	nonZeroslots := make([]int64, config.impCnt)
	copy(nonZeroslots, *config.slots)
	*config.slots = nonZeroslots

	if int64(len(*config.slots)) < config.MinAds || int64(len(*config.slots)) > config.MaxAds {
		fmt.Println("Total Impressions= ", len(*config.slots), "  (either <", config.MinAds, " (config.MinAds)  or  >", config.MaxAds, " (config.MaxAds)).")
		config.impCnt = 0
		//config.freeTime = 0
		*config.slots = make([]int64, 0)
	}

	// use MinDuration and MaxDuration of pod instead of ClosedMinDuration and ClosedMaxDuration
	if config.totalSlotTime < config.MinDuration || config.totalSlotTime > config.MaxDuration {

		// if float64(config.totalSlotTime) < config.ClosedMinDuration || float64(config.totalSlotTime) > config.ClosedMaxDuration {
		fmt.Println("Total slot time ", config.totalSlotTime, " sec  (either < ", config.MinDuration, "(minpodtime) or > ", config.MaxDuration, " (maxpodtime)).")
		config.impCnt = 0
		//config.freeTime = 0
		*config.slots = make([]int64, 0)
	}

}

func (config AdPodConfig) print(freeTime float64) {
	setEq := false

	totalTime := int64(0.0)

	for i := 0; i < len(*config.slots); i++ {
		if setEq {
			fmt.Print("  + ")
		}
		fmt.Print((*config.slots)[i])
		totalTime += (*config.slots)[i]
		setEq = true
	}
	totalTime += int64(freeTime)
	fmt.Print("  + (", freeTime, ")          = ", totalTime, " sec (Max Duration =", config.MaxDuration, "sec)")
	fmt.Println()
	setEq = false
	for i := 1; i <= len(*config.slots); i++ {
		if setEq {
			fmt.Print("   ")
		}
		fmt.Print("S"+strconv.Itoa(i), " ")
		setEq = true
	}
	fmt.Println("  (free)")
}

// return 1 -> freetime
func (config *AdPodConfig) adjustFreeTime() int64 {

	config.computeTotalSlotTime()

	timeToMatchWithPodMaxDuration := 0.0
	if config.adjustWithPodMaxDuration && config.totalSlotTime != config.MaxDuration {
		timeToMatchWithPodMaxDuration = float64(config.MinDuration) - float64(config.ClosedMinDuration)
	}

	//freeTime := int(config.freeTime)
	if config.freeTime == 0 && timeToMatchWithPodMaxDuration == 0.0 {
		return int64(config.freeTime)
	}

	// if freetime is in multiples of given number (**multipleOf)
	// get the smallest factor that can be assigned to each slot
	closetSlotDuration := config.freeTime
	if isMultipleOf(config.freeTime, multipleOf) {
		closetSlotDuration = multipleOf
	}

	slotCount := int64(0)
	//closetSlotDuration := float64(config.ClosedSlotMinDuration)
	i := 0.0
	slotsFullWithCapacity := 0
	timeAdjustedFromFreeTime := 0

	// total time counter.  Here taking abosolute value of timeToMatchWithPodMaxDuration
	// timeToMatchWithPodMaxDuration can be negative. See testcase #24
	totalTimeCounter := config.freeTime + math.Abs(timeToMatchWithPodMaxDuration)
	config.imps = make([][2]int64, config.impCnt)
	for i <= totalTimeCounter {
		if config.freeTime == 0 && timeToMatchWithPodMaxDuration == 0 {
			break
		}

		// check if there are any slots which are with 0 time. Give priority to them
		isSlotWithZeroTime := (*config.slots)[slotCount] == 0 && config.slotsWithZeroTime > 0
		slotMaxDuration := config.ClosedSlotMaxDuration // for  config.freeTime

		if config.freeTime == 0 && timeToMatchWithPodMaxDuration != 0 {
			// directly assign timeToMatchWithPodMaxDuration to closetSlotDuration
			// it may not be multipe of given number, which is allowed
			closetSlotDuration = timeToMatchWithPodMaxDuration
			slotMaxDuration = config.SlotMaxDuration // use pod config limit as timeToMatchWithPodMaxDuration is not bound to config.ClosedSlotMaxDuration
		}

		// following condition is true if
		//    1. the slot + closetSlotDuration <= slot max duration and there is no slot with zero time
		// OR 2. the slot is itself with zero time and config.freetime i.e. closetSlotDuration >= config.ClosedSlotMinDuration
		if ((*config.slots)[slotCount]+int64(closetSlotDuration) <= slotMaxDuration && config.slotsWithZeroTime == 0) || (isSlotWithZeroTime && int64(closetSlotDuration) >= config.ClosedSlotMinDuration) {
			// assign
			if (*config.slots)[slotCount] == 0 {
				config.slotsWithZeroTime--
			}
			(*config.slots)[slotCount] += int64(closetSlotDuration)

			config.imps[slotCount][1] = (*config.slots)[slotCount]
			config.imps[slotCount][0] = config.imps[slotCount][1]
			// ensure alloted time is considered by slot time counter
			i = i + closetSlotDuration
			if config.freeTime != 0 {
				timeAdjustedFromFreeTime += int(closetSlotDuration)
				config.freeTime -= float64(closetSlotDuration)
			} else {
				fmt.Println("timeToMatchWithPodMaxDuration (", timeToMatchWithPodMaxDuration, ") is adjusted by slot s", slotCount)
				timeToMatchWithPodMaxDuration = 0 // reset
				closetSlotDuration = timeToMatchWithPodMaxDuration
				config.imps[slotCount][0] = config.MinDuration
				i++ // break for loop
			}
		} else {
			// consider this slot as full of capacity
			slotsFullWithCapacity++
		}

		if slotsFullWithCapacity == len(*config.slots) {
			// all slots are full of capacity
			break
		}

		slotCount++
		if slotCount >= int64(len(*config.slots)) {
			slotCount = 0
			slotsFullWithCapacity = 0
		}
	}

	if config.freeTime != 0 {
		// none of the slot is able adust free time
		fmt.Println("\n\n**** Free time", config.freeTime, "sec, is not adjusted by any slot, because each slot time  is reached max limit or freetime < config.ClosedSlotMinDuration")

	}

	// Sum of all slot times, after free time assignment
	config.computeTotalSlotTime()

	// check if we have opportunity to create slots
	// based on
	// 1. max no of slots
	// 2. min slotduratoin
	// 3. max slot duaration
	if int64(len(*config.slots)) < config.MaxAds && (config.SlotMinDuration <= int64(config.freeTime) && int64(config.freeTime) <= config.SlotMaxDuration) {
		//	config.slots[len(po)]
	}

	return int64(config.freeTime)
}

/*
 Check if Pod Minimum and Maximum Duration is same
*/
func isEqualMinMaxPodDuration(config AdPodConfig) bool {
	return config.MinDuration == config.MaxDuration
}

/*
Checks if computed closed Minimum and Maximum duration ooverlapping.
If overlapping then
*/
func (config *AdPodConfig) adjustIfOverlappingClosedMinMaxPodDuration() {
	if config.ClosedMaxDuration < config.ClosedMinDuration {
		config.adjustWithPodMaxDuration = true
		// set config.ClosedMaxDuration value in  config.ClosedMinDuration
		// For Example,
		// config.ClosedMaxDuration = 125
		// config.ClosedMinDuration = 130
		// then config.ClosedMaxDuration = 125 => config.ClosedMinDuration
		config.ClosedMinDuration = config.ClosedMaxDuration
	} else {
		//config.adjustWithPodMaxDuration = false
	}
}

func (config *AdPodConfig) computeTotalSlotTime() {
	// reset
	config.slotWithNonZeroTime = 0
	config.totalSlotTime = 0
	// Sum of all slot times, after free time assignment
	for i := int64(0); i < config.impCnt; i++ {
		if (*config.slots)[i] != 0 {
			config.totalSlotTime += (*config.slots)[i]
			config.slotWithNonZeroTime++
		}
	}
}
