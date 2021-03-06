package endpoints

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PubMatic-OpenWrap/prebid-server/analytics"
	"github.com/PubMatic-OpenWrap/prebid-server/config"
	"github.com/PubMatic-OpenWrap/prebid-server/gdpr"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
	"github.com/PubMatic-OpenWrap/prebid-server/pbsmetrics"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
	"github.com/julienschmidt/httprouter"
)

const (
	chromeStr       = "Chrome/"
	chromeiOSStr    = "CriOS/"
	chromeMinVer    = 67
	chromeStrLen    = len(chromeStr)
	chromeiOSStrLen = len(chromeiOSStr)
)

func NewSetUIDEndpoint(cfg config.HostCookie, syncers map[openrtb_ext.BidderName]usersync.Usersyncer, perms gdpr.Permissions, pbsanalytics analytics.PBSAnalyticsModule, metrics pbsmetrics.MetricsEngine) httprouter.Handle {
	cookieTTL := time.Duration(cfg.TTL) * 24 * time.Hour

	validFamilyNameMap := make(map[string]struct{})
	for _, s := range syncers {
		validFamilyNameMap[s.FamilyName()] = struct{}{}
	}

	return httprouter.Handle(func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

		so := analytics.SetUIDObject{
			Status: http.StatusOK,
			Errors: make([]error, 0),
		}

		defer pbsanalytics.LogSetUIDObject(&so)

		pc := usersync.ParsePBSCookieFromRequest(r, &cfg)
		if !pc.AllowSyncs() {
			w.WriteHeader(http.StatusUnauthorized)
			metrics.RecordUserIDSet(pbsmetrics.UserLabels{
				Action: pbsmetrics.RequestActionOptOut,
			})
			so.Status = http.StatusUnauthorized
			return
		}

		query := r.URL.Query()

		familyName, err := getFamilyName(query, validFamilyNameMap)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			metrics.RecordUserIDSet(pbsmetrics.UserLabels{
				Action: pbsmetrics.RequestActionErr,
			})
			so.Status = http.StatusBadRequest
			return
		}
		so.Bidder = familyName

		if shouldReturn, status, body := preventSyncsGDPR(query.Get("gdpr"), query.Get("gdpr_consent"), perms); shouldReturn {
			w.WriteHeader(status)
			w.Write([]byte(body))
			metrics.RecordUserIDSet(pbsmetrics.UserLabels{
				Action: pbsmetrics.RequestActionGDPR,
				Bidder: openrtb_ext.BidderName(familyName),
			})
			so.Status = status
			return
		}

		uid := query.Get("uid")
		so.UID = uid

		if uid == "" {
			pc.Unsync(familyName)
		} else {
			err = pc.TrySync(familyName, uid)
		}

		if err == nil {
			labels := pbsmetrics.UserLabels{
				Action: pbsmetrics.RequestActionSet,
				Bidder: openrtb_ext.BidderName(familyName),
			}
			metrics.RecordUserIDSet(labels)
			so.Success = true
		}

		setSiteCookie := siteCookieCheck(r.UserAgent())

		secParam := r.URL.Query().Get("sec")
		pc.SetCookieOnResponse(w, setSiteCookie, secParam, &cfg, cookieTTL)
	})
}

func getFamilyName(query url.Values, validFamilyNameMap map[string]struct{}) (string, error) {
	// The family name is bound to the 'bidder' query param. In most cases, these values are the same.
	familyName := query.Get("bidder")

	if familyName == "" {
		return "", errors.New(`"bidder" query param is required`)
	}

	if _, ok := validFamilyNameMap[familyName]; !ok {
		return "", errors.New("The bidder name provided is not supported by Prebid Server")
	}

	return familyName, nil
}

// siteCookieCheck scans the input User Agent string to check if browser is Chrome and browser version is greater than the minimum version for adding the SameSite cookie attribute
func siteCookieCheck(ua string) bool {
	result := false

	index := strings.Index(ua, chromeStr)
	criOSIndex := strings.Index(ua, chromeiOSStr)
	if index != -1 {
		result = checkChromeBrowserVersion(ua, index, chromeStrLen)
	} else if criOSIndex != -1 {
		result = checkChromeBrowserVersion(ua, criOSIndex, chromeiOSStrLen)
	}
	return result
}

func checkChromeBrowserVersion(ua string, index int, chromeStrLength int) bool {
	result := false
	vIndex := index + chromeStrLength
	dotIndex := strings.Index(ua[vIndex:], ".")
	if dotIndex == -1 {
		dotIndex = len(ua[vIndex:])
	}
	version, _ := strconv.Atoi(ua[vIndex : vIndex+dotIndex])
	if version >= chromeMinVer {
		result = true
	}
	return result
}

func preventSyncsGDPR(gdprEnabled string, gdprConsent string, perms gdpr.Permissions) (bool, int, string) {
	switch gdprEnabled {
	case "0":
		return false, 0, ""
	case "1":
		if gdprConsent == "" {
			return true, http.StatusBadRequest, "gdpr_consent is required when gdpr=1"
		}
		fallthrough
	case "":
		if allowed, err := perms.HostCookiesAllowed(context.Background(), gdprConsent); err != nil {
			if _, ok := err.(*gdpr.ErrorMalformedConsent); ok {
				return true, http.StatusBadRequest, "gdpr_consent was invalid. " + err.Error()
			} else {
				// We can't really distinguish between requests that are for a new version of the global vendor list, and
				// ones which are simply malformed (version number is much too large).
				// Since we try to fetch new versions as requests come in for them, PBS *should* self-correct
				// rather quickly, meaning that most of these will be malformed strings.
				return true, http.StatusBadRequest, "No global vendor list was available to interpret this consent string. If this is a new, valid version, it should become available soon."
			}
		} else if !allowed {
			return true, http.StatusOK, "The gdpr_consent string prevents cookies from being saved"
		} else {
			return false, 0, ""
		}
	default:
		return true, http.StatusBadRequest, "the gdpr query param must be either 0 or 1. You gave " + gdprEnabled
	}
}
