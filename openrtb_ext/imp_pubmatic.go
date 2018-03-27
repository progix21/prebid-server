package openrtb_ext

import "encoding/json"

// ExtImpPubmatic defines the contract for bidrequest.imp[i].ext.pubmatic
type ExtImpPubmatic struct {
	AdSlot     string                  `json:"adSlot"`
	PubID      string                  `json:"publisherId"`
	Lat        float64                 `json:"lat"`
	Lon        float64                 `json:"lon"`
	Yob        int                     `json:"yob"`
	Kadpageurl string                  `json:"kadpageurl"`
	Gender     string                  `json:"gender"`
	Kadfloor   float64                 `json:"kadfloor"`
	WrapExt    json.RawMessage         `json:"wrapper"`
	Keywords   []*ImpExtPubmaticKeyVal `json:"keywords"`
}

// ImpExtPubmaticKeyVal defines the contract for bidrequest.imp[i].ext.pubmatic.keywords[i]
type ImpExtPubmaticKeyVal struct {
	Key    string   `json:"key,omitempty"`
	Values []string `json:"vals,omitempty"`
}
