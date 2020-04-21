package openrtb2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/PubMatic-OpenWrap/prebid-server/analytics"
	"github.com/PubMatic-OpenWrap/prebid-server/config"
	"github.com/PubMatic-OpenWrap/prebid-server/endpoints/openrtb2/ctv"
	"github.com/PubMatic-OpenWrap/prebid-server/exchange"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
	"github.com/PubMatic-OpenWrap/prebid-server/pbsmetrics"
	"github.com/PubMatic-OpenWrap/prebid-server/stored_requests"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
)

//AdPodConfig configuration for creating ads in adpod
type ImpAdPodConfig struct {
	ImpID          string
	SequenceNumber int8
	MinDuration    int64
	MaxDuration    int64
}

//ImpAdPodConfigMap map of imp ad pod configurations
type ImpAdPodConfigMap map[string][]ImpAdPodConfig

//AdPodBids map of adpods impression bids
type AdPodBids []ctv.ImpBids

//AdPodBidsMap map of impression with adpod details
type AdPodBidsMap map[string]AdPodBids

//CTV Specific Endpoint
type ctvEndpointDeps struct {
	endpointDeps
	impAdpodConfigs ImpAdPodConfigMap
}

func NewCTVEndpoint(
	ex exchange.Exchange,
	validator openrtb_ext.BidderParamValidator,
	requestsById stored_requests.Fetcher,
	videoFetcher stored_requests.Fetcher,
	categories stored_requests.CategoryFetcher,
	cfg *config.Configuration,
	met pbsmetrics.MetricsEngine,
	pbsAnalytics analytics.PBSAnalyticsModule,
	disabledBidders map[string]string,
	defReqJSON []byte,
	bidderMap map[string]openrtb_ext.BidderName) (httprouter.Handle, error) {

	if ex == nil || validator == nil || requestsById == nil || cfg == nil || met == nil {
		return nil, errors.New("NewCTVEndpoint requires non-nil arguments.")
	}
	defRequest := defReqJSON != nil && len(defReqJSON) > 0

	return httprouter.Handle((&ctvEndpointDeps{
		endpointDeps: endpointDeps{
			ex,
			validator,
			requestsById,
			videoFetcher,
			categories,
			cfg,
			met,
			pbsAnalytics,
			disabledBidders,
			defRequest,
			defReqJSON,
			bidderMap,
		},
	}).CTVAuctionEndpoint), nil
}

func (deps *ctvEndpointDeps) CTVAuctionEndpoint(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	ao := analytics.AuctionObject{
		Status: http.StatusOK,
		Errors: make([]error, 0),
	}

	// Prebid Server interprets request.tmax to be the maximum amount of time that a caller is willing
	// to wait for bids. However, tmax may be defined in the Stored Request data.
	//
	// If so, then the trip to the backend might use a significant amount of this time.
	// We can respect timeouts more accurately if we note the *real* start time, and use it
	// to compute the auction timeout.
	start := time.Now()
	//Prebid Stats
	labels := pbsmetrics.Labels{
		Source:        pbsmetrics.DemandUnknown,
		RType:         pbsmetrics.ReqTypeVideo,
		PubID:         pbsmetrics.PublisherUnknown,
		Browser:       getBrowserName(r),
		CookieFlag:    pbsmetrics.CookieFlagUnknown,
		RequestStatus: pbsmetrics.RequestStatusOK,
	}
	defer func() {
		deps.metricsEngine.RecordRequest(labels)
		deps.metricsEngine.RecordRequestTime(labels, time.Since(start))
		deps.analytics.LogAuctionObject(&ao)
	}()

	//Parse ORTB Request and do Standard Validation
	req, errL := deps.parseRequest(r)
	if fatalError(errL) && writeError(errL, w, &labels) {
		return
	}

	//Set Default Values
	deps.setDefaultValues(req)

	//Validate CTV BidRequest
	if err := deps.validateBidRequest(req); err != nil {
		errL = append(errL, err)
		writeError(errL, w, &labels)
		return
	}

	//Create New BidRequest
	ctvReq := deps.createBidRequest(req)

	ctx := context.Background()

	//Setting Timeout for Request
	timeout := deps.cfg.AuctionTimeouts.LimitAuctionTimeout(time.Duration(ctvReq.TMax) * time.Millisecond)
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, start.Add(timeout))
		defer cancel()
	}

	//Parsing Cookies and Set Stats
	usersyncs := usersync.ParsePBSCookieFromRequest(r, &(deps.cfg.HostCookie))
	if ctvReq.App != nil {
		labels.Source = pbsmetrics.DemandApp
		labels.RType = pbsmetrics.ReqTypeVideo
		labels.PubID = effectivePubID(ctvReq.App.Publisher)
	} else { //ctvReq.Site != nil
		labels.Source = pbsmetrics.DemandWeb
		if usersyncs.LiveSyncCount() == 0 {
			labels.CookieFlag = pbsmetrics.CookieFlagNo
		} else {
			labels.CookieFlag = pbsmetrics.CookieFlagYes
		}
		labels.PubID = effectivePubID(ctvReq.Site.Publisher)
	}

	//Validate Accounts
	if acctIdErr := validateAccount(deps.cfg, labels.PubID); acctIdErr != nil {
		errL = append(errL, acctIdErr)
		writeError(errL, w, &labels)
		return
	}

	//Hold OpenRTB Standard Auction
	response, err := deps.ex.HoldAuction(ctx, ctvReq, usersyncs, labels, &deps.categories)
	ao.Request = ctvReq
	ao.Response = response
	if err != nil {
		labels.RequestStatus = pbsmetrics.RequestStatusErr
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Critical error while running the auction: %v", err)
		glog.Errorf("/openrtb2/video Critical error: %v", err)
		ao.Status = http.StatusInternalServerError
		ao.Errors = append(ao.Errors, err)
		return
	}

	//Validate Bid Response
	if err := deps.validateBidResponse(ctvReq, response); err != nil {
		errL = append(errL, err)
		writeError(errL, w, &labels)
		return
	}

	//Create Impression Bids
	adPodBidsMap := deps.createImpBids(response, deps.impAdpodConfigs)

	//Do AdPod Exclusions
	bids := deps.doAdPodExclusions(ctvReq, adPodBidsMap)

	//deps.doInterAdPodExclusions() {}
	//deps.doAuction() {}

	//Create Bid Response
	ctvResp := deps.createBidResponse(response, bids)

	// Response Generation
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	// Fixes #328
	w.Header().Set("Content-Type", "application/json")

	// If an error happens when encoding the response, there isn't much we can do.
	// If we've sent _any_ bytes, then Go would have sent the 200 status code first.
	// That status code can't be un-sent... so the best we can do is log the error.
	if err := enc.Encode(ctvResp); err != nil {
		labels.RequestStatus = pbsmetrics.RequestStatusNetworkErr
		ao.Errors = append(ao.Errors, fmt.Errorf("/openrtb2/video Failed to send response: %v", err))
	}
}

