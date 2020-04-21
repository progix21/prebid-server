package openrtb_ext

//VideoExtension structure to accept video specific more parameters like adpod
type VideoExtension struct {
	Offset *int        `json:"offset,omitempty"` // Minutes from start where this ad is intended to show
	AdPod  *VideoAdPod `json:"adpod,omitempty"`
}

//ReqAdPodExt holds AdPod specific extension parameters at request level
type ReqAdPodExt struct {
	VideoAdPod
	CrossPodAdvertiserExclusionPercent  *int `json:"crosspodexcladv,omitempty"`    //Percent Value - Across multiple impression there will be no ads from same advertiser. Note: These cross pod rule % values can not be more restrictive than per pod
	CrossPodIABCategoryExclusionPercent *int `json:"crosspodexcliabcat,omitempty"` //Percent Value - Across multiple impression there will be no ads from same advertiser
	IABCategoryExclusionWindow          *int `json:"excliabcatwindow,omitempty"`   //Duration in minute between pods where exclusive IAB rule needs to be applied
	AdvertiserExclusionWindow           *int `json:"excladvwindow,omitempty"`      //Duration in minute between pods where exclusive advertiser rule needs to be applied
}

//VideoAdPod holds Video AdPod specific extension parameters at impression level
type VideoAdPod struct {
	MinAds                      *int `json:"minads,omitempty"`      //Default 1 if not specified
	MaxAds                      *int `json:"maxads,omitempty"`      //Default 1 if not specified
	MinDuration                 *int `json:"minduration,omitempty"` // adpod.minduration*adpod.minads should be greater than or equal to video.minduration
	MaxDuration                 *int `json:"maxduration,omitempty"` // adpod.maxduration*adpod.maxads should be less than or equal to video.maxduration + video.maxextended
	AdvertiserExclusionPercent  *int `json:"excladv,omitempty"`     // Percent value 0 means none of the ads can be from same advertiser 100 means can have all same advertisers
	IABCategoryExclusionPercent *int `json:"excliabcat,omitempty"`  // Percent value 0 means all ads should be of different IAB categories.
}
