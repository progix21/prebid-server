package openrtb_ext

import (
	"errors"
)

var (
	errInvalidCrossPodAdvertiserExclusionPercent  = errors.New("request.ext.adpod.crosspodexcladv must be a number between 0 and 100")
	errInvalidCrossPodIABCategoryExclusionPercent = errors.New("request.ext.adpod.crosspodexcliabcat must be a number between 0 and 100")
	errInvalidIABCategoryExclusionWindow          = errors.New("request.ext.adpod.excliabcatwindow must be postive number")
	errInvalidAdvertiserExclusionWindow           = errors.New("request.ext.adpod.excladvwindow must be postive number")
	errInvalidAdPodOffset                         = errors.New("request.imp.video.ext.offset must be postive number")
	errInvalidMinAds                              = errors.New("request.imp.video.ext.adpod.minads must be positive number")
	errInvalidMaxAds                              = errors.New("request.imp.video.ext.adpod.maxads must be positive number")
	errInvalidMinDuration                         = errors.New("request.imp.video.ext.adpod.adminduration must be positive number")
	errInvalidMaxDuration                         = errors.New("request.imp.video.ext.adpod.admaxduration must be positive number")
	errInvalidAdvertiserExclusionPercent          = errors.New("request.imp.video.ext.adpod.excladv must be number between 0 and 100")
	errInvalidIABCategoryExclusionPercent         = errors.New("request.imp.video.ext.adpod.excliabcat must be number between 0 and 100")
	errInvalidMinMaxAds                           = errors.New("request.imp.video.ext.adpod.minads must be less than request.imp.video.ext.adpod.maxads")
	errInvalidMinMaxDuration                      = errors.New("request.imp.video.ext.adpod.adminduration must be less than request.imp.video.ext.adpod.admaxduration")
)

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
	MinAds                      *int `json:"minads,omitempty"`        //Default 1 if not specified
	MaxAds                      *int `json:"maxads,omitempty"`        //Default 1 if not specified
	MinDuration                 *int `json:"adminduration,omitempty"` // (adpod.adminduration * adpod.minads) should be greater than or equal to video.minduration
	MaxDuration                 *int `json:"admaxduration,omitempty"` // (adpod.admaxduration * adpod.maxads) should be less than or equal to video.maxduration + video.maxextended
	AdvertiserExclusionPercent  *int `json:"excladv,omitempty"`       // Percent value 0 means none of the ads can be from same advertiser 100 means can have all same advertisers
	IABCategoryExclusionPercent *int `json:"excliabcat,omitempty"`    // Percent value 0 means all ads should be of different IAB categories.
}

/*
//UnmarshalJSON will unmarshal extension into VideoExtension object
func (ext *VideoExtension) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, ext)
}

//UnmarshalJSON will unmarshal extension into ReqAdPodExt object
func (ext *ReqAdPodExt) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, ext)
}
*/

func getIntPtr(v int) *int {
	return &v
}

//Validate will validate AdPod object
func (pod *VideoAdPod) Validate() (err []error) {
	if nil != pod.MinAds && *pod.MinAds <= 0 {
		err = append(err, errInvalidMinAds)
	}

	if nil != pod.MaxAds && *pod.MaxAds <= 0 {
		err = append(err, errInvalidMaxAds)
	}

	if nil != pod.MinDuration && *pod.MinDuration <= 0 {
		err = append(err, errInvalidMinDuration)
	}

	if nil != pod.MaxDuration && *pod.MaxDuration <= 0 {
		err = append(err, errInvalidMaxDuration)
	}

	if nil != pod.AdvertiserExclusionPercent && (*pod.AdvertiserExclusionPercent < 0 || *pod.AdvertiserExclusionPercent > 100) {
		err = append(err, errInvalidAdvertiserExclusionPercent)
	}

	if nil != pod.IABCategoryExclusionPercent && (*pod.IABCategoryExclusionPercent < 0 || *pod.IABCategoryExclusionPercent > 100) {
		err = append(err, errInvalidIABCategoryExclusionPercent)
	}

	if nil != pod.MinAds && nil != pod.MaxAds && *pod.MinAds > *pod.MaxAds {
		err = append(err, errInvalidMinMaxAds)
	}

	if nil != pod.MinDuration && nil != pod.MaxDuration && *pod.MinDuration > *pod.MaxDuration {
		err = append(err, errInvalidMinMaxDuration)
	}

	return
}

