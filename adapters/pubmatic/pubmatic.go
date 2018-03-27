package pubmatic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/pbs"
	"golang.org/x/net/context/ctxhttp"
)

const MAX_IMPRESSIONS_PUBMATIC = 30

const pmuri = "http://hbopenbid.pubmatic.com/translator?source=prebid-server"

type PubmaticAdapter struct {
	http *adapters.HTTPAdapter
	URI  string
}

// used for cookies and such
func (a *PubmaticAdapter) Name() string {
	return "pubmatic"
}

func (a *PubmaticAdapter) SkipNoCookies() bool {
	return false
}

// Below is bidder specific parameters for pubmatic adaptor,
// PublisherId and adSlot are mandatory parameters, others are optional parameters
// PmZoneid, Kadfloor are bid specific parameters,
// other parameters Lat,Lon, Yob, Kadpageurl, Gender, Yob, WrapExt needs to sent once per bid  request
type pubmaticParams struct {
	PublisherId string          `json:"publisherId"`
	AdSlot      string          `json:"adSlot"`
	Lat         float64         `json:"lat"`
	Lon         float64         `json:"Lon"`
	Yob         int             `json:"yob"`
	Kadpageurl  string          `json:"kadpageurl"`
	Gender      string          `json:"gender"`
	Kadfloor    float64         `json:"kadfloor"`
	WrapExt     json.RawMessage `json:"wrapper"`
	PmZoneid    string          `json:"pmzoneid"`
}

func PrepareLogMessage(tID, pubId, adUnitId, bidID, details string, args ...interface{}) string {
	return fmt.Sprintf("[PUBMATIC] ReqID [%s] PubID [%s] AdUnit [%s] BidID [%s] %s \n",
		tID, pubId, adUnitId, bidID, details)
}

