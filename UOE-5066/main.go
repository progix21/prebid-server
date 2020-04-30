package main

import (
	"fmt"
	"math"
	"strconv"
)

//AdPod
type AdPod struct {
	MinAds          int `json:"minads,omitempty"`      //Default 1 if not specified
	MaxAds          int `json:"maxads,omitempty"`      //Default 1 if not specified
	SlotMinDuration int `json:"minduration,omitempty"` // adpod.minduration*adpod.minads should be greater than or equal to video.minduration
	SlotMaxDuration int `json:"maxduration,omitempty"` // adpod.maxduration*adpod.maxads should be less than or equal to video.maxduration + video.maxextended
	//	AdvertiserExclusionPercent  int `json:"excladv,omitempty"`     // Percent value 0 means none of the ads can be from same advertiser 100 means can have all same advertisers
	//IABCategoryExclusionPercent int `json:"excliabcat,omitempty"`  // Percent value 0 means all ads should be of different IAB categories.
	Minduration int
	MaxDuration int

	slots                 *[]int
	impCnt                int
	freeTime              float64
	ClosedSlotMinDuration int
	ClosedSlotMaxDuration int
	ClosedMinDuration     float64
	ClosedMaxDuration     float64
	// indicates whether there are one/more slots with no time fill
	hasSlotWithZeoTime bool
}

func main() {
	pod := AdPod{}
	pod.Minduration = 5
	pod.MaxDuration = 55
	pod.SlotMinDuration = 2
	pod.SlotMaxDuration = 12
	pod.MinAds = 1
	pod.MaxAds = 1

	pod.getImpressionObjects()

}

var multipleOf float64 = 5

func initAlgoParams(pod *AdPod) {
	if pod.Minduration == pod.MaxDuration {
		// in this case compute closed factor w.r.t multiple of
		pod.ClosedMinDuration = getClosetFactor(float64(pod.Minduration), multipleOf)
		pod.ClosedMaxDuration = pod.ClosedMinDuration
	} else {
		// get closet pod min duration
		pod.ClosedMinDuration = getClosetFactorForMinDuration(float64(pod.Minduration), multipleOf)
		// get closet pod max duration
		pod.ClosedMaxDuration = getClosetFactorForMaxDuration(float64(pod.MaxDuration), multipleOf)
	}

	if pod.SlotMinDuration == pod.SlotMaxDuration {
		// in this case compute closed factor w.r.t multiple of
		pod.ClosedSlotMinDuration = int(getClosetFactor(float64(pod.SlotMinDuration), multipleOf))
		pod.ClosedSlotMaxDuration = pod.ClosedSlotMinDuration
	} else {
		pod.ClosedSlotMinDuration = int(getClosetFactorForMinDuration(float64(pod.SlotMinDuration), multipleOf))
		pod.ClosedSlotMaxDuration = int(getClosetFactorForMaxDuration(float64(pod.SlotMaxDuration), multipleOf))
	}
}

func (pod *AdPod) getImpressionObjects() int {

	initAlgoParams(pod)
	pod.getImpressionCount()
	fmt.Printf("\n\n%+v\n\n", pod)
	pod.getTimeForEachSlot(multipleOf)
	// absoluteTime := pod.getTimeForEachSlot(multipleOf)
	//pod.constructImpressionObjects(absoluteTime)
	pod.adjustFreeTime()
	pod.validateTotalImpressions()
	fmt.Println("\nTotal Impressions =", pod.impCnt, "(With Free Time Adjusted As Much as possible)")
	fmt.Println("Free Time Unable To Adjust = ", pod.freeTime, "sec")
	if pod.impCnt > 0 {
		pod.print(pod.freeTime)
	}

	return pod.impCnt
}

func isMultipleOf(num float64, multipleOf float64) bool {
	return math.Mod(num, multipleOf) == 0
}

func getClosetFactor(num float64, multipleOf float64) float64 {
	return math.Round(num/multipleOf) * float64(multipleOf)
}

