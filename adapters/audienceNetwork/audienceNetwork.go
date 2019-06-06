package audienceNetwork

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/errortypes"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
	"github.com/PubMatic-OpenWrap/prebid-server/pbs"
	"github.com/mxmCherry/openrtb"
)

type AudienceNetworkAdapter struct {
	http       *adapters.HTTPAdapter
	URI        string
	PlatformID string
}

type audienceNetworkParams struct {
	AppID       string `json:"appId"`
	PlacementID string `json:"placementId"`
}

var anSupportedHeight = map[uint64]bool{
	50:  true,
	250: true,
}

// used for cookies and such
func (a *AudienceNetworkAdapter) Name() string {
	return string(openrtb_ext.BidderFacebook)
}

func (a *AudienceNetworkAdapter) SkipNoCookies() bool {
	return true
}

func (a *AudienceNetworkAdapter) Call(ctx context.Context, req *pbs.PBSRequest, bidder *pbs.PBSBidder) (pbs.PBSBidSlice, error) {
	return nil, nil
}

//func (a *AudienceNetworkAdapter) MakeOpenRtbBidRequest(req *pbs.PBSRequest, bidder *pbs.PBSBidder, placementId string, mtype pbs.MediaType, pubId string, unitInd int) (openrtb.BidRequest, error) {
//	return openrtb.BidRequest{}, nil
//}

func (a *AudienceNetworkAdapter) MakeRequests(request *openrtb.BidRequest) ([]*adapters.RequestData, []error) {
	errs := make([]error, 0, len(request.Imp))

	pubID, err := getAppIDFromRequest(request)
	if err != nil {
		return nil, []error{&errortypes.BadInput{
			Message: fmt.Sprintf("AppID is not present in the request"),
		}}
	}
	for i := 0; i < len(request.Imp); i++ {
		err := preprocess(&request.Imp[i])
		// If the preprocessing failed, the server won't be able to bid on this Imp. Delete it, and note the error.
		if err != nil {
			errs = append(errs, err)
			request.Imp = append(request.Imp[:i], request.Imp[i+1:]...)
			i--
		}
	}

	// If all the requests were malformed, don't bother making a server call with no impressions.
	if len(request.Imp) == 0 {
		return nil, errs
	}

	if request.App != nil {
		appClone := *request.App
		if appClone.Publisher != nil {
			publisherCopy := *appClone.Publisher
			publisherCopy.ID = pubID
			appClone.Publisher = &publisherCopy
		} else {
			appClone.Publisher = &openrtb.Publisher{ID: pubID}
		}
		request.App = &appClone
	} else if request.Site != nil {
		siteClone := *request.Site
		if siteClone.Publisher != nil {
			publisherCopy := *siteClone.Publisher
			publisherCopy.ID = pubID
			siteClone.Publisher = &publisherCopy
		} else {
			siteClone.Publisher = &openrtb.Publisher{ID: pubID}
		}
		request.Site = &siteClone
	}

	request.Ext = json.RawMessage(fmt.Sprintf("{\"platformid\": \"%s\"}", a.PlatformID))
	//TODO remove this code
	request.Test = 1
	thisURI := a.URI

	reqJSON, err := json.Marshal(request)
	if err != nil {
		errs = append(errs, err)
		return nil, errs
	}

	headers := http.Header{}
	headers.Add("Content-Type", "application/json;charset=utf-8")
	headers.Add("Accept", "application/json")

	return []*adapters.RequestData{{
		Method:  "POST",
		Uri:     thisURI,
		Body:    reqJSON,
		Headers: headers,
	}}, errs

}

func preprocess(imp *openrtb.Imp) error {
	var bidderExt adapters.ExtImpBidder
	if err := json.Unmarshal(imp.Ext, &bidderExt); err != nil {
		return err
	}

	var impExt openrtb_ext.ExtImpAudienceNetwork
	if err := json.Unmarshal(bidderExt.Bidder, &impExt); err != nil {
		return err
	}

	imp.TagID = impExt.PlacementID
	imp.Ext = nil
	// if instl = 1 sent in, pass size (0,0) to facebook
	if imp.Instl == 1 && imp.Banner != nil {
		imp.Banner.W = openrtb.Uint64Ptr(0)
		imp.Banner.H = openrtb.Uint64Ptr(0)

		// if instl = 0 and type is banner, do not send non supported size
	} /*else if imp.Instl == 0 && imp.Banner != nil && imp.Banner.H != nil {
		if !anSupportedHeight[*imp.Banner.H] {
			return &errortypes.BadInput{
				Message: fmt.Sprintf("AudienceNetwork do not support banner height other than 50 and 250"),
			}
		}
		imp.Banner.W = openrtb.Uint64Ptr(0)

	} else if imp.Instl == 0 && imp.Banner.Format != nil && len(imp.Banner.Format) != 0 {
		if !anSupportedHeight[imp.Banner.Format[0].H] {
			return &errortypes.BadInput{
				Message: fmt.Sprintf("AudienceNetwork do not support banner height other than 50 and 250"),
			}
		}
		imp.Banner.W = openrtb.Uint64Ptr(0)
		imp.Banner.H = openrtb.Uint64Ptr(imp.Banner.Format[0].H)
	}*/
	return nil
}