//setDefaultValues will set adpod and other default values
func (deps *ctvEndpointDeps) setDefaultValues(req *openrtb.BidRequest) {}

//validateBidRequest will validate AdPod specific mandatory Parameters and returns error
func (deps *ctvEndpointDeps) validateBidRequest(req *openrtb.BidRequest) error {
	return nil
}

//createBidRequest will return new bid request with all things copy from bid request except impression objects
func (deps *ctvEndpointDeps) createBidRequest(req *openrtb.BidRequest) *openrtb.BidRequest {
	ctvRequest := *req

	//get configurations for all impressions
	deps.impAdpodConfigs = deps.getAllAdPodImpsConfigs(req)

	//createImpressions
	ctvRequest.Imp = deps.createImpressions(req, deps.impAdpodConfigs)

	//TODO: remove adpod extension if not required to send further
	return &ctvRequest
}

//createImpressions will create multiple impressions based on adpod configurations
func (deps *ctvEndpointDeps) createImpressions(req *openrtb.BidRequest, impAdpodConfigs ImpAdPodConfigMap) []openrtb.Imp {
	impCount := 0
	for _, config := range impAdpodConfigs {
		impCount = impCount + len(config)
	}

	count := 0
	imps := make([]openrtb.Imp, impCount)
	for _, imp := range req.Imp {
		adPodConfig := impAdpodConfigs[imp.ID]
		for _, config := range adPodConfig {
			imps[count] = *(deps.newImpression(&imp, &config))
			count++
		}
	}

	return imps[:]
}

//getAllAdPodImpsConfigs will return all impression adpod configurations
func (deps *ctvEndpointDeps) getAllAdPodImpsConfigs(req *openrtb.BidRequest) ImpAdPodConfigMap {
	result := make(ImpAdPodConfigMap, len(req.Imp))

	for _, imp := range req.Imp {
		ext := openrtb_ext.VideoExtension{}
		json.Unmarshal(imp.Ext, &ext)
		result[imp.ID] = deps.getAdPodImpsConfigs(&imp, ext.AdPod)
	}

	return result
}

//getAdPodImpsConfigs will return number of impressions configurations within adpod
func (deps *ctvEndpointDeps) getAdPodImpsConfigs(imp *openrtb.Imp, adpod *openrtb_ext.VideoAdPod) []ImpAdPodConfig {
	config := []ImpAdPodConfig{
		{
			ImpID:          imp.ID,
			MinDuration:    int64(*adpod.MinDuration),
			MaxDuration:    int64(*adpod.MaxDuration),
			SequenceNumber: 1, /* Must be starting with 1 */
		},
	}
	return config[:]
}

//newImpression will clone existing impression object and create video object with ImpAdPodConfig.
func (deps *ctvEndpointDeps) newImpression(imp *openrtb.Imp, config *ImpAdPodConfig) *openrtb.Imp {
	video := *imp.Video
	video.MinDuration = config.MinDuration
	video.MaxDuration = config.MaxDuration
	video.Sequence = config.SequenceNumber
	video.MaxExtended = 0
	//TODO: remove video adpod extension if not required

	newImp := *imp
	newImp.ID = fmt.Sprintf("%s:%d", imp.ID, config.SequenceNumber)
	//newImp.BidFloor = 0
	newImp.Video = &video
	return &newImp
}