func (a *PubmaticAdapter) Call(ctx context.Context, req *pbs.PBSRequest, bidder *pbs.PBSBidder) (pbs.PBSBidSlice, error) {
	mediaTypes := []pbs.MediaType{pbs.MEDIA_TYPE_BANNER, pbs.MEDIA_TYPE_VIDEO}
	pbReq, err := adapters.MakeOpenRTBGeneric(req, bidder, a.Name(), mediaTypes, true)

	if err != nil {
		glog.Warningf("[PUBMATIC] Failed to make ortb request for request id [%s] \n", pbReq.ID)
		return nil, err
	}

	adSlotFlag := false
	pubId := ""
	wrapExt := ""
	lat := 0.0
	lon := 0.0
	yob := 0
	kadPageURL := ""
	gender := ""
	if len(bidder.AdUnits) > MAX_IMPRESSIONS_PUBMATIC {
		glog.Warningf("[PUBMATIC] First %d impressions will be considered from request tid %s\n",
			MAX_IMPRESSIONS_PUBMATIC, pbReq.ID)
	}

	for i, unit := range bidder.AdUnits {
		var params pubmaticParams
		err := json.Unmarshal(unit.Params, &params)
		if err != nil {
			glog.Warningf(PrepareLogMessage(pbReq.ID, params.PublisherId, unit.Code, unit.BidID,
				fmt.Sprintf("Ignored bid: invalid JSON  [%s] err [%s]", unit.Params, err.Error())))
			continue
		}

		if params.PublisherId == "" {
			glog.Warningf(PrepareLogMessage(pbReq.ID, params.PublisherId, unit.Code, unit.BidID,
				fmt.Sprintf("Ignored bid: Publisher Id missing")))
			continue
		}
		pubId = params.PublisherId
		if params.AdSlot == "" {
			glog.Warningf(PrepareLogMessage(pbReq.ID, params.PublisherId, unit.Code, unit.BidID,
				fmt.Sprintf("Ignored bid: adSlot missing")))
			continue
		}

		// Parse Wrapper Extension i.e. ProfileID and VersionID only once per request
		if wrapExt == "" && len(string(params.WrapExt)) != 0 {
			wrapExt = string(params.WrapExt)
		}

		if params.Lat != 0.0 && lat == 0.0 {
			lat = params.Lat
		}

		if params.Lon != 0.0 && lon == 0.0 {
			lon = params.Lon
		}

		if params.Yob != 0 && yob == 0 {
			yob = params.Yob
		}

		if len(params.Gender) != 0 && gender == "" {
			gender = params.Gender
		}

		if len(params.Kadpageurl) != 0 && kadPageURL == "" {
			kadPageURL = params.Kadpageurl
		}

		adSlotStr := strings.TrimSpace(params.AdSlot)
		adSlot := strings.Split(adSlotStr, "@")
		if len(adSlot) == 2 && adSlot[0] != "" && adSlot[1] != "" {
			// Fixes some segfaults. Since this is legacy code, I'm not looking into it too deeply
			if len(pbReq.Imp) <= i {
				break
			}
			if pbReq.Imp[i].Banner != nil {
				adSize := strings.Split(strings.ToLower(strings.TrimSpace(adSlot[1])), "x")
				if len(adSize) == 2 {
					width, err := strconv.Atoi(strings.TrimSpace(adSize[0]))
					if err != nil {
						glog.Warningf(PrepareLogMessage(pbReq.ID, params.PublisherId, unit.Code, unit.BidID,
							fmt.Sprintf("Ignored bid: invalid adSlot width [%s]", adSize[0])))
						continue
					}

					heightStr := strings.Split(strings.TrimSpace(adSize[1]), ":")
					height, err := strconv.Atoi(strings.TrimSpace(heightStr[0]))
					if err != nil {
						glog.Warningf(PrepareLogMessage(pbReq.ID, params.PublisherId, unit.Code, unit.BidID,
							fmt.Sprintf("Ignored bid: invalid adSlot height [%s]", heightStr[0])))
						continue
					}

					pbReq.Imp[i].TagID = strings.TrimSpace(adSlot[0])
					pbReq.Imp[i].Banner.H = openrtb.Uint64Ptr(uint64(height))
					pbReq.Imp[i].Banner.W = openrtb.Uint64Ptr(uint64(width))

					if params.Kadfloor != 0.0 {
						pbReq.Imp[i].BidFloor = params.Kadfloor
					}

					impExtJsonStr := ""
					if len(params.PmZoneid) != 0 {
						impExtJsonStr = fmt.Sprintf("\"pmZoneId\": \"%s\"", params.PmZoneid)
						impExtJsonStr = "{" + impExtJsonStr + "}"
						pbReq.Imp[i].Ext = openrtb.RawJSON([]byte(impExtJsonStr))
					}

					adSlotFlag = true
				} else {
					glog.Warningf(PrepareLogMessage(pbReq.ID, params.PublisherId, unit.Code, unit.BidID,
						fmt.Sprintf("Ignored bid: invalid adSize [%s]", adSize)))
					continue
				}
			}
		} else {
			glog.Warningf(PrepareLogMessage(pbReq.ID, params.PublisherId, unit.Code, unit.BidID,
				fmt.Sprintf("Ignored bid: invalid adSlot [%s]", params.AdSlot)))
			continue
		}

		if pbReq.Site != nil {
			siteCopy := *pbReq.Site
			siteCopy.Publisher = &openrtb.Publisher{ID: params.PublisherId, Domain: req.Domain}
			pbReq.Site = &siteCopy
			if kadPageURL != "" {
				pbReq.Site.Page = kadPageURL
			}
		}
		if pbReq.App != nil {
			appCopy := *pbReq.App
			appCopy.Publisher = &openrtb.Publisher{ID: params.PublisherId, Domain: req.Domain}
			pbReq.App = &appCopy
		}
	}

	if !(adSlotFlag) {
		return nil, errors.New("Incorrect adSlot / Publisher param")
	}

	if lat != 0.0 || lon != 0.0 {
		geo := new(openrtb.Geo)
		geo.Lat = lat
		geo.Lon = lon
		if pbReq.User == nil {
			pbReq.User = new(openrtb.User)
		}
		pbReq.User.Geo = geo
	}

	if gender != "" {
		pbReq.User.Gender = gender
	}

	if yob != 0 {
		pbReq.User.Yob = int64(yob)
	}

	if wrapExt != "" {
		rawExt := fmt.Sprintf("{\"wrapper\": %s}", wrapExt)
		pbReq.Ext = openrtb.RawJSON(rawExt)
	} else {
		glog.Warning("Wrapper Extension not present in request")
	}

	reqJSON, err := json.Marshal(pbReq)

	debug := &pbs.BidderDebug{
		RequestURI: a.URI,
	}

	if req.IsDebug {
		debug.RequestBody = string(reqJSON)
		bidder.Debug = append(bidder.Debug, debug)
	}

	userId, _, _ := req.Cookie.GetUID(a.Name())
	httpReq, err := http.NewRequest("POST", a.URI, bytes.NewBuffer(reqJSON))
	httpReq.Header.Add("Content-Type", "application/json;charset=utf-8")
	httpReq.Header.Add("Accept", "application/json")
	httpReq.AddCookie(&http.Cookie{
		Name:  "KADUSERCOOKIE",
		Value: userId,
	})

	pbResp, err := ctxhttp.Do(ctx, a.http.Client, httpReq)
	if err != nil {
		return nil, err
	}

	debug.StatusCode = pbResp.StatusCode

	if pbResp.StatusCode == 204 {
		return nil, nil
	}

	if pbResp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP status: %d", pbResp.StatusCode)
	}

	defer pbResp.Body.Close()
	body, err := ioutil.ReadAll(pbResp.Body)
	if err != nil {
		return nil, err
	}

	if req.IsDebug {
		debug.ResponseBody = string(body)
	}

	var bidResp openrtb.BidResponse
	err = json.Unmarshal(body, &bidResp)
	if err != nil {
		return nil, err
	}

	bids := make(pbs.PBSBidSlice, 0)

	numBids := 0
	for _, sb := range bidResp.SeatBid {
		for _, bid := range sb.Bid {
			numBids++

			bidID := bidder.LookupBidID(bid.ImpID)
			if bidID == "" {
				return nil, errors.New(fmt.Sprintf("Unknown ad unit code '%s'", bid.ImpID))
			}

			pbid := pbs.PBSBid{
				BidID:       bidID,
				AdUnitCode:  bid.ImpID,
				BidderCode:  bidder.BidderCode,
				Price:       bid.Price,
				Adm:         bid.AdM,
				Creative_id: bid.CrID,
				Width:       bid.W,
				Height:      bid.H,
				DealId:      bid.DealID,
			}

			mediaType := getMediaTypeForImp(bid.ImpID, pbReq.Imp)
			pbid.CreativeMediaType = string(mediaType)
			bids = append(bids, &pbid)
			if glog.V(2) {
				glog.Infof("[PUBMATIC] Returned Bid for PubID [%s] AdUnit [%s] BidID [%s] Size [%dx%d] Price [%f] \n",
					pubId, pbid.AdUnitCode, pbid.BidID, pbid.Width, pbid.Height, pbid.Price)
			}
		}
	}

	return bids, nil
}