//Validate will validate ReqAdPodExt object
func (ext *ReqAdPodExt) Validate() (err []error) {
	if nil == ext {
		return
	}

	if nil != ext.CrossPodAdvertiserExclusionPercent &&
		(*ext.CrossPodAdvertiserExclusionPercent < 0 || *ext.CrossPodAdvertiserExclusionPercent > 100) {
		err = append(err, errInvalidCrossPodAdvertiserExclusionPercent)
	}

	if nil != ext.CrossPodIABCategoryExclusionPercent &&
		(*ext.CrossPodIABCategoryExclusionPercent < 0 || *ext.CrossPodIABCategoryExclusionPercent > 100) {
		err = append(err, errInvalidCrossPodIABCategoryExclusionPercent)
	}

	if nil != ext.IABCategoryExclusionWindow && *ext.IABCategoryExclusionWindow < 0 {
		err = append(err, errInvalidIABCategoryExclusionWindow)
	}

	if nil != ext.AdvertiserExclusionWindow && *ext.AdvertiserExclusionWindow < 0 {
		err = append(err, errInvalidAdvertiserExclusionWindow)
	}

	if errL := ext.VideoAdPod.Validate(); nil != errL {
		err = append(err, errL...)
	}

	return
}

//Validate will validate video extension object
func (ext *VideoExtension) Validate() (err []error) {
	if nil != ext.Offset && *ext.Offset < 0 {
		err = append(err, errInvalidAdPodOffset)
	}

	if nil != ext.AdPod {
		if errL := ext.AdPod.Validate(); nil != errL {
			err = append(err, errL...)
		}
	}

	return
}

//SetDefaultValue will set default values if not present
func (pod *VideoAdPod) SetDefaultValue() {
	//pod.MinAds setting default value
	if nil == pod.MinAds {
		pod.MinAds = getIntPtr(2)
	}

	//pod.MaxAds setting default value
	if nil == pod.MaxAds {
		pod.MaxAds = getIntPtr(3)
	}

	//pod.AdvertiserExclusionPercent setting default value
	if nil == pod.AdvertiserExclusionPercent {
		pod.AdvertiserExclusionPercent = getIntPtr(100)
	}

	//pod.IABCategoryExclusionPercent setting default value
	if nil == pod.IABCategoryExclusionPercent {
		pod.IABCategoryExclusionPercent = getIntPtr(100)
	}
}

//SetDefaultValue will set default values if not present
func (ext *ReqAdPodExt) SetDefaultValue() {
	//ext.VideoAdPod setting default value
	ext.VideoAdPod.SetDefaultValue()

	//ext.CrossPodAdvertiserExclusionPercent setting default value
	if nil == ext.CrossPodAdvertiserExclusionPercent {
		ext.CrossPodAdvertiserExclusionPercent = getIntPtr(100)
	}

	//ext.CrossPodIABCategoryExclusionPercent setting default value
	if nil == ext.CrossPodIABCategoryExclusionPercent {
		ext.CrossPodIABCategoryExclusionPercent = getIntPtr(100)
	}

	//ext.IABCategoryExclusionWindow setting default value
	if nil == ext.IABCategoryExclusionWindow {
		ext.IABCategoryExclusionWindow = getIntPtr(0)
	}

	//ext.AdvertiserExclusionWindow setting default value
	if nil == ext.AdvertiserExclusionWindow {
		ext.AdvertiserExclusionWindow = getIntPtr(0)
	}
}

//SetDefaultValue will set default values if not present
func (ext *VideoExtension) SetDefaultValue() {
	//ext.Offset setting default values
	if nil == ext.Offset {
		ext.Offset = getIntPtr(0)
	}

	//ext.AdPod setting default values
	if nil == ext.AdPod {
		ext.AdPod = &VideoAdPod{}
	}
	ext.AdPod.SetDefaultValue()
}

//SetDefaultAdDuration will set default pod ad slot durations
func (pod *VideoAdPod) SetDefaultAdDurations(podMinDuration, podMaxDuration int64) {
	//pod.MinDuration setting default adminduration
	if nil == pod.MinDuration {
		duration := int(podMinDuration / 2)
		pod.MinDuration = &duration
	}

	//pod.MaxDuration setting default admaxduration
	if nil == pod.MaxDuration {
		duration := int(podMaxDuration / 2)
		pod.MaxDuration = &duration
	}
}

//Merge VideoAdPod Values
func (pod *VideoAdPod) Merge(parent *VideoAdPod) {
	//pod.MinAds setting default value
	if nil == pod.MinAds {
		pod.MinAds = parent.MinAds
	}

	//pod.MaxAds setting default value
	if nil == pod.MaxAds {
		pod.MaxAds = parent.MaxAds
	}

	//pod.AdvertiserExclusionPercent setting default value
	if nil == pod.AdvertiserExclusionPercent {
		pod.AdvertiserExclusionPercent = parent.AdvertiserExclusionPercent
	}

	//pod.IABCategoryExclusionPercent setting default value
	if nil == pod.IABCategoryExclusionPercent {
		pod.IABCategoryExclusionPercent = parent.IABCategoryExclusionPercent
	}
}