func getClosetFactorForMinDuration(minDuration float64, multipleOf float64) float64 {
	closedMinDuration := getClosetFactor(minDuration, multipleOf)

	if closedMinDuration == 0 {
		return multipleOf
	}

	if closedMinDuration == minDuration {
		return minDuration
	}

	if closedMinDuration < minDuration {
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

func (pod *AdPod) getImpressionCount() int {
	// compute imp count based
	//pod.ClosedSlotMinDuration = int(getClosetFactorForMinDuration(float64(pod.SlotMinDuration), multipleOf))
	impCntByMinAdSlotDuration := int(pod.ClosedMaxDuration) / pod.ClosedSlotMinDuration
	//pod.ClosedSlotMaxDuration = int(getClosetFactorForMaxDuration(float64(pod.SlotMaxDuration), multipleOf))
	impCntByMaxAdSlotDuration := int(pod.ClosedMaxDuration) / pod.ClosedSlotMaxDuration

	// get max impression count
	if impCntByMaxAdSlotDuration > impCntByMinAdSlotDuration {
		pod.impCnt = impCntByMaxAdSlotDuration
	} else {
		pod.impCnt = impCntByMinAdSlotDuration
	}

	if pod.impCnt > pod.MaxAds {
		fmt.Println("max impression count ", pod.impCnt, ">", pod.MaxAds, "Max allowed ads. Hence, setting to max allowed")
		pod.impCnt = pod.MaxAds
	}

	if pod.impCnt < pod.MinAds {
		fmt.Println("max impression count ", pod.impCnt, "<", pod.MinAds, "Min allowed ads. Hence, setting to min allowed")
		pod.impCnt = pod.MinAds
	}
	return pod.impCnt
}

func (pod *AdPod) getTimeForEachSlot(multipier float64) float64 {
	absslottime, mantissa := math.Modf(float64(pod.ClosedMaxDuration) / float64(pod.impCnt))

	if float64(pod.ClosedSlotMinDuration) > float64(pod.ClosedSlotMaxDuration) {
		fmt.Println("closetSlotMinDuration > pod.SlotMaxDuration")
		return 0

	}

	closeFactorAdjustment := 0.0
	// if !isMultipleOf(absslottime, multipleOf) {
	// 	closeFactor := getClosetFactor(absslottime, multipleOf)
	// 	// if absslottime = 12 then closet factor will be 10 (in case of multiple of 5). it means will not allocated 2 seconds
	// 	// if absslottime = 13 then closet factor will be 15 (in case of multiple of 5). it means it will allocated extra 2 seconds
	// 	closeFactorAdjustment = closeFactor - absslottime
	// 	//absslottime = closeFactor
	// }
	absslottime = getClosetFactor(absslottime, multipleOf)

	// if int(absslottime) > pod.ClosedSlotMinDuration {
	if int(absslottime) > pod.ClosedSlotMaxDuration {
		// add  subtracted time for each slot  in   freeTime
		//pod.freeTime += (absslottime - float64(pod.ClosedSlotMaxDuration)) * float64(pod.impCnt)
		closeFactor := float64(pod.ClosedSlotMaxDuration)
		closeFactorAdjustment = absslottime - closeFactor
		absslottime = closeFactor
	}

	if absslottime < float64(pod.ClosedSlotMinDuration) {
		closeFactor := float64(pod.ClosedSlotMinDuration)
		closeFactorAdjustment = closeFactor - absslottime
		absslottime = closeFactor
	}

	pod.constructImpressionObjects(absslottime)

	// compute total slots with non-zero value
	// and slot contains value = absslottime
	slotCountFieldWithAbsSlotTime := 0
	for _, slot := range *pod.slots {
		if slot != 0 && float64(slot) == absslottime {
			slotCountFieldWithAbsSlotTime++
		}
	}

	if mantissa == 0.0 {
		pod.freeTime += closeFactorAdjustment * float64(slotCountFieldWithAbsSlotTime)
	} else {
		//	localAbsTime, _ := math.Modf(float64(pod.ClosedMaxDuration) / float64(pod.impCnt))
		// if !isMultipleOf(localAbsTime, multipleOf) {
		// 	localAbsTime = getClosetFactor(localAbsTime, multipleOf)
		// }
		//pod.freeTime += math.Abs((pod.ClosedMaxDuration) - localAbsTime*float64(pod.impCnt))

		// if closeFactorAdjustment > 0.0 {
		// 	// it means each slot is alloted with extra seconds
		// 	// hence do not consider that many seconds in free time
		// 	// i.e. minus the value = impression count * closeFactorAdjustment
		// 	closeFactorAdjustment = float64(pod.impCnt) * closeFactorAdjustment * -1
		// }

		// if closeFactorAdjustment < 0.0 {
		// 	// it means each slot is alloted with less seconds than  computed absslottime (before closeFactorAdjustment)
		// 	// hence consider that many seconds in free time
		// 	// i.e. plus the value = impression count * closeFactorAdjustment
		// 	// NOTE: both statements in above if condition and here looks same
		// 	// but in this case final value will be positive number. Because here closeFactorAdjustment < 0.0
		// 	closeFactorAdjustment = float64(pod.impCnt) * closeFactorAdjustment * -1
		// }

		// if closeFactorAdjustment > 0.0 {
		// 	// it means each slot is alloted with extra seconds
		// 	// hence do not consider that many seconds in free time
		// 	// i.e. minus the value = impression count * closeFactorAdjustment
		// 	closeFactorAdjustment = float64(pod.impCnt) * closeFactorAdjustment * -1
		// }

		if closeFactorAdjustment != 0.0 {
			// it means each slot is alloted with less seconds than  computed absslottime (before closeFactorAdjustment)
			// hence consider that many seconds in free time
			// i.e. plus the value = impression count * closeFactorAdjustment
			// NOTE: both statements in above if condition and here looks same
			// but in this case final value will be positive number. Because here closeFactorAdjustment < 0.0
			closeFactorAdjustment = float64(pod.impCnt) * closeFactorAdjustment * -1
		}
		//pod.freeTime += math.Abs((pod.ClosedMaxDuration) - absslottime*float64(pod.impCnt))
		pod.freeTime += math.Abs((pod.ClosedMaxDuration) - absslottime*float64(slotCountFieldWithAbsSlotTime))
		// consideration of closeFactorAdjustment
		//  always += because when closeFactorAdjustment it will subtracted
		//		pod.freeTime += closeFactorAdjustment
	}

	fmt.Println("Possible number of Ad Slots / Impression with timing assigned (Exclusive of Free Time)")
	fmt.Println("Free Time to be adjused =", pod.freeTime, "sec")
	pod.print(0)

	return absslottime
}

func (pod *AdPod) constructImpressionObjects(absslottime float64) {

	slots := make([]int, pod.impCnt)
	pod.slots = &slots

	totalSumOfSlotTime := 0.0

	for i := 0; i < pod.impCnt; i++ {
		if totalSumOfSlotTime+absslottime <= pod.ClosedMaxDuration {
			slots[i] = int(absslottime)
			totalSumOfSlotTime += (absslottime)
		} else {
			// flag that there are some slots with 0 time
			pod.hasSlotWithZeoTime = true
			break
		}
	}

}

func (pod *AdPod) validateTotalImpressions() {
	totalSlotTime := 0
	slotWithNonZeroTime := 0
	for i := 0; i < pod.impCnt; i++ {
		if (*pod.slots)[i] != 0 {
			totalSlotTime += (*pod.slots)[i]
			slotWithNonZeroTime++
		}
	}

	// allow to remove slot only if > pod.MinAds
	//if slotWithNonZeroTime > pod.MinAds {
	// update impCnt with slotWithNonZeroTime
	pod.impCnt = slotWithNonZeroTime
	// update slots inside pod object
	nonZeroslots := make([]int, pod.impCnt)
	copy(nonZeroslots, *pod.slots)
	*pod.slots = nonZeroslots
	//}

	if len(*pod.slots) < pod.MinAds || len(*pod.slots) > pod.MaxAds {
		fmt.Println("Total Impressions= ", len(*pod.slots), "  (either <", pod.MinAds, " (pod.MinAds)  or  >", pod.MaxAds, " (pod.MaxAds)).")
		pod.impCnt = 0
		*pod.slots = make([]int, 0)
	}

	if float64(totalSlotTime) < pod.ClosedMinDuration || float64(totalSlotTime) > pod.ClosedMaxDuration {
		fmt.Println("Total slot time ", totalSlotTime, " sec  (either < ", pod.ClosedMinDuration, "(minpodtime) or > ", pod.ClosedMaxDuration, " (maxpodtime)).")
		pod.impCnt = 0
		*pod.slots = make([]int, 0)
	}

}

func (pod AdPod) print(freeTime float64) {
	setEq := false

	totalTime := 0

	for i := 0; i < len(*pod.slots); i++ {
		if setEq {
			fmt.Print("  + ")
		}
		fmt.Print((*pod.slots)[i])
		totalTime += (*pod.slots)[i]
		setEq = true
	}
	totalTime += int(freeTime)
	fmt.Print("  + (", freeTime, ")          = ", totalTime, " sec (Max Duration =", pod.MaxDuration, "sec)")
	fmt.Println()
	setEq = false
	for i := 1; i <= len(*pod.slots); i++ {
		if setEq {
			fmt.Print("   ")
		}
		fmt.Print("S"+strconv.Itoa(i), " ")
		setEq = true
	}
	fmt.Println("  (free)")
}

// return 1 -> freetime
func (pod *AdPod) adjustFreeTime() int {

	freeTime := int(pod.freeTime)
	if freeTime == 0 {
		return freeTime
	}

	// // check slot0
	// slot0 := pod.slots[0]

	// // check next closet value of each slot duration
	// // ideally it should be same based on algo logic
	// // assuming it will same for all slots
	// // considering value at slot0
	// closetSlotDuration := getClosetFactor(float64(slot0), multipleOf)

	// if freetime is in multiples of given number (**multipleOf)
	// get the smallest factor that can be assigned to each slot
	closetSlotDuration := pod.freeTime
	if isMultipleOf(pod.freeTime, multipleOf) {
		closetSlotDuration = multipleOf
	}

	// // check if adding closetSlotDuration to each slot not exceeding
	// // ClosedSlotMaxDuration
	// // we will check only for slot0, assumimg all slots have same values
	// if float64(slot0)+closetSlotDuration <= float64(pod.ClosedSlotMaxDuration) {
	// assign closetSlotDuration to each slot
	// till free time value = 0
	slotCount := 0
	//closetSlotDuration := float64(pod.ClosedSlotMinDuration)
	i := 0.0
	slotsFullWithCapacity := 0
	timeAdjustedFromFreeTime := 0
	for i <= pod.freeTime {
		// check if there are any slots which are with 0 time. Give priority to them
		isSlotWithZeroTime := (*pod.slots)[slotCount] == 0 && pod.hasSlotWithZeoTime
		// following condition is true if
		//    1. the slot + closetSlotDuration = slotmax duration and there is no slot with zero time
		// OR 2. the slot is itself, slot with zero time
		if ((*pod.slots)[slotCount]+int(closetSlotDuration) <= pod.ClosedSlotMaxDuration && !pod.hasSlotWithZeoTime) || isSlotWithZeroTime {
			// assign
			(*pod.slots)[slotCount] += int(closetSlotDuration)
			// ensure alloted time is considered by slot time counter
			i = i + closetSlotDuration
			timeAdjustedFromFreeTime += int(closetSlotDuration)
			pod.freeTime -= float64(closetSlotDuration)
		} else {
			// consider this slot as full of capacity
			slotsFullWithCapacity++
		}

		if slotsFullWithCapacity == len(*pod.slots) {
			// all slots are full of capacity
			//subtract adjusted from free time
			//pod.freeTime -= float64(timeAdjustedFromFreeTime)
			// come out of for
			break
		}

		slotCount++
		if slotCount >= len(*pod.slots) {
			slotCount = 0
		}
	}

	// if i-closetSlotDuration == pod.freeTime {
	// 	// store freeTime which was not adjusted if any
	// 	pod.freeTimeNotAdjusted = pod.freeTime
	// 	// reset free time
	// 	pod.freeTime = 0

	// }
	// } else {
	if pod.freeTime != 0 {
		//
		// none of the slot is able adust free time
		fmt.Println("\n\n**** Free time", freeTime, "sec, is not adjusted by any slot, because each slot time  is reached max limit")

	}

	// check if we have opportunity to create slots
	// based on
	// 1. max no of slots
	// 2. min slotduratoin
	// 3. max slot duaration
	if len(*pod.slots) < pod.MaxAds && (pod.SlotMinDuration <= freeTime && freeTime <= pod.SlotMaxDuration) {
		//	pod.slots[len(po)]
	}

	return freeTime
}