//validateBidResponse
func (deps *ctvEndpointDeps) validateBidResponse(req *openrtb.BidRequest, resp *openrtb.BidResponse) error {
	//checking bids for category and advertiser domain values.
	//remove bids without bid.id
	//remove bids with price=0
	return nil
}

//createImpBids will sort Bid Responses of each impressions
func (deps *ctvEndpointDeps) createImpBids(resp *openrtb.BidResponse, impAdpodConfigs ImpAdPodConfigMap) AdPodBidsMap {
	result := make(AdPodBidsMap)

	for _, seat := range resp.SeatBid {
		for _, bid := range seat.Bid {
			originalImpID, sequence := decodeImpressionID(bid.ImpID)

			if _, ok := result[originalImpID]; !ok {
				//create new map for ad pods
				config := impAdpodConfigs[originalImpID]
				result[originalImpID] = make(AdPodBids, len(config))
			}

			result[originalImpID][sequence-1] = append(result[originalImpID][sequence-1], &ctv.ImpBid{
				SeatName: seat.Seat,
				Bid:      &bid,
			})
		}
	}

	//Sort Bids by Price
	for originalImpID, imps := range result {
		for sequence, bids := range imps {
			//sort bids
			sort.Slice(bids[:], func(i, j int) bool { return bids[i].Price > bids[j].Price })
			result[originalImpID][sequence] = bids
		}
	}
	return result
}

//doAdPodExclusions
func (deps *ctvEndpointDeps) doAdPodExclusions(req *openrtb.BidRequest, bidsMap AdPodBidsMap) AdPodBids {
	result := AdPodBids{}
	for _, imp := range req.Imp {
		bids, ok := bidsMap[imp.ID]
		if ok {
			adpodGenerator := ctv.NewCombinations(bids, func(x *ctv.ImpBid, y *ctv.ImpBid) bool { return true })
			adpod := adpodGenerator.GetAdPod()
			if adpod != nil {
				result = append(result, adpod)
			}
		}
	}
	return result
}

//doInterAdPodExclusions
//func (deps *ctvEndpointDeps) doInterAdPodExclusions() {}

//doAuction
//func (deps *ctvEndpointDeps) doAuction() {}

//createBidResponse
func (deps *ctvEndpointDeps) createBidResponse(resp *openrtb.BidResponse, adpods []ctv.ImpBids) *openrtb.BidResponse {
	bidResp := &openrtb.BidResponse{}
	bidResp.Ext = resp.Ext
	for _, adpod := range adpods {
		if len(adpod) == 0 {
			continue
		}
		bid := deps.getAdPodBid(adpod)
		if bid != nil {
			found := false
			for _, seat := range bidResp.SeatBid {
				if seat.Seat == adpod[0].SeatName {
					seat.Bid = append(seat.Bid, *bid)
					found = true
					break
				}
			}
			if found == false {
				bidResp.SeatBid = append(bidResp.SeatBid, openrtb.SeatBid{
					Seat: adpod[0].SeatName,
					Bid: []openrtb.Bid{
						*bid,
					},
				})
			}
		}
	}
	return bidResp
}

//getAdPodBid
func (deps *ctvEndpointDeps) getAdPodBid(adpod ctv.ImpBids) *openrtb.Bid {
	bid := openrtb.Bid{}
	bid = *adpod[0].Bid
	//TODO: Write single for loop to get all details
	bid.AdM = *deps.getAdPodBidCreative(adpod)
	bid.Price = deps.getAdPodBidPrice(adpod)
	bid.ADomain = deps.getAdPodBidAdvertiserDomain(adpod)
	bid.Cat = deps.getAdPodBidCategories(adpod)
	bid.Ext = deps.getAdPodBidExtension(adpod)
	return &bid
}

//getAdPodBidCreative get commulative adpod bid details
func (deps *ctvEndpointDeps) getAdPodBidCreative(adpod ctv.ImpBids) *string {
	return &adpod[0].AdM
}

//getAdPodBidPrice get commulative adpod bid details
func (deps *ctvEndpointDeps) getAdPodBidPrice(adpod ctv.ImpBids) float64 {
	return adpod[0].Price
}

//getAdPodBidAdvertiserDomain get commulative adpod bid details
func (deps *ctvEndpointDeps) getAdPodBidAdvertiserDomain(adpod ctv.ImpBids) []string {
	return adpod[0].ADomain[:]
}

//getAdPodBidCategories get commulative adpod bid details
func (deps *ctvEndpointDeps) getAdPodBidCategories(adpod ctv.ImpBids) []string {
	return adpod[0].Cat[:]
}

//getAdPodBidExtension get commulative adpod bid details
func (deps *ctvEndpointDeps) getAdPodBidExtension(adpod ctv.ImpBids) json.RawMessage {
	return adpod[0].Ext
}

func decodeImpressionID(id string) (string, int) {
	values := strings.Split(id, ":")
	if len(values) == 1 {
		return values[0], 1
	}
	sequence, err := strconv.Atoi(values[1])
	if err != nil {
		sequence = 1
	}
	return values[0], sequence
}