func (a *PubmaticAdapter) MakeRequests(request *openrtb.BidRequest) ([]*adapters.RequestData, []error) {

	errs := make([]error, 0, len(request.Imp))

	var err error
	wrapExt := ""
	pubID := ""
	lat := 0.0
	lon := 0.0
	yob := 0
	kadpageUrl := ""
	gender := ""

	for i := 0; i < len(request.Imp); i++ {
		err = parseImpressionObject(&request.Imp[i], &wrapExt, &pubID, &lat, &lon, &yob, &kadpageUrl, &gender)

		// If the parsing is failed, remove imp and add the error.
		if err != nil {
			errs = append(errs, err)
			request.Imp = append(request.Imp[:i], request.Imp[i+1:]...)
			i--
		}
	}

	if wrapExt != "" {
		rawExt := fmt.Sprintf("{\"wrapper\": %s}", wrapExt)
		request.Ext = openrtb.RawJSON(rawExt)
	} else {
		glog.Warning("Wrapper Extension not present in request")
	}

	if request.Site != nil {
		if request.Site.Publisher != nil {
			request.Site.Publisher.ID = pubID
		} else {
			request.Site.Publisher = &openrtb.Publisher{ID: pubID}
		}
		if kadpageUrl != "" {
			request.Site.Page = kadpageUrl
		}

	} else {
		if request.App.Publisher != nil {
			request.App.Publisher.ID = pubID
		} else {
			request.App.Publisher = &openrtb.Publisher{ID: pubID}
		}
	}

	if lat != 0.0 || lon != 0.0 {
		geo := new(openrtb.Geo)
		geo.Lat = lat
		geo.Lon = lon
		if request.User == nil {
			request.User = new(openrtb.User)
		}
		request.User.Geo = geo
	}

	if gender != "" {
		request.User.Gender = gender
	}

	if yob != 0 {
		request.User.Yob = int64(yob)
	}

	thisUri := pmuri

	// If all the requests are invalid, Call to adaptor is skipped
	if len(request.Imp) == 0 {
		return nil, errs
	}

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
		Uri:     thisUri,
		Body:    reqJSON,
		Headers: headers,
	}}, errs
}

