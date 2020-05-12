package ctv

const (
	// instructs impression algorithm to define ad slots
	// using Ad Pod Max Duration and Max Ads
	inclineToMaxPodDurationMaxAds = iota

	// instructs impression algorithm to define ad slots
	// using Ad Pod Max Duration and Min Ads
	inclineToMaxPodDurationMinAds = iota

	// instructs impression algorithm to define ad slots
	// using both Ad Pod Min and Max Duration
	inclineToMinMaxPodDuration = iota
)
