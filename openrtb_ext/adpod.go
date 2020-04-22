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
	errInvalidMinDuration                         = errors.New("request.imp.video.ext.adpod.minduration must be positive number")
	errInvalidMaxDuration                         = errors.New("request.imp.video.ext.adpod.maxduration must be positive number")
	errInvalidAdvertiserExclusionPercent          = errors.New("request.imp.video.ext.adpod.excladv must be number between 0 and 100")
	errInvalidIABCategoryExclusionPercent         = errors.New("request.imp.video.ext.adpod.excliabcat must be number between 0 and 100")
	errInvalidMinMaxAds                           = errors.New("request.imp.video.ext.adpod.minads must be less than request.imp.video.ext.adpod.maxads")
	errInvalidMinMaxDuration                      = errors.New("request.imp.video.ext.adpod.minduration must be less than request.imp.video.ext.adpod.maxduration")
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
	MinAds                      *int `json:"minads,omitempty"`      //Default 1 if not specified
	MaxAds                      *int `json:"maxads,omitempty"`      //Default 1 if not specified
	MinDuration                 *int `json:"minduration,omitempty"` // adpod.minduration*adpod.minads should be greater than or equal to video.minduration
	MaxDuration                 *int `json:"maxduration,omitempty"` // adpod.maxduration*adpod.maxads should be less than or equal to video.maxduration + video.maxextended
	AdvertiserExclusionPercent  *int `json:"excladv,omitempty"`     // Percent value 0 means none of the ads can be from same advertiser 100 means can have all same advertisers
	IABCategoryExclusionPercent *int `json:"excliabcat,omitempty"`  // Percent value 0 means all ads should be of different IAB categories.
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

//Validate will validate AdPod object
func (pod *VideoAdPod) Validate() error {
	if nil != pod.MinAds && *pod.MinAds <= 0 {
		return errInvalidMinAds
	}

	if nil != pod.MaxAds && *pod.MaxAds <= 0 {
		return errInvalidMaxAds
	}

	if nil != pod.MinDuration && *pod.MinDuration <= 0 {
		return errInvalidMinDuration
	}

	if nil != pod.MaxDuration && *pod.MaxDuration <= 0 {
		return errInvalidMaxDuration
	}

	if nil != pod.AdvertiserExclusionPercent && (*pod.AdvertiserExclusionPercent < 0 || *pod.AdvertiserExclusionPercent > 100) {
		return errInvalidAdvertiserExclusionPercent
	}

	if nil != pod.IABCategoryExclusionPercent && (*pod.IABCategoryExclusionPercent < 0 || *pod.IABCategoryExclusionPercent > 100) {
		return errInvalidIABCategoryExclusionPercent
	}

	if nil != pod.MinAds && nil != pod.MaxAds && *pod.MinAds > *pod.MaxAds {
		return errInvalidMinMaxAds
	}

	if nil != pod.MinDuration && nil != pod.MaxDuration && *pod.MinDuration > *pod.MaxDuration {
		return errInvalidMinMaxDuration
	}

	return nil
}

//Validate will validate ReqAdPodExt object
func (ext *ReqAdPodExt) Validate() error {
	if nil == ext {
		return nil
	}

	if nil != ext.CrossPodAdvertiserExclusionPercent &&
		(*ext.CrossPodAdvertiserExclusionPercent < 0 || *ext.CrossPodAdvertiserExclusionPercent > 100) {
		return errInvalidCrossPodAdvertiserExclusionPercent
	}

	if nil != ext.CrossPodIABCategoryExclusionPercent &&
		(*ext.CrossPodIABCategoryExclusionPercent < 0 || *ext.CrossPodIABCategoryExclusionPercent > 100) {
		return errInvalidCrossPodIABCategoryExclusionPercent
	}

	if nil != ext.IABCategoryExclusionWindow && *ext.IABCategoryExclusionWindow < 0 {
		return errInvalidIABCategoryExclusionWindow
	}

	if nil != ext.AdvertiserExclusionWindow && *ext.AdvertiserExclusionWindow < 0 {
		return errInvalidAdvertiserExclusionWindow
	}

	if err := ext.VideoAdPod.Validate(); nil != err {
		return err
	}

	return nil
}

//Validate will validate video extension object
func (ext *VideoExtension) Validate() error {
	if nil != ext.Offset && *ext.Offset < 0 {
		return errInvalidAdPodOffset
	}

	if nil != ext.AdPod {
		if err := ext.AdPod.Validate(); nil != err {
			return err
		}
	}

	return nil
}