func (a *AudienceNetworkAdapter) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	if response.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if response.StatusCode == http.StatusBadRequest {
		return nil, []error{&errortypes.BadInput{
			Message: fmt.Sprintf("Unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode),
		}}
	}

	if response.StatusCode != http.StatusOK {
		return nil, []error{fmt.Errorf("Unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode)}
	}

	var bidResp openrtb.BidResponse
	if err := json.Unmarshal(response.Body, &bidResp); err != nil {
		return nil, []error{err}
	}

	bidResponse := adapters.NewBidderResponseWithBidsCapacity(5)

	for _, sb := range bidResp.SeatBid {
		for i := range sb.Bid {
			bidResponse.Bids = append(bidResponse.Bids, &adapters.TypedBid{
				Bid:     &sb.Bid[i],
				BidType: getMediaTypeForImp(sb.Bid[i].ImpID, internalRequest.Imp),
			})
		}
	}
	return bidResponse, nil
}

// getMediaTypeForImp figures out which media type this bid is for.
// If both banner and video exist, take banner as we do not want in-banner video.
func getMediaTypeForImp(impId string, imps []openrtb.Imp) openrtb_ext.BidType {
	mediaType := openrtb_ext.BidTypeBanner
	for _, imp := range imps {
		if imp.ID == impId {
			if imp.Banner == nil && imp.Video != nil {
				mediaType = openrtb_ext.BidTypeVideo
			}
			return mediaType
		}
	}
	return mediaType
}

func getAppIDFromRequest(request *openrtb.BidRequest) (string, error) {
	bytes, err := getBidderParam(request, "appId")
	if err != nil {
		return "", err
	}

	if bytes == nil {
		return "", nil
	}
	return string(bytes), nil
}

func getBidderParam(request *openrtb.BidRequest, key string) ([]byte, error) {
	var reqExt openrtb_ext.ExtRequest
	err := json.Unmarshal(request.Ext, &reqExt)
	if err != nil {
		err := fmt.Errorf("%s Error unmarshalling request.ext: %v", string(openrtb_ext.BidderFacebook), string(request.Ext))
		return nil, err
	}

	if reqExt.Prebid.BidderParams == nil {
		return nil, nil
	}

	bidderParams, ok := reqExt.Prebid.BidderParams.(map[string]interface{})
	if !ok {
		err := fmt.Errorf("%s Error retrieving request.ext.prebid.ext: %v", string(openrtb_ext.BidderFacebook), reqExt.Prebid.BidderParams)
		return nil, err
	}

	iface, ok := bidderParams[key]
	if !ok {
		return nil, nil
	}

	bytes, err := json.Marshal(iface)
	if err != nil {
		err := fmt.Errorf("%s Error retrieving '%s' from request.ext.prebid.ext: %v", string(openrtb_ext.BidderFacebook), key, bidderParams)
		return nil, err
	}

	return bytes, nil
}

func NewAudienceNetworkAdapter(config *adapters.HTTPAdapterConfig, endpoint string, platformID string) *AudienceNetworkAdapter {
	a := adapters.NewHTTPAdapter(config)
	return &AudienceNetworkAdapter{
		http: a,
		URI:  endpoint,
		//for AB test
		// nonSecureUri: "http://an.facebook.com/placementbid.ortb",
		PlatformID: platformID,
	}
}

func NewAudienceNetworkBidder(client *http.Client, uri string, platformID string) *AudienceNetworkAdapter {
	fmt.Println("uri:", uri)
	fmt.Println("platform:", platformID)
	a := &adapters.HTTPAdapter{Client: client}
	return &AudienceNetworkAdapter{
		http:       a,
		URI:        uri,
		PlatformID: platformID,
	}
}