// parseImpressionObject parase  the imp to get it ready to send to pubmatic
func parseImpressionObject(imp *openrtb.Imp, wrapExt *string, pubID *string,
	lat *float64, lon *float64, yob *int, kadPageURL *string, gender *string) error {
	// PubMatic supports native, banner and video impressions.
	if imp.Audio != nil {
		return fmt.Errorf("PubMatic doesn't support audio. Ignoring ImpID = %s", imp.ID)
	}

	var bidderExt adapters.ExtImpBidder
	if err := json.Unmarshal(imp.Ext, &bidderExt); err != nil {
		return err
	}

	var pubmaticExt openrtb_ext.ExtImpPubmatic
	if err := json.Unmarshal(bidderExt.Bidder, &pubmaticExt); err != nil {
		return err
	}

	if pubmaticExt.AdSlot == "" || pubmaticExt.PubID == "" {
		return errors.New("No AdSlot  or PubID  provided")
	}

	if *pubID == "" {
		*pubID = pubmaticExt.PubID
	}

	// Parse Wrapper Extension i.e. ProfileID and VersionID only once per request
	if *wrapExt == "" && len(string(pubmaticExt.WrapExt)) != 0 {
		*wrapExt = string(pubmaticExt.WrapExt)
	}

	if *lat == 0.0 && pubmaticExt.Lat != 0.0 {
		*lat = pubmaticExt.Lat
	}

	if *lon == 0.0 && pubmaticExt.Lon != 0.0 {
		*lon = pubmaticExt.Lon
	}

	if *yob == 0 && pubmaticExt.Yob != 0 {
		*yob = pubmaticExt.Yob
	}

	if *kadPageURL == "" && len(pubmaticExt.Kadpageurl) != 0 {
		*kadPageURL = pubmaticExt.Kadpageurl
	}

	if *gender == "" && len(pubmaticExt.Gender) != 0 {
		*gender = pubmaticExt.Gender
	}

	if pubmaticExt.Kadfloor != 0.0 {
		imp.BidFloor = pubmaticExt.Kadfloor
	}

	adSlotStr := strings.TrimSpace(pubmaticExt.AdSlot)
	if imp.Banner != nil {

		adSlot := strings.Split(adSlotStr, "@")
		if len(adSlot) == 2 && adSlot[0] != "" && adSlot[1] != "" {
			imp.TagID = strings.TrimSpace(adSlot[0])

			adSize := strings.Split(strings.ToLower(strings.TrimSpace(adSlot[1])), "x")
			if len(adSize) == 2 {
				width, err := strconv.Atoi(strings.TrimSpace(adSize[0]))
				if err != nil {
					return errors.New("Invalid Width Provided ")
				}

				heightStr := strings.Split(strings.TrimSpace(adSize[1]), ":")
				height, err := strconv.Atoi(strings.TrimSpace(heightStr[0]))
				if err != nil {
					return errors.New("Invalid Height Provided ")
				}

				imp.Banner.W = openrtb.Uint64Ptr(uint64(height))
				imp.Banner.W = openrtb.Uint64Ptr(uint64(width))

			} else {
				return errors.New("Invalid adSizes Provided ")
			}
		} else {
			return errors.New("Invalid adSlot  Provided ")
		}
	} else {
		imp.TagID = strings.TrimSpace(adSlotStr)
	}

	keyValStr := makeImpressionExt(pubmaticExt.Keywords)
	if len(keyValStr) != 0 {
		imp.Ext = openrtb.RawJSON([]byte(keyValStr))
	} else {
		imp.Ext = nil
	}

	return nil

}

