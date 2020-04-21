package ctv

import (
	"github.com/PubMatic-OpenWrap/openrtb"
)

//ImpBid type of data to be present for combinations
type ImpBid struct {
	*openrtb.Bid
	SeatName string
}

//Comparator check exclusion conditions
type Comparator func(*ImpBid, *ImpBid) bool

//ImpBids combination contains ImpBid
type ImpBids []*ImpBid

//IAdPodGenerator interface for generating AdPod from Ads
type IAdPodGenerator interface {
	GetAdPod() ImpBids
}
