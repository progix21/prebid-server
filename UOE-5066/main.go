package main

import (
	"fmt"
	"math"
	"os"
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

	slots                 []int
	impCnt                int
	freeTime              float64
	ClosedSlotMinDuration int
	ClosedSlotMaxDuration int
	ClosedMinDuration     float64
	ClosedMaxDuration     float64
}

func main() {
	pod := AdPod{}
	pod.Minduration = 127
	pod.MaxDuration = 128
	pod.SlotMinDuration = 1
	pod.SlotMaxDuration = 12
	pod.MinAds = 7
	pod.MaxAds = 13

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

	pod.getImpressionObjects()
}

var multipleOf float64 = 5

func (pod *AdPod) getImpressionObjects() int {

	pod.getImpressionCount()
	fmt.Printf("%+v\n\n", pod)
	absoluteTime := pod.getTimeForEachSlot(multipleOf)
	pod.constructImpressionObjects(absoluteTime)
	pod.adjustFreeTime()
	pod.validateTotalImpressions()
	fmt.Println("Total Impressions =", pod.impCnt)
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
	impCntByMaxAdSlotDuration := int(pod.MaxDuration) / pod.ClosedSlotMaxDuration

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
		os.Exit(1)

	}

	if int(absslottime) > pod.ClosedSlotMinDuration {
		// add  subtracted time for each slot  in   freeTime
		pod.freeTime += (absslottime - float64(pod.ClosedSlotMinDuration)) * float64(pod.impCnt)
		absslottime = float64(pod.ClosedSlotMinDuration)
	}

	if absslottime < float64(pod.ClosedSlotMinDuration) {
		absslottime = float64(pod.ClosedSlotMinDuration)
	}

	if mantissa != 0.0 {
		localAbsTime, _ := math.Modf(float64(pod.ClosedMaxDuration) / float64(pod.impCnt))
		pod.freeTime += float64(pod.ClosedMaxDuration) - localAbsTime*float64(pod.impCnt)
	}

	return absslottime
}

func (pod *AdPod) constructImpressionObjects(absslottime float64) {
	pod.slots = make([]int, pod.impCnt)
	for i := 0; i < pod.impCnt; i++ {
		pod.slots[i] = int(absslottime)
	}
	pod.print(pod.freeTime)
}

func (pod *AdPod) validateTotalImpressions() {
	totalSlotTime := 0
	for i := 0; i < pod.impCnt; i++ {
		totalSlotTime += pod.slots[i]
	}

	if float64(totalSlotTime) < pod.ClosedMinDuration || float64(totalSlotTime) > pod.ClosedMaxDuration {
		fmt.Println("Total slot time ", totalSlotTime, " sec  (either < minpodtime or > maxpodtime).")
		pod.impCnt = 0
	}
}

func (pod AdPod) print(freeTime float64) {
	setEq := false

	totalTime := 0

	for i := 0; i < len(pod.slots); i++ {
		if setEq {
			fmt.Print("  + ")
		}
		fmt.Print(pod.slots[i])
		totalTime += pod.slots[i]
		setEq = true
	}
	totalTime += int(freeTime)
	fmt.Print("  + ", freeTime, "           = ", totalTime, "sec (Max Duration =", pod.MaxDuration, "sec)")
	fmt.Println()
	setEq = false
	for i := 1; i <= len(pod.slots); i++ {
		if setEq {
			fmt.Print("   ")
		}
		fmt.Print("S" + strconv.Itoa(i))
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

	// check slot0
	slot0 := pod.slots[0]

	// check next closet value of each slot duration
	// ideally it should be same based on algo logic
	// assuming it will same for all slots
	// considering value at slot0
	closetSlotDuration := getClosetFactor(float64(slot0), multipleOf)

	// if freetime is in multiples of given number (**multipleOf)
	// get the smallest factor that can be assigned to each slot
	if isMultipleOf(pod.freeTime, multipleOf) {
		closetSlotDuration = multipleOf
	}

	// check if adding closetSlotDuration to each slot not exceeding
	// ClosedSlotMaxDuration
	// we will check only for slot0, assumimg all slots have same values
	if float64(slot0)+closetSlotDuration <= float64(pod.ClosedSlotMaxDuration) {
		// assign closetSlotDuration to each slot
		// till free time value = 0
		slotCount := 0
		i := closetSlotDuration
		slotsFullWithCapacity := 0
		timeAdjustedFromFreeTime := 0
		for i <= pod.freeTime {
			// check if slot time + closetSlotDuration < pod.ClosedSlotMaxDuration
			if pod.slots[slotCount]+int(closetSlotDuration) <= pod.ClosedSlotMaxDuration {
				// assign
				pod.slots[slotCount] += int(closetSlotDuration)
				// ensure alloted time is considered by slot time counter
				i = i + closetSlotDuration
				timeAdjustedFromFreeTime += int(closetSlotDuration)
			} else {
				// consider this slot as full of capacity
				slotsFullWithCapacity++
			}

			if slotsFullWithCapacity == len(pod.slots) {
				// all slots are full of capacity
				//subtract adjusted from free time
				pod.freeTime -= float64(timeAdjustedFromFreeTime)
				// come out of for
				break
			}

			slotCount++
			if slotCount >= len(pod.slots) {
				slotCount = 0
			}
		}

		if i-closetSlotDuration == pod.freeTime {
			pod.freeTime = 0
		}
	} else {
		//
		// none of the slot is able adust free time
		fmt.Println("\n\n**** Free time", freeTime, "sec, is not adjusted by any slot, because each slot time  is reached max limit")

	}

	// check if we have opportunity to create slots
	// based on
	// 1. max no of slots
	// 2. min slotduratoin
	// 3. max slot duaration
	if len(pod.slots) < pod.MaxAds && (pod.SlotMinDuration <= freeTime && freeTime <= pod.SlotMaxDuration) {
		//	pod.slots[len(po)]
	}

	return freeTime
}