func makeImpressionExt(keywords []*openrtb_ext.ImpExtPubmaticKeyVal) string {

	kvStr := ""
	for _, kv := range keywords {
		eachkvStr := ""
		if len(kv.Values) == 1 {
			eachkvStr = fmt.Sprintf("\"%s\": \"%s\"", kv.Key, kv.Values[0])
		} else {

			for i, val := range kv.Values {
				if i == 0 {
					eachkvStr = fmt.Sprintf("\"%s\": \"%s", kv.Key, val)
				} else {
					eachkvStr = eachkvStr + "," + val
				}
			}
			eachkvStr = eachkvStr + "\""
		}
		if len(kvStr) == 0 {
			kvStr = eachkvStr
		} else {
			kvStr = kvStr + "," + eachkvStr
		}
	}
	if len(kvStr) != 0 {
		kvStr = "{" + kvStr + "}"
	}
	return kvStr
}

func (a *PubmaticAdapter) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) ([]*adapters.TypedBid, []error) {
	if response.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if response.StatusCode != http.StatusOK {
		return nil, []error{fmt.Errorf("Unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode)}
	}

	var bidResp openrtb.BidResponse
	if err := json.Unmarshal(response.Body, &bidResp); err != nil {
		return nil, []error{err}
	}

	bids := make([]*adapters.TypedBid, 0, 5)

	for _, sb := range bidResp.SeatBid {
		for i, bid := range sb.Bid {

			bids = append(bids, &adapters.TypedBid{
				Bid:     &sb.Bid[i],
				BidType: getMediaTypeForImp(bid.ImpID, internalRequest.Imp),
			})
		}
	}
	return bids, nil
}

// getMediaTypeForImp figures out which media type this bid is for.
func getMediaTypeForImp(impId string, imps []openrtb.Imp) openrtb_ext.BidType {
	mediaType := openrtb_ext.BidTypeBanner
	for _, imp := range imps {
		if imp.ID == impId {
			if imp.Video != nil {
				mediaType = openrtb_ext.BidTypeVideo
			} else if imp.Audio != nil {
				mediaType = openrtb_ext.BidTypeAudio
			} else if imp.Native != nil {
				mediaType = openrtb_ext.BidTypeNative
			}
			return mediaType
		}
	}
	return mediaType
}
func NewPubmaticAdapter(config *adapters.HTTPAdapterConfig, uri string) *PubmaticAdapter {
	a := adapters.NewHTTPAdapter(config)

	return &PubmaticAdapter{
		http: a,
		URI:  uri,
	}
}
func NewPubmaticBidder(client *http.Client) *PubmaticAdapter {
	a := &adapters.HTTPAdapter{Client: client}
	return &PubmaticAdapter{
		http: a,
		URI:  pmuri,
	}
}
