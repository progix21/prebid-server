package ctv

import (
	"fmt"

	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
)

//GetImpressionsByConfig by config
func GetImpressionsByConfig(podMinDuration, podMaxDuration int64, slotMinDuration, slotMaxDuration, minAds, maxAds int) [][2]int64 {

	// input as  is
	imps := GetImpressions(podMinDuration, podMaxDuration, constructPod(slotMinDuration, slotMaxDuration, minAds, maxAds))
	// impressions by pod max duration
	imps0 := GetImpressions(podMaxDuration, podMaxDuration, constructPod(slotMinDuration, slotMaxDuration, minAds, maxAds))
	// impressions by pod max duration and min ads
	imps1 := GetImpressions(podMaxDuration, podMaxDuration, constructPod(slotMinDuration, slotMaxDuration, minAds, minAds))

	// impressions by pod min duration
	imps2 := GetImpressions(podMinDuration, podMinDuration, constructPod(slotMinDuration, slotMaxDuration, minAds, maxAds))
	// impressions by pod max duration and min ads
	imps3 := GetImpressions(podMinDuration, podMinDuration, constructPod(slotMinDuration, slotMaxDuration, minAds, minAds))

	fmt.Println(imps)
	fmt.Println(imps0)
	fmt.Println(imps1)
	fmt.Println(imps2)
	fmt.Println(imps3)

	return imps0
}

func constructPod(slotMinDuration, slotMaxDuration, minAds, maxAds int) openrtb_ext.VideoAdPod {
	pod := openrtb_ext.VideoAdPod{}
	pod.MinAds = &minAds
	pod.MaxAds = &maxAds
	pod.MinDuration = &slotMinDuration
	pod.MaxDuration = &slotMaxDuration
	return pod
}
